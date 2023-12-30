package inteltsc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestTSC(t *testing.T) {
	t.Skipf("Intel TSC test instance is down")
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)
	pgsqlC.Reset(t, ctx, database.MigrationFiles)
	conn := pgsqlC.Connect(t, ctx)
	defer conn.Close()

	// schedule job
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	now := time.Now()
	data, certs, err := Schedule(ctx, tx, "Test", "PF2B5BEE", now)
	assert.Equal(t, ErrInProgress, err)
	assert.Nil(t, data)
	assert.Nil(t, certs)

	err = tx.Commit(ctx)
	assert.NoError(t, err)

	// fetch pending results
	ref := Reference("Test", "PF2B5BEE")
	data, certs, err = Fetch(ctx, conn, ref)
	assert.Equal(t, ErrInProgress, err)
	assert.Nil(t, data)
	assert.Empty(t, certs)

	// run job
	proc, err := NewProcessor(ctx, conn, []Site{KaisTestCredentials})
	assert.NoError(t, err)
	queue.RunProcessor(t, conn, proc)

	// fetch results
	ref = Reference("Test", "PF2B5BEE")
	data, certs, err = Fetch(ctx, conn, ref)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotEmpty(t, certs)

	// fetch results via Schedule
	tx, err = conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	data, certs, err = Schedule(ctx, tx, "Test", "PF2B5BEE", now)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotEmpty(t, certs)

	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func TestTSCNonExistance(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)
	pgsqlC.Reset(t, ctx, database.MigrationFiles)
	conn := pgsqlC.Connect(t, ctx)
	defer conn.Close()

	// fetch pending results
	ref := Reference("Test", "PF2B5BEE")
	data, certs, err := Fetch(ctx, conn, ref)
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, data)
	assert.Nil(t, certs)
}

func TestTSCNotFound(t *testing.T) {
	t.Skipf("Intel TSC test instance is down")
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)
	pgsqlC.Reset(t, ctx, database.MigrationFiles)
	conn := pgsqlC.Connect(t, ctx)
	defer conn.Close()

	// schedule job
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	now := time.Now()
	data, certs, err := Schedule(ctx, tx, "Test", "PF2B5BFF", now)
	assert.Equal(t, ErrInProgress, err)
	assert.Nil(t, data)
	assert.Nil(t, certs)

	err = tx.Commit(ctx)
	assert.NoError(t, err)

	// fetch pending results
	ref := Reference("Test", "PF2B5BFF")
	data, certs, err = Fetch(ctx, conn, ref)
	assert.Equal(t, ErrInProgress, err)
	assert.Nil(t, data)
	assert.Empty(t, certs)

	// run job
	proc, err := NewProcessor(ctx, conn, []Site{KaisTestCredentials})
	assert.NoError(t, err)
	queue.RunProcessor(t, conn, proc)

	// fetch results
	ref = Reference("Test", "PF2B5BFF")
	data, certs, err = Fetch(ctx, conn, ref)
	assert.NoError(t, err)
	assert.Nil(t, data)
	assert.Empty(t, certs)

	// fetch results via Schedule
	tx, err = conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)

	data, certs, err = Schedule(ctx, tx, "Test", "PF2B5BFF", now)
	assert.NoError(t, err)
	assert.Nil(t, data)
	assert.Empty(t, certs)

	err = tx.Commit(ctx)
	assert.NoError(t, err)
}
