package inteltsc

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

type Row struct {
	Id        string    `db:"id"`
	Reference string    `db:"reference"` // weak, deterministic ref. used by appraisals
	CreatedAt time.Time `db:"created_at"`

	// done
	Data         *string    `db:"data"`
	Certificates []string   `db:"certificates"`
	FinishedAt   *time.Time `db:"finished_at"`
}

const (
	currentExternalVersion = 1
)

func getRow(ctx context.Context, q pgxscan.Querier, ref string) (*Row, error) {
	var row Row

	err := pgxscan.Get(ctx, q, &row, `
    select * from v2.inteltsc where reference = $1
  `, ref)
	if err != nil {
		return nil, database.Error(err)
	}

	return &row, nil
}

func updateRow(ctx context.Context, q pgxscan.Querier, row *Row) error {
	err := pgxscan.Get(ctx, q, row, `
    update v2.inteltsc
    set (
      data,
      certificates,
      finished_at
    ) = (
      $1::xml,
      $2::text[],
      $3::timestamptz
    )
    where id = $4
    returning *
  `,
		row.Data,         // $1
		row.Certificates, // $2
		row.FinishedAt,   // $3
		row.Id)           // $4

	return database.Error(err)
}

func insertRow(ctx context.Context, q pgxscan.Querier, row *Row) error {
	err := pgxscan.Get(ctx, q, row, `
    insert into v2.inteltsc (
      reference,
      created_at
    ) values (
      $1,
      $2::timestamptz
    )
    returning *
  `,
		row.Reference, // $1
		row.CreatedAt) // $2

	return database.Error(err)
}
