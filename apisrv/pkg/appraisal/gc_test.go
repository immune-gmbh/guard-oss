package appraisal

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestGC(t *testing.T) {
	ctx := context.Background()
	//XXX how about only instancing postgres once per package
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)
	pgsqlC.Reset(t, ctx, database.MigrationFiles)
	db := pgsqlC.Connect(t, ctx)
	defer db.Close()
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

	values := evidence.Values{Type: evidence.ValuesType}
	bline := baseline.New()
	pol := policy.New()
	ref := "aaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var aikname api.Name
	db.QueryRow(ctx, "select fpr from v2.keys").Scan(&aikname)
	assert.NoError(t, err)
	dev := device.DevAikRow{Id: devid, AIK: &device.KeysRow{QName: aikname}}

	// old evidence
	ev, err := evidence.Persist(ctx, db, &values, bline, pol, &dev, ref, ref, ref, now.AddDate(-1, -1, 0))
	assert.NoError(t, err)
	assert.NotEmpty(t, ev.Id)

	// get key for device
	devaikrow, err := device.GetByFingerprint(ctx, db, &aikname)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	tx, err = db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	assert.NoError(t, err)

	// appraise now evidence
	rep := new(api.Report)
	rep.Type = api.ReportType
	id1, err := Create(ctx, tx, ev, rep, makeIssuesTestData(), devaikrow.AIK, devid, now, "test-actor")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, id1)
	}

	// appraise old evidence
	id2, err := Create(ctx, tx, ev, rep, makeIssuesTestData(), devaikrow.AIK, devid, now.AddDate(-1, 0, 0), "test-actor")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, id2)
	}

	err = GarbageCollect(ctx, tx)
	assert.NoError(t, err)
	tx.Commit(ctx)

	// check that only the old appraisal was deleted
	var rows []Row
	err = pgxscan.Select(ctx, db, &rows, "select id from v2.appraisals order by appraised_at desc")
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, id1, rows[0].Id)

	// verify old issues are also removed
	issues, err := GetIssuesByAppraisal(ctx, db, id2)
	assert.NoError(t, err)
	assert.Len(t, issues, 0)

	// others must still be there
	issues, err = GetIssuesByAppraisal(ctx, db, id1)
	assert.NoError(t, err)
	assert.Len(t, issues, 2)
}
