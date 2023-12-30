package queue

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

func RunProcessor(t *testing.T, pool *pgxpool.Pool, proc Processor) {
	RunProcessorWithObserver(t, pool, proc, func(context.Context, string, string) {})
}

func RunProcessorWithObserver(t *testing.T, pool *pgxpool.Pool, proc Processor, notify NotifyFn) {
	q := Queue{
		pool:               pool,
		workerCount:        0,
		workerPollInterval: time.Second,
		lockTimeout:        time.Hour,
		processors:         &map[string]Processor{proc.Type(): proc},
		instanceName:       "test",
		observer:           notify,
		types:              []string{proc.Type()},
	}
	ctx := context.Background()

	for {
		progress, err := q.dequeue(ctx)
		assert.NoError(t, err)
		if !progress {
			break
		}
	}
}
