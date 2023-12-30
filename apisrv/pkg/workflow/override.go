package workflow

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrNoAppraisals = errors.New("no appraisals")
)

func Override(ctx context.Context, tx pgx.Tx, store *blob.Storage, id int64, orgId int64, overrides []string, serviceName string, now time.Time, actor string) (*device.Row, error) {
	// fetch device
	devrow, err := device.Get(ctx, tx, id, orgId, now)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch device")
		return nil, err
	}

	// fetch latest evidence
	ev, err := evidence.MostRecent(ctx, tx, strconv.FormatInt(id, 10), orgId)
	if err == database.ErrNotFound {
		return devrow, nil
	}
	if err != nil {
		tel.Log(ctx).WithError(err).Error("fetch latest evidence")
		return nil, err
	}

	// fetch all dependent tasks
	subj, err := buildSubject(ctx, tx, store, ev, true /* cache may has been emptied */)
	if err != nil {
		return nil, err
	}

	// update baseline
	check.Override(ctx, overrides, subj)

	// reattest against new baseline, writes baseline to db
	_, err = appraiseOne(ctx, tx, store, ev, subj, serviceName, now)
	if err != nil {
		return nil, err
	}

	// refetch device with new state
	devrow, err = device.Get(ctx, tx, id, orgId, now)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("refetch device")
		return nil, err
	}

	return devrow, nil
}
