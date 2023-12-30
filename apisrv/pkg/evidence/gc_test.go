package evidence

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

func testGC(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	assert.NoError(t, err)

	now := time.Now()
	rng := mrand.New(mrand.NewSource(42))
	ek := api.GenerateEK(rng)
	root, _ := api.GenerateIdentityKey(rng)
	aik, attest, sig, _ := api.GenerateDeviceKey(root, rng)
	enrollment := &api.Enrollment{
		NameHint:               "localhost",
		EndorsementKey:         api.PublicKey(ek),
		EndorsementCertificate: nil,
		Root:                   api.PublicKey(root),
		Keys: map[string]api.Key{
			"aik": {
				Public:                 api.PublicKey(aik),
				CreationProof:          api.Attest(attest),
				CreationProofSignature: api.Signature(sig),
			},
		},
		Cookie: "hello",
	}
	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)
	devid, _, err := device.Enroll(ctx, tx, caPriv, caKid, *enrollment, "testsrv", "ext-1", "Kai", now)
	assert.NoError(t, err)
	tx.Commit(ctx)

	values := Values{Type: ValuesType}
	bline := baseline.New()
	pol := policy.New()
	ref := "aaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var aikname api.Name
	db.QueryRow(ctx, "select fpr from v2.keys").Scan(&aikname)
	assert.NoError(t, err)
	dev := device.DevAikRow{Id: devid, AIK: &device.KeysRow{QName: aikname}}

	// now evidence
	row1, err := Persist(ctx, db, &values, bline, pol, &dev, ref, ref, ref, now)
	assert.NoError(t, err)
	assert.NotEmpty(t, row1.Id)

	// old evidence
	row2, err := Persist(ctx, db, &values, bline, pol, &dev, ref, ref, ref, now.AddDate(-1, 0, 0))
	assert.NoError(t, err)
	assert.NotEmpty(t, row2.Id)

	// old evidence
	row3, err := Persist(ctx, db, &values, bline, pol, &dev, ref, ref, ref, now.AddDate(-1, 1, 0))
	assert.NoError(t, err)
	assert.NotEmpty(t, row3.Id)

	// link appraisal to evidence (prevents deletion by GC)
	// XXX: we can't use regular functions from appraisal pkg here because it would produce an import cycle
	_, err = db.Exec(ctx, `
    insert into v2.appraisals (evidence_id, received_at, appraised_at, expires, verdict, report, device_id)
    values (
      $1::v2.ksuid,
      $2::timestamptz,
      $2::timestamptz,
      $2::timestamptz + '2 days',
      '{ "type": "verdict/2" }'::jsonb,
      '{ "type": "report/2" }'::jsonb,
      $3
    )
    `, row2.Id, row2.ReceivedAt, devid)
	assert.NoError(t, err)

	tx, err = db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	assert.NoError(t, err)
	err = GarbageCollect(ctx, tx)
	assert.NoError(t, err)
	tx.Commit(ctx)

	// check that only the old evidence was deleted
	var rows []Row
	err = pgxscan.Select(ctx, db, &rows, "select * from v2.evidence order by received_at desc")
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, row1.Id, rows[0].Id)
	assert.Equal(t, row2.Id, rows[1].Id)
}
