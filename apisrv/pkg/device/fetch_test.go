package device

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestFetch(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		// List()
		"List":           testListDevices,
		"ListByIterator": testListByIterator,
		// Set()
		"GetSet": testGetSet,
		// Get()
		"GetSingle":      testGetSingle,
		"GetNonExistent": testGetNonExistentDevice,
		"GetWrongOrg":    testGetWrongOrg,
		// Fingerprint()
		"GetFpr":            testGetFpr,
		"GetNonExistentFpr": testGetNonExistentFpr,
		"GetRetiredDevFpr":  testGetRetiredDevFpr,
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

func testListDevices(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	rng := rand.New(rand.NewSource(int64(os.Getpid())))

	for i := 0; i < 100; i += 1 {
		Random(rng, db, "trusted", 100, 10, 0, nil, now)
	}

	conn := database.ExplainQuerier{Database: db, MaxCost: 97}

	// fetch all
	devs, _, err := ListRow(ctx, conn, nil, 10, int64(100), now)
	assert.NoError(t, err)
	assert.Lenf(t, devs, 10, "expected 10 devices, got %d\n", len(devs))

	// fetch one
	devs, next, err := ListRow(ctx, conn, nil, 1, int64(100), now)
	assert.NoError(t, err)
	assert.Lenf(t, devs, 1, "expected 1 device, got %d\n", len(devs))
	assert.NotNil(t, next, "expected iterator")

	// fetch next
	devs, _, err = ListRow(ctx, conn, next, 1, int64(100), now)
	assert.NoError(t, err)
	assert.Lenf(t, devs, 1, "expected 1 device")
}

func testGetSet(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	conn := database.ExplainQuerier{Database: db, MaxCost: 94}

	// get empty set
	devs, err := SetRow(ctx, conn, []int64{}, int64(100), now)
	assert.NoError(t, err)
	assert.Emptyf(t, devs, "expected empty list")

	// get dup IDs set
	devs, err = SetRow(ctx, conn, []int64{100, 100, 100, 100, 100, 100}, int64(100), now)
	assert.NoError(t, err)

	if len(devs) != 1 || devs[0].Id != 100 {
		t.Fatal("expected list with single device with Id = 100")
	}

	// get non-existing IDs set
	devs, err = SetRow(ctx, conn, []int64{99, 999, 9999}, int64(100), now)
	assert.NoError(t, err)
	assert.Emptyf(t, devs, "expected empty list")

	// get partially non-existing IDs set
	devs, err = SetRow(ctx, conn, []int64{100, 99, 91}, int64(100), now)
	assert.NoError(t, err)

	if len(devs) != 1 || devs[0].Id != 100 {
		t.Fatal("expected list with single device with Id = 100")
	}
}

func testListByIterator(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	rng := rand.New(rand.NewSource(int64(os.Getpid())))

	for i := 0; i < 1000; i += 1 {
		Random(rng, db, "trusted", 100, 2, 0, nil, now)
	}

	conn := database.ExplainQuerier{Database: db, MaxCost: 400}

	// get id list
	start := "1500"
	devs, next, err := ListRow(ctx, conn, &start, 100, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}

	if len(devs) != 100 {
		t.Fatal("expected 100 devices")
	}

	if next == nil || *next != "1400" {
		t.Fatalf("iter %s, expected 1400", *next)
	}

	k := 0
	for j := (int64)(1500); j < 1400; j -= 1 {
		if devs[k].Id != j {
			t.Fatalf("expected device #%d to have ID %d\n", k, j)
		}
		k += 1
	}

	// get small id list
	devs, next, err = ListRow(ctx, conn, &start, 1, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}

	if len(devs) != 1 {
		t.Fatal("expected a single device")
	}

	if next == nil || *next != "1499" {
		t.Fatalf("iter %#v, expected 1499", next)
	}

	if devs[0].Id != 1500 {
		t.Fatalf("expected device Id to be 1500")
	}

	// get 0 id list
	start = "0"
	devs, next, err = ListRow(ctx, conn, &start, 1, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}

	if next != nil {
		t.Fatalf("iter %#v, expected nil", next)
	}

	if len(devs) != 0 {
		t.Fatal("expected empty device list")
	}

	// get -1 id list
	start = "-1"
	devs, next, err = ListRow(ctx, conn, &start, 1, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}

	if next != nil {
		t.Fatalf("iter %#v, expected nil", next)
	}

	if len(devs) != 0 {
		t.Fatal("expected empty device list")
	}

	// get all ids as small batches
	var cursor *string = nil
	var last int64
	allDevs := []Row{}
	for {
		batch, next, err := ListRow(ctx, conn, cursor, 10, int64(100), now)
		if err != nil {
			t.Fatal(err)
		}

		allDevs = append(allDevs, batch...)

		if next == nil {
			break
		}

		if len(batch) == 0 {
			t.Fatalf("got empty batch")
		}

		if last == batch[0].Id {
			t.Fatalf("didn't make any progress with %#v\n", cursor)
		}
		last = batch[0].Id

		cursor = next
	}

	if len(allDevs) != 1000+18 {
		t.Fatalf("expected 1018 devices, got %d", len(allDevs))
	}

	k = 0
	for j := (int64)(2000); j < 1000; j -= 1 {
		if devs[k].Id != j {
			t.Fatalf("expected device #%d to have ID %d\n", k, j)
		}
		k += 1
	}

	// get all ids in one large batch
	cursor = nil
	allDevs2, next, err := ListRow(ctx, conn, cursor, 9999, int64(100), now)
	if err != nil {
		t.Fatal(err)
	}

	if cursor != nil {
		t.Fatalf("iter %#v, expected nil", next)
	}

	if len(allDevs2) != 1000+18 {
		t.Fatalf("expected 1018 devices, got %d", len(allDevs2))
	}

	for i, d1 := range allDevs {
		if (allDevs2)[i].Id != d1.Id {
			t.Fatalf("#%d: expected device %d to be equal to %d\n", i, allDevs2[i].Id, d1.Id)
		}
	}
}

func testGetSingle(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	conn := database.ExplainQuerier{Database: db, MaxCost: 94.0}

	dev, err := Get(ctx, conn, 100, int64(100), now)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(100), dev.Id)
	}
}

