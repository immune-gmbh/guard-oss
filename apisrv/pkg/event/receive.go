package event

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/appraisal"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/change"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

func Receive(ctx context.Context, pool *pgxpool.Pool, serviceName string, ev *ce.Event, now time.Time) error {
	ctx, span := tel.Start(ctx, "Receive")
	defer span.End()

	switch ev.Type() {
	case api.QuotaUpdateEventType:
		var q api.QuotaUpdateEvent

		err := json.Unmarshal(ev.Data(), &q)
		if err != nil {
			return err
		}
		return receiveQuotaUpdate(ctx, pool, ev.Subject(), &q, now)
	case api.HeartbeatEventType:
		var q api.HeartbeatEvent

		err := json.Unmarshal(ev.Data(), &q)
		if err != nil {
			return err
		}
		return receiveHeartbeat(ctx, pool, serviceName, &q)

	case api.RevokeCredentialsEventType:
		var q api.RevokeCredentialsEvent

		err := json.Unmarshal(ev.Data(), &q)
		if err != nil {
			return err
		}
		return receiveRevokeCredentials(ctx, pool, &q)

	default:
		return errors.New("unknow event type")
	}
}

func receiveQuotaUpdate(ctx context.Context, pool *pgxpool.Pool, org string, ev *api.QuotaUpdateEvent, now time.Time) error {
	return organization.UpdateQuota(ctx, pool, org, ev.Devices, now)
}

func receiveHeartbeat(ctx context.Context, pool *pgxpool.Pool, serviceName string, ev *api.HeartbeatEvent) error {
	//now := time.Now()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// GC on append-only tables needs to run in correct order (b/c of forgein key constraints) to be effective
	// in particular, appraisals needs to run before evidence in
	err = appraisal.GarbageCollect(ctx, tx)
	if err != nil {
		return err
	}
	err = evidence.GarbageCollect(ctx, tx)
	if err != nil {
		return err
	}
	err = change.GarbageCollect(ctx, tx)
	if err != nil {
		return err
	}
	err = queue.GarbageCollect(ctx, tx)
	if err != nil {
		return err
	}

	if ev.CompressRevocations {
		//err := auth.GarbageCollectRevokations(ctx, tx)
		//if err != nil {
		//	tel.Log(ctx).WithError(err).Error("gc revokations")
		//	return err
		//}
	}

	//if ev.ExpireAppraisals {
	//	apprs, orgs, err := appraisal.Expire(ctx, tx, now)
	//	if err != nil {
	//		tel.Log(ctx).WithError(err).Error("expire appraisals")
	//		return err
	//	}
	//	if len(apprs) != len(orgs) {
	//		return errors.New("invalid value")
	//	}

	//	for i := 0; i < len(apprs); i += 1 {
	//		appr := apprs[i]
	//		org := orgs[i]

	//		tx, err := pool.Begin(ctx)
	//		if err != nil {
	//			tel.Log(ctx).WithError(err).Error("begin tx")
	//			return err
	//		}
	//		defer tx.Rollback(ctx)

	//		_, err = AppraisalExpired(ctx, tx, serviceName, org, appr.Device, appr, now)
	//		if err != nil {
	//			tel.Log(ctx).WithError(err).Error("create event")
	//			return err
	//		}

	//		err = tx.Commit(ctx)
	//		if err != nil {
	//			tel.Log(ctx).WithError(err).Error("commit tx")
	//			return err
	//		}
	//	}
	//}

	if ev.ReportUsage {
		// nop
	}

	return tx.Commit(ctx)
}

func receiveRevokeCredentials(ctx context.Context, pool *pgxpool.Pool, ev *api.RevokeCredentialsEvent) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	//for _, tok := range ev.TokenIDs {
	//	err := auth.RevokeToken(ctx, pool, tok, ev.Expires)
	//	if err != nil {
	//		tel.Log(ctx).WithError(err).WithField("jti", tok).Error("revoke token")
	//		return err
	//	}
	//}

	return tx.Commit(ctx)
}
