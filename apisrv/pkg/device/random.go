package device

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"

	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func Random(rng *rand.Rand, pool *pgxpool.Pool, state string, org int64, numAppraisals int, numPreviousDevices int, replacedBy *int64, now time.Time) int64 {
	var dev Row

	ctx := context.Background()
	err := pgxscan.Get(ctx, pool, &dev, `
		INSERT INTO v2.devices (
			id,
			hwid,
			fpr,
			name,
			baseline,
			cookie,
			organization_id,
			replaced_by,
			retired
		) VALUES (DEFAULT, $1, $2, $3,$4, $5, $6, $7, $8)
		RETURNING *
	`,
		api.Name(api.GenerateName(rng)),
		api.Name(api.GenerateName(rng)),
		fmt.Sprintf("Test device #%d", rng.Intn(100)),
		baseline.New(),
		test.GenerateBase64(rng, 32, 64),
		org,
		replacedBy,
		state == "retired")
	if err != nil {
		panic(err)
	}

	var key *KeysRow

	if state != "new" {
		key = &KeysRow{}
		err := pgxscan.Get(ctx, pool, key, `
			INSERT INTO v2.keys (
				id,
				public,
				name,
				fpr,
				credential,
				device_id
			) VALUES (DEFAULT, $1, $2, $3, $4, $5)
			RETURNING *
		`,
			api.PublicKey(api.GeneratePublic(rng)),
			"aik",
			api.Name(api.GenerateName(rng)),
			"XXX",
			dev.Id)
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < numAppraisals; i += 1 {
		var verdict string
		if state == "vulnerable" && i == 0 {
			verdict = api.Vulnerable
		} else if state == "trusted" && i == 0 {
			verdict = api.Trusted
		} else if rng.Intn(100) > 10 {
			verdict = api.Trusted
		} else {
			verdict = api.Vulnerable
		}
		verdictDoc, err := database.NewDocument(api.Verdict{
			Type:               api.VerdictType,
			Result:             verdict,
			SupplyChain:        verdict,
			Configuration:      verdict,
			Firmware:           verdict,
			Bootloader:         verdict,
			OperatingSystem:    verdict,
			EndpointProtection: verdict,
		})
		if err != nil {
			panic(err)
		}
		reportDoc, err := database.NewDocument(api.Report{Type: api.ReportType})
		if err != nil {
			panic(err)
		}

		_, err = pool.Exec(ctx, `
				INSERT INTO v2.appraisals (
					id,
					received_at,
					appraised_at,
					expires,
					verdict,
					report,
					key_id,
					device_id
				) VALUES (DEFAULT, $1, $1, $2, $3, $4, $5, $6)
			`,
			now.Add((time.Duration(-1*i)*time.Hour)-(30*time.Minute)),
			now.Add(time.Hour*time.Duration(-1*i)),
			verdictDoc, reportDoc,
			key.Id,
			key.DeviceId)
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < numPreviousDevices; i += 1 {
		Random(rng, pool, "retired", org, numAppraisals, 0, &dev.Id, now.Add(time.Hour*time.Duration(-1000*i)))
	}

	return dev.Id
}
