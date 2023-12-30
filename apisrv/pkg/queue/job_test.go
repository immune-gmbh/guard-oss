package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExponential(t *testing.T) {
	exp := Exponetial{
		Min: 5 * time.Second,
		Max: 30 * time.Minute,
	}

	assert.Equal(t, 5*time.Second, exp.Value(0))
	assert.Equal(t, 5*time.Second*2, exp.Value(1))
	assert.Equal(t, 5*time.Second*4, exp.Value(2))
	assert.Equal(t, 5*time.Second*8, exp.Value(3))
	assert.Equal(t, 5*time.Second*16, exp.Value(4))
	assert.Equal(t, 5*time.Second*32, exp.Value(5))
	assert.Equal(t, 5*time.Second*64, exp.Value(6))
	assert.Equal(t, 5*time.Second*128, exp.Value(7))
	assert.Equal(t, 5*time.Second*256, exp.Value(8))
	assert.Equal(t, 30*time.Minute, exp.Value(9))
	assert.Equal(t, 30*time.Minute, exp.Value(10))
	assert.Equal(t, 30*time.Minute, exp.Value(11))
	assert.Equal(t, 30*time.Minute, exp.Value(0xffffffff))
}

func TestDone(t *testing.T) {
	job := Job{
		Ctx: context.TODO(),
		Row: new(Row),
	}

	job.Done()
	assert.True(t, job.Set)
	assert.True(t, *job.Row.Successful)
}

func TestRetryInt(t *testing.T) {
	job := Job{
		Ctx: context.TODO(),
		Row: new(Row),
	}

	job.Retry(5)
	assert.True(t, job.Set)
	assert.Equal(t, 1, job.Row.ErrorCount)
	assert.Nil(t, job.Row.Successful)
	assert.WithinDuration(t, time.Now().Add(time.Second*5), job.Row.NextRunAt, time.Millisecond)
}

func TestRetryTime(t *testing.T) {
	job := Job{
		Ctx: context.TODO(),
		Row: new(Row),
	}

	job.Retry(time.Now().Add(time.Minute))
	assert.True(t, job.Set)
	assert.Equal(t, 1, job.Row.ErrorCount)
	assert.Nil(t, job.Row.Successful)
	assert.WithinDuration(t, time.Now().Add(time.Minute), job.Row.NextRunAt, time.Millisecond)
}

func TestRetryDuration(t *testing.T) {
	job := Job{
		Ctx: context.TODO(),
		Row: new(Row),
	}

	job.Retry(time.Minute)
	assert.True(t, job.Set)
	assert.Equal(t, 1, job.Row.ErrorCount)
	assert.Nil(t, job.Row.Successful)
	assert.WithinDuration(t, time.Now().Add(time.Minute), job.Row.NextRunAt, time.Millisecond)
}

func TestRetryExponential(t *testing.T) {
	job := Job{
		Ctx: context.TODO(),
		Row: new(Row),
	}
	exp := Exponetial{
		Min: 5 * time.Second,
		Max: 30 * time.Minute,
	}

	job.Retry(exp)
	assert.True(t, job.Set)
	assert.Equal(t, 1, job.Row.ErrorCount)
	assert.Nil(t, job.Row.Successful)
	assert.WithinDuration(t, time.Now().Add(time.Second*5), job.Row.NextRunAt, time.Millisecond)
}
