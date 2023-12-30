package queue

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func TestQueue(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"QueueDequeue": testQueueDequeue,
		"FailingJobs":  testFailingJobs,
		"Config":       testConfig,
		"Observer":     testObserver,
		"AreReady":     testAreReady,
		"ExpiredJob":   testExpiredJob,
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

var (
	testErr  = errors.New("transient test error")
	fatalErr = errors.New("fatal test error")
)

type testProcessor struct {
	Ty     string
	Worker func(context.Context, *Job)
}

func (p *testProcessor) Type() string {
	return p.Ty
}
func (p *testProcessor) Run(ctx context.Context, job *Job) {
	p.Worker(ctx, job)
}

func testQueueDequeue(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithCancel(context.Background())
	worker := func(ctx context.Context, job *Job) {
		var val []byte
		fmt.Println(job.Row.Args)
		err := job.Arguments(&val)
		assert.NoError(t, err)
		assert.Equal(t, val, []byte{1, 2, 3, 4})

		var spanctx trace.SpanContext
		assert.False(t, job.Row.ScheduledCtx.IsNull())
		assert.NoError(t, job.Row.ScheduledCtx.Decode(&spanctx))

		cancel()
	}
	queue, err := New(ctx, pool,
		WithProcessor{Processor: &testProcessor{
			Ty:     "test/v1",
			Worker: worker,
		}})
	assert.NoError(t, err)
	assert.NotNil(t, queue)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := queue.Start(ctx)
		assert.NoError(t, err)
	}()

	now := time.Now()
	_, err = Enqueue(ctx, pool, "test/v1", "blah", []byte{1, 2, 3, 4}, now, now)
	assert.NoError(t, err)

	wg.Wait()
}

func testExpiredJob(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithCancel(context.Background())
	worker := func(ctx context.Context, job *Job) {
		var val []byte
		err := job.Arguments(&val)
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3, 4}, val)
		cancel()
	}
	queue, err := New(ctx, pool,
		WithProcessor{Processor: &testProcessor{
			Ty:     "test/v1",
			Worker: worker,
		}})
	assert.NoError(t, err)
	assert.NotNil(t, queue)

	now := time.Now().Add(-10 * time.Minute)
	_, err = Enqueue(ctx, pool, "test/v1", "blah", []byte{1, 2, 3, 4}, now, now)
	assert.NoError(t, err)
	timeout := now.Add(5 * time.Minute)

	row, err := lockJob(ctx, pool, "test-instance", []string{"test/v1"}, timeout, now)
	assert.NotNil(t, row)
	assert.NoError(t, err)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := queue.Start(ctx)
		assert.NoError(t, err)
	}()

	wg.Wait()
}

func testFailingJobs(t *testing.T, pool *pgxpool.Pool) {
	type jobArgs struct {
		Blah string
		Blub int
	}

	finished := int64(0)
	successful := int64(0)
	ctx, cancel := context.WithCancel(context.Background())
	worker := func(ctx context.Context, job *Job) {
		var a jobArgs
		err := job.Arguments(&a)
		assert.NoError(t, err)
		switch rand.Intn(3) {
		case 0:
			atomic.AddInt64(&finished, 1)
			atomic.AddInt64(&successful, 1)
		case 1:
			job.Retry(time.Second)
		default:
			atomic.AddInt64(&finished, 1)
			job.Failed()
		}
	}
	queue, err := New(ctx, pool,
		WithProcessor{Processor: &testProcessor{
			Ty:     "test/v1",
			Worker: worker,
		}})
	assert.NoError(t, err)
	assert.NotNil(t, queue)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := queue.Start(ctx)
		assert.NoError(t, err)
	}()

	now := time.Now()
	for i := 0; i < 50; i += 1 {
		args := jobArgs{
			Blub: i,
			Blah: fmt.Sprintf("%d", i),
		}
		_, err = Enqueue(ctx, pool, "test/v1", fmt.Sprintf("blah-%d", i), args, now, now)
		assert.NoError(t, err)
	}

	for finished < 50 {
		fmt.Printf("%d/50 jobs done\n", finished)
		time.Sleep(time.Millisecond * 300)
	}
	fmt.Printf("%d/50 jobs done\n", finished)
	cancel()
	wg.Wait()

	ctx = context.Background()
	assert.Greater(t, successful, int64(0))

	for i := 0; i < 50; i += 1 {
		row, err := ByReference(ctx, pool, "test/v1", fmt.Sprintf("blah-%d", i))
		assert.NoError(t, err)
		assert.NotNil(t, row.Successful)
		if *row.Successful {
			successful -= 1
		}
	}

	assert.Equal(t, int64(0), successful)
}

