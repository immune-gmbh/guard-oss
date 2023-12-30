package tag

import (
	"context"
	_ "fmt"
	"testing"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
)

func TestWrite(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"Simple": testSimple,
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

func tagsForDevice(ctx context.Context, qq pgxscan.Querier, dev int64) ([]string, error) {
	now := time.Now()
	devrow, err := device.Get(ctx, qq, dev, int64(100), now)
	if err != nil {
		return nil, err
	}
	tagrows, err := GetTagsByDeviceId(ctx, qq, devrow.Id)
	if err != nil {
		return nil, err
	}
	tags := make([]string, len(tagrows))
	for i, row := range tagrows {
		tags[i] = row.Key
	}
	return tags, nil
}

func testSimple(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	//conn := database.ExplainQuerier{Database: db, MaxCost: 42.68}
	conn := db

	tags, err := tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"Test tag #1", "Test tag #2"}, tags)

	//conn = database.ExplainQuerier{Database: db, MaxCost: 14.77}
	changed, err := Device(ctx, conn, 100, 100, []string{"a", "b", "c"})
	assert.NoError(t, err)
	assert.True(t, changed)
	//conn = database.ExplainQuerier{Database: db, MaxCost: 42.68}
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, tags)

	changed, err = Device(ctx, conn, 100, 100, []string{"a", "b", "c"})
	assert.NoError(t, err)
	assert.False(t, changed)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, tags)

	changed, err = Device(ctx, conn, 100, 100, []string{"b", "c"})
	assert.NoError(t, err)
	assert.True(t, changed)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"b", "c"}, tags)

	changed, err = Device(ctx, conn, 100, 100, []string{"b", "c", "a"})
	assert.NoError(t, err)
	assert.True(t, changed)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"b", "c", "a"}, tags)

	changed, err = Device(ctx, conn, 100, 100, []string{"Test tag #2", "c", "a"})
	assert.NoError(t, err)
	assert.True(t, changed)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"Test tag #2", "c", "a"}, tags)

	changed, err = Device(ctx, conn, 100, 100, nil)
	assert.NoError(t, err)
	assert.True(t, changed)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.Empty(t, tags)

	changed, err = Device(ctx, conn, 102, 100, []string{"b", "c", "a"})
	assert.Error(t, err)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.Empty(t, tags)

	changed, err = Device(ctx, conn, 100, 101, []string{"b", "c", "a"})
	assert.Error(t, err)
	tags, err = tagsForDevice(ctx, conn, 100)
	assert.NoError(t, err)
	assert.Empty(t, tags)
}
