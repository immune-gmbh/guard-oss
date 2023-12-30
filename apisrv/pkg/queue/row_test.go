package queue

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func TestJob(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"InsertLockUnlock":   testInsertLockUnlock,
		"UnlockDiffInstance": testUnlockDiffInstance,
		"AllLocked":          testAllLocked,
		"PastDeadline":       testPastDeadline,
		"UnlockExpired":      testUnlockExpired,
		"ByReference":        testByRef,
		"DupReference":       testDupRef,
		"UnlockUnlocked":     testUnlockUnlocked,
		"GarbageCollect":     testGC,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.Reset(t, ctx, database.MigrationFiles)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testInsertLockUnlock(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	types := []string{"test/v1", "blub/v1"}

	for i := 0; i < 50; i += 1 {
		row, err := Enqueue(ctx, pool, "test/v1", fmt.Sprintf("testref-%d", i), nil, now, now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}

	for i := 0; i < 50; i += 1 {
		row, err := Enqueue(ctx, pool, "blub/v1", fmt.Sprintf("blubref-%d", i), nil, now, now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}

	wg := new(sync.WaitGroup)
	wg.Add(10)
	cnt := int64(0)

	for i := 0; i < 10; i += 1 {
		id := fmt.Sprintf("blah-%d", i)
		go func() {
			defer wg.Done()
			for {
				now := time.Now()
				row, err := lockJob(ctx, pool, id, types, now.Add(time.Minute), now)
				if err == database.ErrNotFound {
					return
				}
				assert.NoError(t, err)
				if err != nil {
					return
				}

				retval := rand.Intn(1) == 0

				if retval {
					row.Successful = &retval
					row.FinishedAt = &now

					err = unlockJob(ctx, pool, row)
					assert.NoError(t, err)
					if err != nil {
						return
					}
					fmt.Printf("%d jobs done\n", atomic.AddInt64(&cnt, 1))
				} else {
					err = unlockJob(ctx, pool, row)
					assert.NoError(t, err)
					if err != nil {
						return
					}
				}
			}
		}()
	}

	wg.Wait()

	now = time.Now()
	for i := 0; i < 50; i += 1 {
		row, err := ByReference(ctx, pool, "test/v1", fmt.Sprintf("testref-%d", i))
		assert.NoError(t, err)
		assert.NotNil(t, row)
		assert.NotNil(t, row.Successful)
		assert.NotNil(t, row.FinishedAt)
		assert.Less(t, *row.FinishedAt, now)
	}
	for i := 0; i < 50; i += 1 {
		row, err := ByReference(ctx, pool, "blub/v1", fmt.Sprintf("blubref-%d", i))
		assert.NoError(t, err)
		assert.NotNil(t, row)
		assert.NotNil(t, row.Successful)
		assert.NotNil(t, row.FinishedAt)
		assert.Less(t, *row.FinishedAt, now)
	}
}

func testAllLocked(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	types := []string{"test/v1"}

	// empty pool
	row, err := lockJob(ctx, pool, "test", types, now.Add(time.Minute), now)
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, row)

	// put 50 jobs in the pool and lock them all
	for i := 0; i < 50; i += 1 {
		row, err := Enqueue(ctx, pool, "test/v1", fmt.Sprintf("testref-%d", i), nil, now, now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}
	for i := 0; i < 50; i += 1 {
		_, err = lockJob(ctx, pool, "test", types, now.Add(time.Minute), now)
		assert.NoError(t, err)
	}

	// all locked
	row, err = lockJob(ctx, pool, "test", types, now.Add(time.Minute), now)
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, row)
}

func testUnlockDiffInstance(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	types := []string{"test/v1"}

	// lock job
	_, err := Enqueue(ctx, pool, "test/v1", "testref", nil, now, now)
	assert.NoError(t, err)
	row, err := lockJob(ctx, pool, "test", types, now.Add(time.Minute), now)
	assert.NoError(t, err)

	// unlock with wrong instance name
	instanceName := "blub"
	row.LockedBy = &instanceName
	err = unlockJob(ctx, pool, row)
	assert.Equal(t, database.ErrNotFound, err)
}

func testUnlockUnlocked(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()

	// unlock unlocked job
	row, err := Enqueue(ctx, pool, "test/v1", "testref", nil, now, now)
	assert.NoError(t, err)

	// unlock unlocked job
	err = unlockJob(ctx, pool, row)
	assert.NoError(t, err)
}

func testPastDeadline(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	types := []string{"test/v1"}

	// lock job with deadline in the past
	_, err := Enqueue(ctx, pool, "test/v1", "testref", nil, now, now)
	assert.NoError(t, err)
	_, err = lockJob(ctx, pool, "test", types, now.Add(-1*time.Minute), now)
	assert.Error(t, err)
}

func testByRef(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 50; i += 1 {
		row, err := Enqueue(ctx, pool, "test/v1", fmt.Sprintf("testref-%d", i), nil, now, now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}

	// fetch job by ref
	id := fmt.Sprintf("testref-%d", rand.Intn(49))
	row, err := ByReference(ctx, pool, "test/v1", id)
	assert.NoError(t, err)
	assert.Equal(t, id, row.Reference)

	// fetch job by non existent ref
	row, err = ByReference(ctx, pool, "test/v1", "blah")
	assert.Equal(t, database.ErrNotFound, err)
}

func testDupRef(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()

	_, err := Enqueue(ctx, pool, "test/v1", "blah", nil, now, now)
	assert.NoError(t, err)
	row, err := Enqueue(ctx, pool, "test/v1", "blah", nil, now, now)
	assert.Error(t, err)
	assert.Nil(t, row)
}

func testUnlockExpired(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().Add(-1 * time.Hour)
	types := []string{"test/v1"}

	for i := 0; i < 50; i += 1 {
		row, err := Enqueue(ctx, pool, "test/v1", fmt.Sprintf("testref-%d", i), nil, now, now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}

	num, err := unlockExpired(ctx, pool, now)
	assert.NoError(t, err)
	assert.Equal(t, 0, num)

	for i := 0; i < 50; i += 1 {
		row, err := lockJob(ctx, pool, "test", types, now.Add(time.Duration(rand.Intn(30)+1)*time.Minute), now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}

	now = time.Now()
	num, err = unlockExpired(ctx, pool, now)
	assert.NoError(t, err)
	assert.Equal(t, 50, num)
	for i := 0; i < 50; i += 1 {
		row, err := lockJob(ctx, pool, "test", types, now.Add(2*time.Minute), now)
		assert.NoError(t, err)
		assert.NotNil(t, row)
	}
}