func testConfig(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	queue, err := New(ctx, pool, WithInstanceName{Name: ""})
	assert.Error(t, err)
	assert.Nil(t, queue)

	_, err = New(ctx, pool, WithInstanceName{Name: "a"})
	assert.NoError(t, err)

	queue, err = New(ctx, pool, WithWorkerPool{NumWorkers: 0, PollInterval: 0})
	assert.Error(t, err)
	assert.Nil(t, queue)

	queue, err = New(ctx, pool, WithWorkerPool{NumWorkers: 1, PollInterval: 0})
	assert.Error(t, err)
	assert.Nil(t, queue)

	_, err = New(ctx, pool, WithWorkerPool{NumWorkers: 1, PollInterval: time.Second})
	assert.NoError(t, err)

	queue, err = New(ctx, pool, WithTimeout{LockTimeout: 0})
	assert.Error(t, err)
	assert.Nil(t, queue)

	_, err = New(ctx, pool, WithTimeout{LockTimeout: time.Second})
	assert.NoError(t, err)

	queue, err = New(ctx, pool, 42)
	assert.Error(t, err)
	assert.Nil(t, queue)

	queue, err = New(ctx, pool,
		WithProcessor{Processor: &testProcessor{Ty: "a"}},
		WithProcessor{Processor: &testProcessor{Ty: "a"}})
	assert.Error(t, err)
	assert.Nil(t, queue)
}

func testObserver(t *testing.T, pool *pgxpool.Pool) {
	observed := int64(0)
	ctx, cancel := context.WithCancel(context.Background())
	worker := func(ctx context.Context, job *Job) {
		var a bool
		err := job.Arguments(&a)
		assert.NoError(t, err)
		if !a {
			job.Failed()
		}
	}
	observer := func(ctx context.Context, ty string, ref string) {
		assert.Equal(t, "test/v1", ty)
		assert.Regexp(t, "blah(1|2)", ref)
		atomic.AddInt64(&observed, 1)
	}
	queue, err := New(ctx, pool,
		WithProcessor{Processor: &testProcessor{
			Ty:     "test/v1",
			Worker: worker,
		}},
		WithObserver{Fn: observer})
	assert.NoError(t, err)
	assert.NotNil(t, queue)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := queue.Start(ctx)
		assert.NoError(t, err)
	}()

	now := time.Now()
	_, err = Enqueue(ctx, pool, "test/v1", "blah1", false, now, now)
	assert.NoError(t, err)
	_, err = Enqueue(ctx, pool, "test/v1", "blah2", true, now, now)
	assert.NoError(t, err)

	for observed != 2 {
		time.Sleep(time.Millisecond * 300)
	}
	cancel()
	wg.Wait()
}

func testAreReady(t *testing.T, pool *pgxpool.Pool) {
	var (
		stop1 int64
		stop2 int64
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker := func(ctx context.Context, job *Job) {
		var id int64
		err := job.Arguments(&id)
		assert.NoError(t, err)
		if id == 1 && stop1 > 0 {
			job.Done()
		} else if id == 2 && stop2 > 0 {
			job.Done()
		} else {
			job.Retry()
		}
	}
	queue, err := New(ctx, pool,
		WithProcessor{Processor: &testProcessor{
			Ty:     "test/v1",
			Worker: worker,
		}})
	assert.NoError(t, err)
	assert.NotNil(t, queue)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := queue.Start(ctx)
		assert.NoError(t, err)
	}()

	now := time.Now()
	maxAge := now.Add(-30 * time.Second)
	_, err = Enqueue(ctx, pool, "test/v1", "blah1", 1, now, now)
	assert.NoError(t, err)
	_, err = Enqueue(ctx, pool, "test/v1", "blah2", 2, now, now)
	assert.NoError(t, err)

	ready, err := AreReady(ctx, pool, []string{}, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	ready, err = AreReady(ctx, pool, nil, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	ready, err = AreReady(ctx, pool, []string{"blub"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	time.Sleep(time.Second)

	ready, err = AreReady(ctx, pool, []string{"blah1"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	ready, err = AreReady(ctx, pool, []string{"blah2"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	ready, err = AreReady(ctx, pool, []string{"blah1", "blah2"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	ready, err = AreReady(ctx, pool, []string{"blah1", "blah2", "blub"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	// finish 1st job
	atomic.AddInt64(&stop1, 1)

	for {
		ready, err := AreReady(ctx, pool, []string{"blah1"}, maxAge)
		assert.NoError(t, err)
		if ready {
			break
		}
		time.Sleep(time.Second)
	}

	ready, err = AreReady(ctx, pool, []string{"blah2"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	ready, err = AreReady(ctx, pool, []string{"blah1", "blah2"}, maxAge)
	assert.NoError(t, err)
	assert.False(t, ready)

	// finish 2nd job
	atomic.AddInt64(&stop2, 1)

	for {
		ready, err = AreReady(ctx, pool, []string{"blah2"}, maxAge)
		assert.NoError(t, err)
		if ready {
			break
		}
		time.Sleep(time.Second)
	}

	ready, err = AreReady(ctx, pool, []string{"blah1", "blah2"}, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	ready, err = AreReady(ctx, pool, []string{"blah1", "blah2", "blub"}, maxAge)
	assert.NoError(t, err)
	assert.True(t, ready)

	cancel()
	wg.Wait()
}