func testGetNonExistentDevice(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	dev, err := Get(ctx, tx, 999, int64(100), now)
	if err == nil {
		t.Fatal("expected error")
	}
	if dev != nil {
		t.Fatal("expected nil device")
	}
}

func testGetWrongOrg(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	dev, err := Get(ctx, tx, 100, int64(2342), now)
	if err == nil {
		t.Fatal("expected error")
	}
	if dev != nil {
		t.Fatal("expected nil device")
	}
}

func testGetFpr(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	conn := database.ExplainQuerier{Database: db, MaxCost: 74}

	raw, err := hex.DecodeString("0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a")
	assert.NoError(t, err)
	name, err := tpm2.DecodeName(bytes.NewBuffer(raw))
	assert.NoError(t, err)

	dev, err := GetByFingerprint(ctx, conn, (*api.Name)(name))
	if assert.NoError(t, err) {
		assert.Equal(t, int64(134), dev.Id)
	}
}

func testGetRetiredDevFpr(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	raw, err := hex.DecodeString("0022000bf7f5c2fac339f9cc41f2fadd91c0d86ba0b0c9a05bd4e43cc78761693ccdc9a7")
	assert.NoError(t, err)
	name, err := tpm2.DecodeName(bytes.NewBuffer(raw))
	assert.NoError(t, err)

	dev, err := GetByFingerprint(ctx, tx, (*api.Name)(name))
	if assert.NoError(t, err) {
		assert.True(t, dev.Retired)
	}
}

func testGetNonExistentFpr(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	raw, err := hex.DecodeString("0022000b305c1823252de4480e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a")
	assert.NoError(t, err)
	name, err := tpm2.DecodeName(bytes.NewBuffer(raw))
	assert.NoError(t, err)

	dev, err := GetByFingerprint(ctx, tx, (*api.Name)(name))
	assert.Error(t, err)
	assert.Nil(t, dev)
}
