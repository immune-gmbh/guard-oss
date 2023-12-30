package device

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
)

func TestEnroll(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	rng := mrand.New(mrand.NewSource(42))
	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool, *api.Enrollment){
		"EnrollExisting": testEnrollExisting,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 16)

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

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn, enrollment)
		})
	}
}

func testEnrollExisting(t *testing.T, conn *pgxpool.Pool, enrollment *api.Enrollment) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	// 1st enroll
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	devid, _, err := Enroll(ctx, tx, caPriv, caKid, *enrollment, "testsrv", "ext-1", "Kai", now)
	assert.NoError(t, err)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	row, err := Get(ctx, conn, devid, 0, now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateUnseen, row.State)
	assert.False(t, row.Retired)
	assert.Nil(t, row.ReplacedBy)

	// 2nd enroll
	tx, err = conn.Begin(ctx)
	assert.NoError(t, err)
	devid2, _, err := Enroll(ctx, tx, caPriv, caKid, *enrollment, "testsrv", "ext-1", "Kai", now)
	assert.NoError(t, err)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	row, err = Get(ctx, conn, devid, 0, now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateRetired, row.State)
	assert.True(t, row.Retired)
	assert.NotNil(t, row.ReplacedBy)

	row2, err := Get(ctx, conn, devid2, 0, now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateUnseen, row2.State)
	assert.False(t, row2.Retired)
	assert.Nil(t, row2.ReplacedBy)
	assert.Equal(t, row2.Id, *(row.ReplacedBy))
}
