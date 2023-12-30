package filter

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

//go:embed seed.sql
var seedSql string

func TestFilter(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"Devices": testFilterDevices,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 50)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func rowsToDevIds(rows []device.Row) []int64 {
	ret := make([]int64, len(rows))
	for i := range rows {
		ret[i] = rows[i].Id
	}
	return ret
}

func testFilterDevices(t *testing.T, db *pgxpool.Pool) {
	conn := database.ExplainQuerier{Database: db, MaxCost: 100}
	ctx := context.Background()

	// list devices for org
	orgId := int64(100)
	now := time.Now()
	var tagIds []string
	issueId := ""
	state := ""
	rows, err := ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{103, 102, 100}, rowsToDevIds(rows), "failed list all devices")
	}

	// filter devices by tag
	tagIds = []string{"0ujtsYcgvSTl8PAuAdqWYSMnLOw"}
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{102, 100}, rowsToDevIds(rows), "failed filter one tag")
	}

	// filter devices by two tags anded
	tagIds = append(tagIds, "0ujtsYcgvSTl8PAuAdqWYSMnLOv")
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{100}, rowsToDevIds(rows), "failed filter two tags")
	}

	// filter devices by unknown tag
	tagIds = []string{"42jtsYcgvSTl8PAuAdqWYSMnL23"}
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{}, rowsToDevIds(rows), "failed filter unknown tag")
	}

	// filter devices by issue
	tagIds = nil
	issueId = "issue1"
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{102, 100}, rowsToDevIds(rows), "failed filter issue")
	}

	// filter devices by issue and tag
	tagIds = []string{"0ujtsYcgvSTl8PAuAdqWYSMnLOv"}
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{100}, rowsToDevIds(rows), "failed filter issue and tag")
	}

	// unknown issue
	issueId = "issue3"
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{}, rowsToDevIds(rows), "failed filter unknown issue")
	}

	// filter device by state
	issueId = ""
	tagIds = nil
	state = "trusted"
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{102, 100}, rowsToDevIds(rows), "failed filter state trusted")
	}
	state = "outdated"
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{103}, rowsToDevIds(rows), "failed filter state outdated")
	}
	state = "vulnerable"
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{}, rowsToDevIds(rows), "failed filter state vulnerable")
	}

	// filter device by state and tag
	tagIds = []string{"0ujtsYcgvSTl8PAuAdqWYSMnLOv"}
	state = "outdated"
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{103}, rowsToDevIds(rows), "failed filter state and tag")
	}
	tagIds = []string{"0ujtsYcgvSTl8PAuAdqWYSMnLOw"}
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{}, rowsToDevIds(rows), "failed filter state and tag")
	}

	// filter device by state and issue
	issueId = "issue2"
	tagIds = nil
	rows, err = ListActiveDevicesFiltered(ctx, conn, orgId, 100, nil, now, tagIds, issueId, state)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []int64{103}, rowsToDevIds(rows), "failed filter state and issue")
	}
}
