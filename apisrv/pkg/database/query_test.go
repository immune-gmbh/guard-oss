package database

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestRepeatableRead(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	// Create database
	conn := pgsqlC.ConnectAdmin(t, ctx)
	err := EnsureDatabaseExists(ctx, conn, pgsqlC.UserDatabase)
	assert.NoError(t, err)

	_, err = conn.Exec(ctx, `
		create table test (
			id bigserial primary key,
			value bigint
		);`)
	assert.NoError(t, err)
	_, err = conn.Exec(ctx, `insert into test (id, value) values (1, 1);`)
	assert.NoError(t, err)

	wg := new(sync.WaitGroup)
	wg.Add(5)
	serErrd := false
	originalRetries := queryIsolatedRetries
	queryIsolatedRetries = 500

	for i := 0; i < 5; i += 1 {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for j := 0; j < 10; j += 1 {
				err := QueryIsolatedRetry(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
					_, err := tx.Exec(ctx, "update test set value = value + 1 where id = 1;")
					serErrd = serErrd || errors.Is(Error(err), ErrSerialization)
					return err
				})
				assert.NoError(t, err)
			}
		}()
	}

	wg.Wait()
	queryIsolatedRetries = originalRetries
	var value int
	err = pgxscan.Get(ctx, conn, &value, "select value from test where id = 1;")
	assert.NoError(t, err)
	assert.Equal(t, 5*10+1, value)
	assert.True(t, serErrd)
}

func TestRepeatableReadError(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	// Create database
	conn := pgsqlC.ConnectAdmin(t, ctx)
	err := EnsureDatabaseExists(ctx, conn, pgsqlC.UserDatabase)
	assert.NoError(t, err)

	err = QueryIsolatedRetry(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		return assert.AnError
	})
	assert.Equal(t, assert.AnError, err)
}
