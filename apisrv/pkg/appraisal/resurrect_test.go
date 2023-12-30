package appraisal

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestResurrect(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		// device.Resurrect()
		"ResurrectTrusted":      testResurrectTrusted,
		"ResurrectVuln":         testResurrectVuln,
		"ResurrectUnseen":       testResurrectUnseen,
		"ResurrectOutdated":     testResurrectOutdated,
		"ResurrectActive":       testResurrectActive,
		"ResurrectNotLast":      testResurrectNotLast,
		"ResurrectWrongOrg":     testResurrectWrongOrg,
		"ResurrectNonExistent":  testResurrectNonExistent,
		"ResurrectMultipleKeys": testResurrectMultipleKeys,
		"ResurrectNoAIK":        testResurrectNoAIK,
		"ResurrectNoPolicies":   testResurrectNoPolicies,
	}
	//XXX how about only instancing postgres once per package
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

func testResurrectTrusted(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 104, int64(100), false, nil, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}
	dev, err := device.Get(ctx, tx, 104, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}
	if dev.State != api.StateTrusted {
		t.Fatal("device not trusted")
	}
}

func testResurrectVuln(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 105, int64(100), false, nil, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}
	dev, err := device.Get(ctx, tx, 105, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}
	if dev.State != api.StateVuln {
		t.Fatal("device is trusted")
	}
}

func testResurrectUnseen(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 106, int64(100), false, nil, now, "unittest-actor")
	assert.NoError(t, err)

	dev, err := device.Get(ctx, tx, 106, int64(100), now)
	if assert.NoError(t, err) {
		assert.Equal(t, dev.State, api.StateUnseen)
	}
}

func testResurrectOutdated(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 107, int64(100), false, nil, now, "unittest-actor")
	assert.NoError(t, err)

	dev, err := device.Get(ctx, tx, 107, int64(100), now)
	if assert.NoError(t, err) {
		assert.Equal(t, dev.State, api.StateOutdated)
	}
}

func testResurrectActive(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 101, int64(100), false, nil, now, "unittest-actor")
	if err == nil {
		t.Fatal("can resurrect device that is replaced by an active one")
	}
}

func testResurrectNotLast(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 111, int64(100), false, nil, now, "unittest-actor")
	assert.NoError(t, err)

	dev, err := device.Get(ctx, tx, 111, int64(100), now)
	if assert.NoError(t, err) {
		if dev.State != api.StateTrusted {
			t.Fatalf("device isnt trusted: %s", dev.State)
		}
	}
}

func testResurrectWrongOrg(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 104, int64(2342), false, nil, now, "unittest-actor")
	if err == nil {
		t.Fatal("accepted wrong org")
	}
}

func testResurrectNonExistent(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 999, int64(100), false, nil, now, "unittest-actor")
	if err == nil {
		t.Fatal("accepted wrong id")
	}
}

func testResurrectMultipleKeys(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 120, int64(100), false, nil, now, "unittest-actor")
	assert.NoError(t, err)

	dev, err := device.Get(ctx, tx, 120, int64(100), now)
	if assert.NoError(t, err) {
		assert.Equalf(t, api.StateTrusted, dev.State, "device not trusted")
	}
}

func testResurrectNoAIK(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 121, int64(100), false, nil, now, "unittest-actor")
	assert.NoError(t, err)

	dev, err := device.Get(ctx, tx, 121, int64(100), now)
	if assert.NoError(t, err) {
		// to make device.Get more efficient StateUnseen is returned instead of StateNew
		if dev.State != api.StateUnseen {
			t.Fatalf("device isnt new: %s", dev.State)
		}
	}
}

func testResurrectNoPolicies(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = device.Retire(ctx, tx, 122, int64(100), false, nil, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}
	dev, err := device.Get(ctx, tx, 122, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}

	if dev.State != api.StateUnseen {
		t.Fatal("device not unseen")
	}
}
