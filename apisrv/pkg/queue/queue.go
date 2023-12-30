// Background job queue.
package queue

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	defaultWorkerNum           = 3
	defaultWorkerPollInterval  = time.Second
	defaultLockTimeout         = 5 * time.Minute
	defaultMetricsPollInterval = 10 * time.Second
)

var (
	ErrUnknownOption  = errors.New("unknown queue option")
	ErrInvalidOption  = errors.New("invalid queue option")
	ErrUnknownJobType = errors.New("unknown job type")
	ErrNoArguments    = errors.New("no arguments")

	// metrics
	sizeGauge   *prometheus.GaugeVec
	execCounter *prometheus.CounterVec
	execRuntime *prometheus.HistogramVec
)

func init() {
	sizeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_size",
		Help: "Number of items in the queue partitioned by state and type.",
	}, []string{"type", "state"})
	execCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_execution_counter",
		Help: "Number of jobs executed partitioned by type and whether they were successful.",
	}, []string{"type", "failed"})
	execRuntime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "queue_execution_runtime",
		Help: "Running time of jobs in seconds.",
	}, []string{"type", "failed"})

	prometheus.DefaultRegisterer.MustRegister(sizeGauge)
	prometheus.DefaultRegisterer.MustRegister(execCounter)
	prometheus.DefaultRegisterer.MustRegister(execRuntime)
}

type Processor interface {
	Type() string
	Run(ctx context.Context, job *Job)
}

type NotifyFn = func(ctx context.Context, ty string, ref string)

type Queue struct {
	processors         *map[string]Processor
	workerCount        int
	workerPollInterval time.Duration
	pool               *pgxpool.Pool
	lockTimeout        time.Duration
	observer           NotifyFn
	instanceName       string
	types              []string
}

type WithProcessor struct {
	Processor Processor
}

type WithWorkerPool struct {
	NumWorkers   int
	PollInterval time.Duration
}

type WithTimeout struct {
	LockTimeout time.Duration
}

type WithObserver struct {
	Fn NotifyFn
}

type WithInstanceName struct {
	Name string
}

func New(ctx context.Context, pool *pgxpool.Pool, opts ...interface{}) (*Queue, error) {
	ps := make(map[string]Processor)
	tys := make([]string, 0)
	num := defaultWorkerNum
	poll := defaultWorkerPollInterval
	timeout := defaultLockTimeout
	notify := func(context.Context, string, string) {}
	name := ""

	for _, o := range opts {
		switch opt := o.(type) {
		case WithProcessor:
			p := opt.Processor
			if opt.Processor == nil {
				continue
			}
			ty := p.Type()
			if _, ok := ps[ty]; ok {
				tel.Log(ctx).WithField("type", ty).Error("dup processor type")
				return nil, ErrInvalidOption
			}
			ps[ty] = p
			tys = append(tys, ty)

		case WithTimeout:
			if opt.LockTimeout == 0 {
				tel.Log(ctx).Error("lock timeout cannot be zero")
				return nil, ErrInvalidOption
			}
			timeout = opt.LockTimeout

		case WithObserver:
			notify = opt.Fn

		case WithWorkerPool:
			if opt.NumWorkers == 0 {
				tel.Log(ctx).Error("worker pool size cannot be zero")
				return nil, ErrInvalidOption
			}
			num = opt.NumWorkers
			if opt.PollInterval < 500*time.Millisecond {
				tel.Log(ctx).Error("poll interval must be at least 500ms")
				return nil, ErrInvalidOption
			}
			poll = opt.PollInterval

		case WithInstanceName:
			if opt.Name == "" {
				tel.Log(ctx).Error("instance name cannot be empty")
				return nil, ErrInvalidOption
			}
			name = opt.Name

		default:
			tel.Log(ctx).WithField("opt", opt).Error("unknown queue option")
			return nil, ErrUnknownOption
		}
	}

	// set default instance name
	if name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			tel.Log(ctx).WithError(err).Error("get hostname")
			name = os.Getenv("HOSTNAME")
			if name == "" {
				name = "unknown"
			}
		} else {
			name = hostname
		}
	}

	q := &Queue{
		pool:               pool,
		workerCount:        num,
		workerPollInterval: poll,
		lockTimeout:        timeout,
		processors:         &ps,
		observer:           notify,
		instanceName:       name,
		types:              tys,
	}

	return q, nil
}

