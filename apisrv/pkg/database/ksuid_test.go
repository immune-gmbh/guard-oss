package database

import (
	"context"
	_ "embed"
	"sort"
	"testing"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestKSuid(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"SmokeTest": testSmokeTest,
		"Sorted":    testSorted,
		"Unique":    testUnique,
		"Edge":      testEdge,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.Reset(t, ctx, MigrationFiles)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testSmokeTest(t *testing.T, db *pgxpool.Pool) {
	var id string
	ctx := context.Background()
	err := pgxscan.Get(ctx, db, &id, "select v2.next_ksuid_v2()")
	assert.NoError(t, err)
	assert.Regexp(t, "^[a-zA-Z0-9]{27}$", id)

	// 1507608047
	// B5A1CD34B5F99D1154FB6853345C9735
	// 0ujtsYcgvSTl8PAuAdqWYSMnLOv
	err = pgxscan.Get(ctx, db, &id, "select v2.next_ksuid_v2(to_timestamp(1507608047)::timestamptz(3), '\\xB5A1CD34B5F99D1154FB6853345C9735')")
	assert.NoError(t, err)
	assert.Equal(t, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", id)

	// 1507610780
	// 73FC1AA3B2446246D6E89FCD909E8FE8
	// 0ujzPyRiIAffKhBux4PvQdDqMHY
	err = pgxscan.Get(ctx, db, &id, "select v2.next_ksuid_v2(to_timestamp(1507610780)::timestamptz(3), '\\x73FC1AA3B2446246D6E89FCD909E8FE8')")
	assert.NoError(t, err)
	assert.Equal(t, "0ujzPyRiIAffKhBux4PvQdDqMHY", id)

	// min
	err = pgxscan.Get(ctx, db, &id, "select v2.next_ksuid_v2(to_timestamp(1400000000)::timestamptz(3), '\\x00000000000000000000000000000000')")
	assert.NoError(t, err)
	assert.Regexp(t, MinKsuid, id)

	// max
	err = pgxscan.Get(ctx, db, &id, "select v2.next_ksuid_v2(to_timestamp(cast(x'ffffffff' as bigint)+1400000000)::timestamptz(3), '\\xffffffffffffffffffffffffffffffff')")
	assert.NoError(t, err)
	assert.Regexp(t, MaxKsuid, id)
}

func testSorted(t *testing.T, db *pgxpool.Pool) {
	var id1, id2, id3 string
	ctx := context.Background()
	err := pgxscan.Get(ctx, db, &id2, "select v2.next_ksuid_v2()")
	assert.NoError(t, err)
	err = pgxscan.Get(ctx, db, &id1, "select v2.next_ksuid_v2('now'::timestamptz(3) - interval '1 year')")
	assert.NoError(t, err)
	err = pgxscan.Get(ctx, db, &id3, "select v2.next_ksuid_v2('now'::timestamptz(3) + interval '1 year')")
	assert.NoError(t, err)
	assert.Less(t, id1, id2)
	assert.Less(t, id2, id3)
}

func testEdge(t *testing.T, db *pgxpool.Pool) {
	var id1, id2 string
	ctx := context.Background()
	err := pgxscan.Get(ctx, db, &id1, "select v2.next_ksuid_v2(to_timestamp(1400000000)::timestamptz(3))")
	assert.NoError(t, err)
	err = pgxscan.Get(ctx, db, &id2, "select v2.next_ksuid_v2()")
	assert.NoError(t, err)
	assert.Regexp(t, "^[a-zA-Z0-9]{27}$", id1)
	assert.Less(t, id1, id2)
}

func testUnique(t *testing.T, db *pgxpool.Pool) {
	var ids []string
	ctx := context.Background()
	err := pgxscan.Select(ctx, db, &ids, "select v2.next_ksuid_v2() from generate_series(0,10000)")
	assert.NoError(t, err)

	sort.Strings(ids)
	for idx, id := range ids {
		assert.Regexp(t, "^[a-zA-Z0-9]{27}$", id)
		if idx > 0 {
			assert.NotEqual(t, id, ids[idx-1])
		}
	}
}
