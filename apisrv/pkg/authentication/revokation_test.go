package authentication

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

//go:embed seed.sql
var seedSql string

func TestRevokations(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"RevokeToken": testRevokeToken,
		"RevokeKey":   testRevokeKey,
		"RevokeTwice": testRevokeTwice,
		"RevokeGC":    testRevokeGC,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 16)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testRevokeToken(t *testing.T, db *pgxpool.Pool) {
	k, priv := test.NewTestKeyset("test", t)
	ctx := context.Background()
	kid := k.KeysFor("test")[0].Kid

	token, err := IssueServiceCredential("apisrv", nil, time.Now(), kid, priv)
	assert.NoError(t, err)

	claims := Credential{}
	p := jwt.Parser{}
	_, _, err = p.ParseUnverified(token, &claims)
	assert.NoError(t, err)

	revoked, err := IsRevoked(ctx, db, token)
	assert.NoError(t, err)
	assert.False(t, revoked)

	err = RevokeToken(ctx, db, claims.Id, time.Unix(claims.ExpiresAt, 0))
	assert.NoError(t, err)

	revoked, err = IsRevoked(ctx, db, token)
	assert.NoError(t, err)
	assert.True(t, revoked)

	token2, err := IssueServiceCredential("apisrv", nil, time.Now(), kid, priv)
	assert.NoError(t, err)
	revoked, err = IsRevoked(ctx, db, token2)
	assert.NoError(t, err)
	assert.False(t, revoked)
}

func testRevokeKey(t *testing.T, db *pgxpool.Pool) {
	k, priv := test.NewTestKeyset("test", t)
	ctx := context.Background()
	kid := k.KeysFor("test")[0].Kid

	err := RevokeKey(ctx, db, kid, time.Now().Add(time.Hour*30*24))
	assert.NoError(t, err)

	token, err := IssueServiceCredential("apisrv", nil, time.Now(), kid, priv)
	claims := Credential{}
	p := jwt.Parser{}
	_, _, err = p.ParseUnverified(token, &claims)
	assert.NoError(t, err)

	revoked, err := IsRevoked(ctx, db, token)
	assert.NoError(t, err)
	assert.True(t, revoked)

	err = RevokeToken(ctx, db, claims.Id, time.Unix(claims.ExpiresAt, 0))
	assert.NoError(t, err)

	revoked, err = IsRevoked(ctx, db, token)
	assert.NoError(t, err)
	assert.True(t, revoked)
}

func testRevokeTwice(t *testing.T, db *pgxpool.Pool) {
	k, priv := test.NewTestKeyset("test", t)
	ctx := context.Background()
	kid := k.KeysFor("test")[0].Kid

	err := RevokeKey(ctx, db, kid, time.Now().Add(time.Hour*30*24))
	assert.NoError(t, err)
	err = RevokeKey(ctx, db, kid, time.Now().Add(time.Hour*30*24))
	assert.NoError(t, err)

	token, err := IssueServiceCredential("apisrv", nil, time.Now(), kid, priv)
	assert.NoError(t, err)
	claims := Credential{}
	p := jwt.Parser{}
	_, _, err = p.ParseUnverified(token, &claims)
	assert.NoError(t, err)

	err = RevokeToken(ctx, db, claims.Id, time.Unix(claims.ExpiresAt, 0))
	assert.NoError(t, err)
	err = RevokeToken(ctx, db, claims.Id, time.Unix(claims.ExpiresAt, 0))
	assert.NoError(t, err)
}

func testRevokeGC(t *testing.T, db *pgxpool.Pool) {
	k, priv := test.NewTestKeyset("test", t)
	ctx := context.Background()
	kid := k.KeysFor("test")[0].Kid

	err := RevokeKey(ctx, db, kid, time.Now().Add(time.Hour*-1))
	assert.NoError(t, err)

	token, err := IssueServiceCredential("apisrv", nil, time.Now(), kid, priv)
	assert.NoError(t, err)
	claims := Credential{}
	p := jwt.Parser{}
	_, _, err = p.ParseUnverified(token, &claims)
	assert.NoError(t, err)

	err = RevokeToken(ctx, db, claims.Id, time.Now().Add(time.Hour*-1))
	assert.NoError(t, err)

	err = GarbageCollectRevokations(ctx, db, time.Now())
	assert.NoError(t, err)

	revoked, err := IsRevoked(ctx, db, token)
	assert.NoError(t, err)
	assert.False(t, revoked)
}