func (q *Queue) Start(ctx context.Context) error {
	wg := new(sync.WaitGroup)
	wg.Add(q.workerCount + 1)

	// queue workers
	for wn := 0; wn < q.workerCount; wn += 1 {
		go func() {
			ticker := time.Tick(time.Minute)
			for exit := false; !exit; {
				progress, err := q.dequeue(ctx)
				if err != nil || !progress {
					select {
					case <-ticker:
						// open expired locks
						unlockExpired(ctx, q.pool, time.Now())

					case <-time.After(q.workerPollInterval):
						// wait until it's time to poll again

					case <-ctx.Done():
						// shutdown signal
						exit = true
					}
				} else {
					select {
					case _, open := <-ctx.Done():
						exit = !open
					default:
					}
				}
			}
			wg.Done()
		}()
	}

	// metrics exporter
	go func() {
		for exit := false; !exit; {
			exportMetrics(ctx, q.pool, sizeGauge)
			select {
			case <-time.After(defaultMetricsPollInterval):
				// nop
			case <-ctx.Done():
				exit = true
			}
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func (q *Queue) dequeue(ctx context.Context) (bool, error) {
	// fetch job
	now := time.Now()
	timeout := now.Add(q.lockTimeout)
	row, err := lockJob(ctx, q.pool, q.instanceName, q.types, timeout, now)
	if err == database.ErrNotFound {
		return false, nil
	}
	if err != nil {
		tel.Log(ctx).WithError(err).Error("lock job")
		return false, err
	}

	// link to queueing span
	var spanctx trace.SpanContext
	if !row.ScheduledCtx.IsNull() {
		err = row.ScheduledCtx.Decode(&spanctx)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("serialize report")
		}
	}
	link := trace.Link{
		SpanContext: spanctx,
		Attributes: []attribute.KeyValue{
			attribute.String("link.relationship", "origin"),
		},
	}
	// run job
	ctx, span := tel.Start(ctx, "Run Job", tel.WithOption{Option: trace.WithLinks(link)})
	defer span.End()

	span.SetAttributes(
		attribute.String("job.id", row.Id),
		attribute.String("job.type", row.Type))

	runtime := math.NaN()
	job := Job{
		Set: false,
		Row: row,
		Ctx: ctx,
	}

	// call Run()
	p, ok := (*q.processors)[row.Type]
	if !ok {
		tel.Log(ctx).WithField("job.type", row.Type).Error("unknown job type")
	} else {
		begin := time.Now()
		p.Run(ctx, &job)
		runtime = float64(time.Since(begin)) / float64(time.Second)

		// done is default
		if !job.Set {
			job.Done()
		}
	}

	// update row
	err = unlockJob(ctx, q.pool, row)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("finish job")
		return false, nil
	} else if q.observer != nil {
		q.observer(ctx, row.Type, row.Reference)
	}

	execCounter.WithLabelValues(row.Type, fmt.Sprintf("%t", row.FinishedAt != nil)).Add(1)
	if !math.IsNaN(runtime) {
		execRuntime.WithLabelValues(row.Type, fmt.Sprintf("%t", row.FinishedAt != nil)).Observe(runtime)
	}
	return true, nil
}

func exportMetrics(ctx context.Context, pool *pgxpool.Pool, gauge *prometheus.GaugeVec) {
	ctx, span := tel.Start(ctx, "Queue Metrics")
	defer span.End()

	var rows []struct {
		Type       string `db:"type"`
		Successful *bool  `db:"successful"`
		Locked     bool   `db:"locked"`
		JobCount   int64  `db:"job_count"`
	}

	err := pgxscan.Select(ctx, pool, &rows, `
			--
			-- Queue-Metrics
			--
			SELECT
				type,
				successful,
				locked_by is not null as "locked",
				count(*)              as "job_count"
			FROM v2.jobs
			GROUP BY (type, successful, locked)
		`)
	if err == nil {
		for _, row := range rows {
			var state string
			if row.Successful == nil {
				if row.Locked {
					state = "in_progress"
				} else {
					state = "waiting"
				}
			} else {
				if *row.Successful {
					state = "done"
				} else {
					state = "failed"
				}
			}
			gauge.WithLabelValues(row.Type, state).Set(float64(row.JobCount))
		}
	} else {
		tel.Log(ctx).WithError(err).Error("query metrics")
	}
}
