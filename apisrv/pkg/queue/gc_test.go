package queue

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

func testGC(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// insert unfinished job
	now := time.Now()
	job1, err := Enqueue(ctx, pool, "test/v1", "testref1", nil, now, now)
	assert.NoError(t, err)

	// insert finished now job
	types := []string{"test/v2"}
	now = time.Now()
	job2, err := Enqueue(ctx, pool, "test/v2", "testref2", nil, now, now)
	assert.NoError(t, err)
	row, err := lockJob(ctx, pool, "test", types, now.Add(time.Duration(rand.Intn(30)+1)*time.Minute), now)
	assert.NoError(t, err)
	assert.NotNil(t, row)
	assert.Equal(t, job2.Id, row.Id)
	instanceName := "test"
	row.LockedBy = &instanceName
	now = now.Add(1 * time.Minute)
	row.FinishedAt = &now
	succs := true
	row.Successful = &succs
	err = unlockJob(ctx, pool, row)
	assert.NoError(t, err)

	// insert finished old job
	earlier := now.AddDate(-1, 0, 0)
	job3, err := Enqueue(ctx, pool, "test/v2", "testref3", nil, earlier, earlier)
	assert.NoError(t, err)
	row, err = lockJob(ctx, pool, "test", types, now.Add(time.Duration(rand.Intn(30)+1)*time.Minute), now)
	assert.NoError(t, err)
	assert.NotNil(t, row)
	assert.Equal(t, job3.Id, row.Id)
	row.LockedBy = &instanceName
	earlier = earlier.Add(1 * time.Minute)
	row.FinishedAt = &earlier
	succs = false
	row.Successful = &succs
	err = unlockJob(ctx, pool, row)
	assert.NoError(t, err)

	// insert unfinishedold job
	earlier = now.AddDate(-2, 0, 0)
	job4, err := Enqueue(ctx, pool, "test/v1", "testref4", nil, now, now)
	assert.NoError(t, err)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	assert.NoError(t, err)
	err = GarbageCollect(ctx, tx)
	assert.NoError(t, err)
	tx.Commit(ctx)

	// check that only the old change was deleted
	var rows []Row
	err = pgxscan.Select(ctx, pool, &rows, "select * from v2.jobs order by scheduled_at asc")
	assert.NoError(t, err)
	assert.Len(t, rows, 3)
	assert.Equal(t, job1.Id, rows[0].Id)
	assert.Equal(t, job2.Id, rows[1].Id)
	assert.Equal(t, job4.Id, rows[2].Id)
}
