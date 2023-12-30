package change

import (
	"context"

	"github.com/jackc/pgx/v4"
)

const SQL_GC_CHANGES = `
DELETE FROM v2.changes
WHERE timestamp < NOW() - INTERVAL '6 month';
`

func GarbageCollect(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, SQL_GC_CHANGES)
	if err != nil {
		return err
	}

	return nil
}
