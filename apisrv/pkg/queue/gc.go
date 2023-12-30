package queue

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// GarbageCollect deletes old jobs from the database.
// finished_at should be telling that the job is done forever, but
// we check for next_run_at as well to be sure.
const SQL_GC_JOBS = `
DELETE FROM v2.jobs
WHERE finished_at IS NOT NULL
    AND finished_at < NOW() - INTERVAL '6 month'
    AND next_run_at < NOW() - INTERVAL '6 month';
`

func GarbageCollect(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, SQL_GC_JOBS)
	if err != nil {
		return err
	}

	return nil
}
