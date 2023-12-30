package workflow

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/event"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

func Enroll(ctx context.Context, pool *pgxpool.Pool, enrollment *api.Enrollment, ca *ecdsa.PrivateKey, caKid string, serviceName string, org string, now time.Time) (int64, []*api.EncryptedCredential, error) {
	ctx, span := tel.Start(ctx, "workflow.Enroll")
	defer span.End()

	// timeout for database operation
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var (
		creds []*api.EncryptedCredential
		devid int64
	)
	err := database.QueryIsolatedRetry(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		// enroll
		devid, creds, err = device.Enroll(ctx, tx, ca, caKid, *enrollment, serviceName, org, "(No actor)", now)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("enroll device")
			return err
		}

		// get organization id for quota
		orgRow, err := organization.Get(ctx, tx, org)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("get org row")
			return err
		}

		// quota
		current, allowed, err := organization.Quota(ctx, tx, orgRow.Id)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("fetch quota")
			return err
		}
		if current > allowed {
			tel.Log(ctx).
				WithFields(log.Fields{"current": current, "allowed": allowed}).
				Error("over quota")
			return ErrQuotaExceeded
		}

		// queue event
		_, err = event.BillingUpdate(ctx, tx, serviceName, map[string]int{org: current}, now, now)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("queue event")
			return err
		}

		return nil
	})

	return devid, creds, err
}
