package tag

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

//go:embed seed.sql
var seedSql string

func TestIdQuery(t *testing.T) {
	mock := database.MockQuerier{
		Responses: map[string]database.MockResult{
			`(\s|-)+Tag fetch(\s|-)+`: {
				Columns: []string{"id"},
				Rows: [][]interface{}{
					{"abcd"},
				},
			},
		},
	}
	ctx := context.Background()

	rows, err := Fetch(ctx, Point("1234", 100)).Columns("id").Do(mock)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestColumns(t *testing.T) {
	mock := database.MockQuerier{
		Responses: map[string]database.MockResult{
			strings.Join(defaultColumns, ","): {
				Columns: []string{"id"},
				Rows:    [][]interface{}{},
			},
		},
	}
	ctx := context.Background()

	_, err := Fetch(ctx, Point("1234", 100)).Columns().Do(mock)
	assert.Error(t, err)
}

func TestFetch(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"Single":   testSingle,
		"Set":      testSet,
		"Relation": testRelation,
		"Range":    testRange,
		"Text":     testText,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 33)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testSingle(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 9}
	ctx := context.Background()
	row, err := Fetch(ctx, Point("0ujtsYcgvSTl8PAuAdqWYSMnLOv", 100)).DoSingle(conn)
	assert.NoError(t, err)
	assert.Equal(t, "Test tag #1", row.Key)
	assert.Nil(t, row.Devices)
	md, err := row.GetMetadata()
	assert.NoError(t, err)
	assert.Equal(t, NewMetadata(), *md)

	row, err = Fetch(ctx, Point("0ujtsYcgvSTl8PAuAdqWYSMnLOv", 100)).Columns("id", "key").DoSingle(conn)
	assert.NoError(t, err)
	assert.Equal(t, "Test tag #1", row.Key)
	assert.True(t, row.Metadata.IsNull())

	_, err = Fetch(ctx, Point("1ujtsYcgvSTl8PAuAdqWYSMnLOv", 100)).DoSingle(conn)
	assert.Error(t, err)

	_, err = Fetch(ctx, Point("0ujtsYcgvSTl8PAuAdqWYSMnLOv", 101)).DoSingle(conn)
	assert.Error(t, err)
}

func testRelation(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 42}
	ctx := context.Background()
	var org int64 = 100

	row, err := Fetch(ctx, Set(&org, "0ujtsYcgvSTl8PAuAdqWYSMnLOv")).
		Relations("devices").DoSingle(conn)
	assert.NoError(t, err)
	assert.Equal(t, []int64{100}, row.Devices)

	rows, err := Fetch(ctx, Set(nil, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", "0ujtsYcgvSTl8PAuAdqWYSMnLOw")).
		Relations("devices").Do(conn)
	assert.NoError(t, err)
	assert.Equal(t, []int64{100, 101}, rows[0].Devices)
	assert.Equal(t, []int64{100}, rows[1].Devices)

	rows, err = Fetch(ctx, Set(nil, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", "0ujtsYcgvSTl8PAuAdqWYSMnLOw")).
		Relations("blah").Do(conn)
	assert.NoError(t, err)
	assert.Nil(t, rows[0].Devices)
	assert.Nil(t, rows[1].Devices)
}

func testSet(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 14}
	ctx := context.Background()
	var org int64 = 100

	rows, err := Fetch(ctx, Set(&org, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", "0ujtsYcgvSTl8PAuAdqWYSMnLOw")).Do(conn)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Test tag #2", rows[0].Key)
	assert.Equal(t, "Test tag #1", rows[1].Key)

	rows2, err := Fetch(ctx, Set(nil, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", "0ujtsYcgvSTl8PAuAdqWYSMnLOw")).Do(conn)
	assert.NoError(t, err)
	assert.Equal(t, rows, rows2)

	rows2, err = Fetch(ctx, Set(nil, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", "0ujtsYcgvSTl8PAuAdqWYSMnLOw", "0ujtsYcgvSTl8PAuAdqWYSMnLOw", "0ujtsYcgvSTl8PAuAdqWYSMnLOw")).Do(conn)
	assert.NoError(t, err)
	assert.Equal(t, rows, rows2)

	rows, err = Fetch(ctx, Set(nil, "0ujtsYcgvSTl8PAuAdqWYSMnLOv", "1ujtsYcgvSTl8PAuAdqWYSMnLOv")).Do(conn)
	fmt.Println(rows)
	assert.Error(t, err)

	rows, err = Fetch(ctx, Set(nil)).Do(conn)
	assert.NoError(t, err)
	assert.Empty(t, rows)

	rows, err = Fetch(ctx, Set(&org)).Do(conn)
	assert.NoError(t, err)
	assert.Empty(t, rows)
}

func testRange(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 9}
	ctx := context.Background()
	var org int64 = 100

	rows, err := Fetch(ctx, Range(nil, org)).Do(conn)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Test tag #2", rows[0].Key)
	assert.Equal(t, "Test tag #1", rows[1].Key)

	conn = database.ExplainQuerier{Database: db, MaxCost: 22}

	rows, err = Fetch(ctx, Range(nil, org)).Relations("devices").Do(conn)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Test tag #2", rows[0].Key)
	assert.Len(t, rows[0].Devices, 2)
	assert.Equal(t, "Test tag #1", rows[1].Key)
	assert.Len(t, rows[1].Devices, 1)

	rows, err = Fetch(ctx, Range(&rows[0].Id, org)).Relations("devices").Do(conn)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "Test tag #1", rows[0].Key)
	assert.Len(t, rows[0].Devices, 1)

	conn = database.ExplainQuerier{Database: db, MaxCost: 9}
	sqli := "'; truncate v2.devices;"
	_, err = Fetch(ctx, Range(&sqli, org)).Do(conn)
	assert.Error(t, err)

	id := database.MinKsuid
	rows, err = Fetch(ctx, Range(&id, org)).Do(conn)
	assert.NoError(t, err)
	assert.Empty(t, rows)

	id = database.MaxKsuid
	rows, err = Fetch(ctx, Range(&id, org)).Do(conn)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)

	org = 101
	rows, err = Fetch(ctx, Range(nil, org)).Do(conn)
	assert.NoError(t, err)
	assert.Empty(t, rows)
}

func testText(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 10}
	ctx := context.Background()
	var org int64 = 100

	rows, err := Fetch(ctx, Text("Test", org)).Do(conn)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Contains(t, []string{"Test tag #1", "Test tag #2"}, rows[0].Key)
	assert.Contains(t, []string{"Test tag #1", "Test tag #2"}, rows[1].Key)

	rows, err = Fetch(ctx, Text("Test", 101)).Do(conn)
	assert.NoError(t, err)
	assert.Empty(t, rows)

	conn = database.ExplainQuerier{Database: db, MaxCost: 24}
	rows, err = Fetch(ctx, Text("Test", org)).Relations("devices").Do(conn)
	if assert.NoError(t, err) {
		assert.Len(t, rows, 2)
	}
}
