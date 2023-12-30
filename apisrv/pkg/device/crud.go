package device

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	bl "github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

var (
	ErrCopy = errors.New("copy device failed")
)

// patch name
// patch attributes
// remove attributes
// add attributes
// empty key attr
// empty value attr vs. nil
// patch both
// patch none
// patch wrong org
// patch non-existent
// patch dup name
func Patch(ctx context.Context, qq pgxscan.Querier, id int64, orgId int64, name *string, baseline *baseline.Values, policy *policy.Values, now time.Time, actor string) error {
	var changes []string
	var ids []int64
	var blinedoc database.Document
	var poldoc database.Document
	var err error

	if name == nil && baseline == nil && policy == nil {
		return nil
	}
	if name != nil {
		changes = append(changes, "rename")
	}
	if baseline != nil {
		blinedoc, err = bl.ToRow(baseline)
		if err != nil {
			return err
		}
	} else {
		blinedoc, _ = database.NewDocument(nil)
	}
	if policy != nil {
		poldoc, err = policy.Serialize()
		if err != nil {
			return err
		}
	} else {
		poldoc, _ = database.NewDocument(nil)
	}

	// XXX simplify this query, we don't need the orgs CTE anymore
	sql := `
    --
    -- Patch-Device
    --
    with orgs as (
      select orgs.id
      from v2.organizations orgs
      where orgs.id = $1
      limit 1
      
    ), changes as (
      insert into v2.changes
        (actor, type, device_id, comment, timestamp, organization_id)
      select
        $6::text,
        ty::v2.changes_type,
        $2,
        null,
        $8::timestamptz,
        orgs.id
      from unnest($7::text[]) ty, orgs
      returning id
      
    )
    update v2.devices dev
    set (name, baseline, policy) = (
      coalesce($3::text, dev.name),
      coalesce($4::jsonb, dev.baseline),
      coalesce($5::jsonb, dev.policy)
    )
    from orgs
    where dev.id = $2
      and dev.organization_id = orgs.id
    returning dev.id
  `

	err = pgxscan.Select(ctx, qq, &ids, sql,
		orgId,    // $1
		id,       // $2
		name,     // $3
		blinedoc, // $4
		poldoc,   // $5
		actor,    // $6
		changes,  // $7
		now)      // $8
	if err != nil {
		return database.Error(err)
	}
	if len(ids) == 0 {
		return database.ErrNotFound
	}

	return nil
}

func Retire(ctx context.Context, tx pgx.Tx, id int64, orgId int64, retire bool, comment *string, now time.Time, actor string) error {
	var change string
	if retire {
		change = "retire"
	} else {
		change = "resurrect"
	}

	// XXX simplify query by removing orgs CTE
	ct, err := tx.Exec(ctx, `
		--
		-- Retire-Device
		--
		WITH orgs AS (
			SELECT v2.organizations.id
			FROM v2.organizations
			WHERE v2.organizations.id = $1

		), updated AS (
			UPDATE v2.devices
			SET retired = $6, replaced_by = (CASE WHEN ($6) THEN v2.devices.replaced_by ELSE NULL END)
			FROM orgs
			WHERE v2.devices.id = $2
				AND v2.devices.organization_id = orgs.id
			RETURNING v2.devices.*

		), change AS (
			INSERT INTO v2.changes
				(actor, type, device_id, comment, timestamp, organization_id)
			SELECT
				$3, $7::v2.changes_type, updated.id, $5, $4, updated.organization_id
			FROM updated
			RETURNING *

		)
		SELECT
			updated.id AS "device",
			change.id  AS "change"
		FROM updated
		INNER JOIN change ON change.device_id = updated.id
	`, orgId, id, actor, now, comment, retire, change)
	if err != nil {
		return database.Error(err)
	}
	if ct.RowsAffected() == 0 {
		return database.ErrNotFound
	}

	return nil
}
