package queue

import (
	"context"
	"time"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type Job struct {
	Set bool
	Row *Row
	Ctx context.Context
}

func (j *Job) Arguments(args interface{}) error {
	if j.Row.Args.IsNull() {
		return ErrNoArguments
	}
	return j.Row.Args.Decode(args)
}

type Exponetial struct {
	Min   time.Duration
	Max   time.Duration
	Steps int
}

func (e *Exponetial) Value(errorCount int) time.Duration {
	if errorCount > 30 {
		errorCount = 30
	}
	backoff := e.Min * (1 << errorCount)
	if backoff > e.Max {
		backoff = e.Max
	}

	return backoff
}

func (j *Job) Done() {
	if j.Set {
		tel.Log(j.Ctx).Error("Done() called on finalized job")
	}
	retval := true
	now := time.Now()
	j.Row.Successful = &retval
	j.Row.FinishedAt = &now
	j.Set = true
}

func (j *Job) Retry(timeout ...interface{}) {
	if j.Set {
		tel.Log(j.Ctx).Error("Retry() called on finalized job")
	}
	for _, t := range timeout {
		switch t := t.(type) {
		case int:
			j.Retry(time.Duration(t) * time.Second)
			return
		case time.Duration:
			j.Row.NextRunAt = time.Now().Add(t)
		case time.Time:
			j.Row.NextRunAt = t
		case Exponetial:
			j.Row.NextRunAt = time.Now().Add(t.Value(j.Row.ErrorCount))
		default:
			tel.Log(j.Ctx).WithField("timeout", t).Error("Retry() called unknown timeout type")
			j.Row.NextRunAt = time.Now().Add(time.Minute)
		}
		break
	}
	j.Row.ErrorCount += 1
	j.Set = true
}

func (j *Job) Failed() {
	if j.Set {
		tel.Log(j.Ctx).Error("Failed() called on finalized job")
	}
	retval := false
	now := time.Now()
	j.Row.Successful = &retval
	j.Row.FinishedAt = &now
	j.Set = true
}
