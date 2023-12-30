// Package change provides ...
package change

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func Get(ctx context.Context, qq pgxscan.Querier, id int64, orgId int64) (*api.Change, error) {
	chs, err := Set(ctx, qq, []int64{id}, orgId)
	if len(chs) == 1 {
		return chs[0], err
	} else {
		return nil, err
	}
}

func Set(ctx context.Context, qq pgxscan.Querier, ids []int64, orgId int64) ([]*api.Change, error) {
	var rows []Row
	err := pgxscan.Select(ctx, qq, &rows, `
		--
		-- Fetch-Organization-Wide-Changes
		--
		SELECT c.id, c.type, c.actor, c.comment, c.timestamp, c.organization_id, c.device_id, c.key_id
		FROM v2.changes AS c
		WHERE c.id = any($2) AND c.organization_id = $1
	`, orgId, ids)
	if err != nil {
		return nil, err
	}

	var changes []*api.Change = make([]*api.Change, len(rows))
	for i, row := range rows {
		ch := FromRow(&row)
		changes[i] = &ch
	}

	return changes, nil
}

func parseIntRef(s *string) (*int64, error) {
	if s != nil {
		if i, err := strconv.ParseInt(*s, 10, 64); err == nil {
			return &i, nil
		} else {
			return nil, err
		}
	}

	return nil, nil
}

func List(ctx context.Context, qq pgxscan.Querier, iter *string, orgId int64, dev *string, limit int) ([]*api.Change, *string, error) {
	var start int64 = math.MaxInt64
	if iter != nil {
		i, err := strconv.ParseInt(*iter, 10, 64)
		if err != nil {
			return nil, nil, nil
		}
		start = i
	}
	devid, err := parseIntRef(dev)
	if err != nil {
		return nil, nil, err
	}

	var rows []Row
	err = pgxscan.Select(ctx, qq, &rows, `
		--
		-- Fetch-D
	  SELECT c.id, c.type, c.actor, c.comment, c.timestamp, c.organization_id, c.device_id, c.key_id
		FROM v2.changes AS c
		WHERE c.id < $3
		  AND c.organization_id = $1
			AND ($4::bigint IS NULL OR c.device_id = $4)
		ORDER BY c.id DESC
		LIMIT $2
	`, orgId, limit, start, devid)
	if err != nil {
		return nil, nil, database.Error(err)
	}
	if len(rows) == 0 {
		return nil, nil, nil
	}

	next := rows[0].Id
	for _, r := range rows {
		if next > r.Id {
			next = r.Id
		}
	}
	var nextStr *string
	if next > 0 {
		s := fmt.Sprintf("%d", next-1)
		nextStr = &s
	}

	chs := make([]*api.Change, len(rows))
	for i, r := range rows {
		ch := FromRow(&r)
		chs[i] = &ch
	}

	return chs, nextStr, nil
}

func New(ctx context.Context, tx pgx.Tx, ty string, cmnt *string, org string, device *string, act *string, now time.Time) (*api.Change, error) {
	dev, err := parseIntRef(device)
	if err != nil {
		return nil, err
	}
	ch := api.Change{
		Id:        "unset",
		Actor:     act,
		Timestamp: now,
		Comment:   cmnt,
		Type:      ty,
		Device:    nil,
	}
	if device != nil {
		ch.Device = &api.Device{Id: *device}
	}

	var id int64
	err = tx.QueryRow(ctx, `
	---
	--- New-Change
	---
	WITH organization AS (
		SELECT id FROM v2.organizations
		WHERE external = $1
		LIMIT 1
	)
	INSERT INTO v2.changes VALUES (
		DEFAULT,
		(SELECT id FROM organization),
		$2,
		$3,
		$4,
		$5,
		NULL,
		$6
	)
	RETURNING id
	`, org, ty, act, cmnt, dev, now).Scan(&id)
	if err != nil {
		return &ch, err
	}

	ch.Id = fmt.Sprintf("%d", id)
	return &ch, err
}
