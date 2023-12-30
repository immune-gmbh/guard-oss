package evidence

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"go.opentelemetry.io/otel/attribute"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type Row struct {
	// meta
	Id         string    `db:"id"`
	ReceivedAt time.Time `db:"received_at"`
	SignedBy   api.Name  `db:"signed_by"`
	DeviceId   *int64    `db:"device_id"`

	// evidence
	Values            database.Document `db:"values"`
	Baseline          database.Document `db:"baseline"`
	Policy            database.Document `db:"policy"`
	ImageReference    *string           `db:"image_ref"`
	BinarlyReference  *string           `db:"binarly_ref"`
	IntelTSCReference *string           `db:"inteltsc_ref"`
}

// Persist accepts a DevAikRow parameter that must contain a matching aik fpr and device id to ensure consistency; device id may be 0 and will be translated to NULL then
func Persist(ctx context.Context, q pgxscan.Querier, values *Values, bline *baseline.Values, pol *policy.Values, device *device.DevAikRow, image string, bin string, tsc string, now time.Time) (*Row, error) {
	ctx, span := tel.Start(ctx, "evidence.Persist")
	defer span.End()
	var row Row

	valsdoc, err := database.NewDocument(*values)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("evidenceSize", valsdoc.SizeSanitized()))

	blrow, err := baseline.ToRow(bline)
	if err != nil {
		return nil, err
	}
	polcol, err := pol.Serialize()
	if err != nil {
		return nil, err
	}

	err = pgxscan.Get(ctx, q, &row, `
    insert into v2.evidence (
      signed_by,
      received_at,
      values,
      baseline,
      policy,
      image_ref,
      binarly_ref,
      inteltsc_ref,
	  device_id
    ) select 
      $1,
      $2::timestamptz,
      $3,
      $4,
      $5,
      nullif($6, ''),
      nullif($7, ''),
      nullif($8, ''),
	  nullif($9, 0)
    returning *
  `,
		&device.AIK.QName, // $1
		now,               // $2
		valsdoc,           // $3
		blrow,             // $4
		polcol,            // $5
		image,             // $6
		bin,               // $7
		tsc,               // $8
		device.Id,         // $9
	)
	if err != nil {
		return nil, database.Error(err)
	}

	return &row, nil
}

func IsReadyAndLatest(ctx context.Context, q pgxscan.Querier, row *Row, maxAge time.Time) (bool, error) {
	ctx, span := tel.Start(ctx, "evidence.IsReadyAndLatest")
	defer span.End()
	refs := []string{}
	if row.BinarlyReference != nil {
		refs = append(refs, *row.BinarlyReference)
	}
	if row.IntelTSCReference != nil {
		refs = append(refs, *row.IntelTSCReference)
	}

	if len(refs) == 0 {
		return true, nil
	}

	isReady, err := queue.AreReady(ctx, q, refs, maxAge)
	if err != nil {
		return false, err
	}
	if !isReady {
		return false, nil
	}

	var isLatest bool
	err = pgxscan.Get(ctx, q, &isLatest, `
    select
      ev.received_at < $1::timestamptz
    from v2.evidence ev
    where ev.signed_by = $2::bytea
      and ev.id != $3::v2.ksuid
      and exists (
        select 1 from v2.appraisals a where a.evidence_id = ev.id)
    order by ev.received_at desc
    limit 1
    `, row.ReceivedAt, row.SignedBy, row.Id)
	if pgxscan.NotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, database.Error(err)
	}

	return isLatest, nil
}

func ByReference(ctx context.Context, q pgxscan.Querier, ref string) ([]Row, error) {
	var rows []Row

	err := pgxscan.Select(ctx, q, &rows, `
    select 
      *
    from v2.evidence
    where v2.evidence.inteltsc_ref = $1
       or v2.evidence.binarly_ref = $1
       or v2.evidence.image_ref = $1
    `, ref)
	if err != nil {
		return nil, database.Error(err)
	}
	return rows, nil
}

func MostRecent(ctx context.Context, q pgxscan.Querier, dev string, orgId int64) (*Row, error) {
	var row Row

	err := pgxscan.Get(ctx, q, &row, `
    select 
      v2.evidence.*
    from v2.evidence
    inner join v2.keys on v2.keys.fpr = v2.evidence.signed_by
    inner join v2.devices on v2.devices.id = v2.keys.device_id
    where v2.devices.id = $1::bigint
      and v2.devices.organization_id = $2::bigint
      and exists (select 1 from v2.appraisals where v2.appraisals.evidence_id in (v2.evidence.id))
    order by v2.evidence.received_at desc
    limit 1
    `, dev, orgId)
	if err != nil {
		return nil, database.Error(err)
	}
	return &row, nil
}
