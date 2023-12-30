package appraisal

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// this deletes all issues referenced by appraisals older than 6 month
const SQL_GC_ISSUES_APPRAISALS = `
DELETE FROM v2.issues_appraisals
WHERE appraisal_id IN (
        SELECT id
        FROM v2.appraisals
        WHERE appraised_at < NOW() - INTERVAL '6 month'
    );
`

// this deletes all appraisals older than 6 month
const SQL_GC_APPRAISALS = `
DELETE FROM v2.appraisals
WHERE appraised_at < NOW() - INTERVAL '6 month';
`

func GarbageCollect(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, SQL_GC_ISSUES_APPRAISALS)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, SQL_GC_APPRAISALS)
	if err != nil {
		return err
	}

	return nil
}
