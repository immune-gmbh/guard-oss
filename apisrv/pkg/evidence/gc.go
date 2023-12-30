package evidence

import (
	"context"

	"github.com/jackc/pgx/v4"
)

const SQL_GC_EVIDENCE = `
DELETE FROM v2.evidence e
WHERE received_at < NOW() - INTERVAL '6 month'
	AND NOT EXISTS (SELECT 1 FROM v2.appraisals WHERE evidence_id = e.id);
`

func GarbageCollect(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, SQL_GC_EVIDENCE)
	if err != nil {
		return err
	}

	return nil
}
