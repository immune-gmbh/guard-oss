package change

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	mrand "math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestGC(t *testing.T) {
	ctx := context.Background()
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
	act := "Kai"
	// enroll will also add a change
	devid, _, err := device.Enroll(ctx, tx, caPriv, caKid, *enrollment, "testsrv", "ext-1", act, now)
	assert.NoError(t, err)

	devstrid := strconv.FormatInt(devid, 10)

	// new change
	c1, err := New(ctx, tx, "rename", nil, "ext-1", &devstrid, &act, time.Now())
	assert.NoError(t, err)

	// old change
	_, err = New(ctx, tx, "rename", nil, "ext-1", &devstrid, &act, time.Now().AddDate(-1, 0, 0))
	assert.NoError(t, err)

	err = GarbageCollect(ctx, tx)
	assert.NoError(t, err)
	tx.Commit(ctx)

	// check that only the old change was deleted
	var rows []Row
	err = pgxscan.Select(ctx, db, &rows, "select * from v2.changes order by timestamp desc")
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, c1.Id, strconv.FormatInt(rows[0].Id, 10))
	assert.Equal(t, "enroll", rows[1].Type)
}
