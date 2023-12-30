package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	queryIsolatedRetries = 5
)

func QueryIsolated(ctx context.Context, pool *pgxpool.Pool, fn func(context.Context, pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err == nil {
		defer tx.Rollback(ctx)
		err = fn(ctx, tx)
		if err == nil {
			err = tx.Commit(ctx)
			if err == nil {
				return nil
			}
		}
	}

	return Error(err)
}

func QueryIsolatedRetry(ctx context.Context, pool *pgxpool.Pool, fn func(context.Context, pgx.Tx) error) error {
	for i := 0; i < queryIsolatedRetries; i += 1 {
		err := QueryIsolated(ctx, pool, fn)
		if errors.Is(err, ErrSerialization) {
			continue
		} else {
			return err
		}
	}

	return ErrTooManyRetries
}
