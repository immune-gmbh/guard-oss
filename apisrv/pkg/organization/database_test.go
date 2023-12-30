package organization

import (
	"context"
	_ "embed"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

//go:embed seed.sql
var seedSql string

func TestDatabase(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"Fetch":       testFetch,
		"GetExternal": testGetExternalById,
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

func testFetch(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 17}
	ctx := context.Background()

	cur, max, err := Quota(ctx, conn, int64(100))
	assert.NoError(t, err)
	assert.Equal(t, 2, cur)
	assert.Equal(t, 1, max)

	cur, max, err = Quota(ctx, conn, int64(101))
	assert.NoError(t, err)
	assert.Equal(t, 2, cur)
	assert.Equal(t, 2, max)
}

func testGetExternalById(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 17}
	ctx := context.Background()

	ext, err := GetExternalById(ctx, conn, 100)
	assert.NoError(t, err)
	assert.Equal(t, "ext-1", *ext)
}
