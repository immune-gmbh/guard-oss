package device

import (
	"context"
	"testing"
	"time"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestCRUD(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		// Patch()
		"PatchName":              testPatchName,
		"PatchNameToEmptyString": testPatchNameToEmptyString,
		"PatchNothing":           testPatchNothing,
		"PatchPolicy":            testPatchPolicy,
		"PatchDeviceWrongOrg":    testPatchDeviceWrongOrg,
		"PatchNonExistentDevice": testPatchNonExistentDevice,
		// Retire()
		"Retire":            testRetire,
		"RetireRetired":     testRetireRetired,
		"RetireNonExistent": testRetireNonExistent,
		"RetireWrongOrg":    testRetireWrongOrg,
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

func testPatchName(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	conn := database.ExplainQuerier{Database: db, MaxCost: 16.7}

	name := "New Device Name"
	err := Patch(ctx, conn, 100, int64(100), &name, nil, nil, now, "unittest-actor")
	assert.NoError(t, err)

	dev2, err := Get(ctx, db, 100, int64(100), now)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equalf(t, int64(100), dev2.Id, "wrong ID")
	assert.Equalf(t, name, dev2.Name, "name not patched. is %s should be %s", dev2.Name, name)

	// changes is likely being refactored and it really should be tested in its own package and not as part of device
	/*
		changes, err := change.Set(ctx, db, []string{dev2.Changes[0].Id}, int64(100))
		assert.NoError(t, err)
		assert.Equalf(t, "unittest-actor", *changes[0].Actor, "updated by not updated (heh)")
	*/
}

func testPatchNameToEmptyString(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer tx.Rollback(ctx)

	name := ""
	err = Patch(ctx, tx, 102, int64(100), &name, nil, nil, now, "unittest-actor")
	if err == nil {
		t.Fatal("expected error")
	}
	tx.Commit(ctx)

	tx, err = db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer tx.Rollback(ctx)

	dev2, err := Get(ctx, tx, 102, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}
	if dev2.Id != 102 {
		t.Fatal("wrong ID")
	}
	if dev2.Name == "" {
		t.Fatal("name was patched")
	}
}

func testPatchPolicy(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	conn := database.ExplainQuerier{Database: db, MaxCost: 16.7}

	pol := policy.New()
	pol.EndpointProtection = policy.False
	err := Patch(ctx, conn, 100, int64(100), nil, nil, pol, now, "unittest-actor")
	assert.NoError(t, err)

	dev, err := Get(ctx, db, 100, 0, now)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	pol2, err := dev.GetPolicy()
	assert.NoError(t, err)
	assert.Equal(t, pol, pol2)
}

func testPatchNothing(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	err = Patch(ctx, tx, 100, int64(100), nil, nil, nil, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}
}

func testPatchDeviceWrongOrg(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	name := "New Device Name"
	err = Patch(ctx, tx, 100, int64(2342), &name, nil, nil, now, "unittest-actor")
	if err == nil {
		t.Fatal("expected error")
	}
}

func testPatchNonExistentDevice(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	name := "New Device Name"
	err = Patch(ctx, tx, 999, int64(100), &name, nil, nil, now, "unittest-actor")
	if err == nil {
		t.Fatal("expected error")
	}
}

func testRetire(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	comment := "comment"
	err = Retire(ctx, tx, 103, int64(100), true, &comment, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}

	dev2, err := Get(ctx, tx, 103, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}
	if dev2.Id != 103 {
		t.Fatal("wrong ID")
	}
	if dev2.State != api.StateResurrectable {
		t.Fatalf("device not retired: %s", dev2.State)
	}
}

func testRetireRetired(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	comment := "comment"
	err = Retire(ctx, tx, 103, int64(100), true, &comment, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}

	err = Retire(ctx, tx, 103, int64(100), true, &comment, now, "unittest-actor")
	if err != nil {
		t.Fatal(err)
	}
	dev, err := Get(ctx, tx, 103, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}
	if dev.Id != 103 {
		t.Fatal("wrong ID")
	}
	if dev.State != api.StateResurrectable {
		t.Fatalf("device not retired: %s", dev.State)
	}
}

func testRetireNonExistent(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	comment := "comment"
	err = Retire(ctx, tx, 999, int64(100), true, &comment, now, "unittest-actor")
	if err == nil {
		t.Fatal("expected error")
	}
}

func testRetireWrongOrg(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	comment := "comment"
	err = Retire(ctx, tx, 103, int64(2342), true, &comment, now, "unittest-actor")
	if err == nil {
		t.Fatal("expected error")
	}
}
