package organization

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

type Row struct {
	Id             int64     `db:"id"`
	External       string    `db:"external"`
	Features       []string  `db:"features"`
	Devices        int64     `db:"devices"`
	UpdatedAt      time.Time `db:"updated_at"`
	CurrentDevices int64     `db:"current_devices"` // virtual
}

func UpdateQuota(ctx context.Context, pool *pgxpool.Pool, external string, devices uint, now time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO v2.organizations (external, devices, updated_at)
		VALUES ($3, $1, $2)
		ON CONFLICT (external) DO UPDATE SET devices = $1, updated_at = $2
	`, devices, now, external)

	return database.Error(err)
}

func Quota(ctx context.Context, qq pgxscan.Querier, orgId int64) (int, int, error) {
	var quota struct {
		Current int
		Allowed int
	}

	err := pgxscan.Get(ctx, qq, &quota, `
		--
		-- Fetch-Device-Quota
		--
    select
      devices.value as current,
      org.devices as allowed
    from v2.organizations org

    left join lateral (
      select
        count(devs.id) as value
      from v2.devices devs
      where devs.organization_id = org.id
        and devs.retired = false
    ) devices on true

    where org.id = $1
    limit 1
	`, orgId)
	if err != nil {
		return -1, -1, database.Error(err)
	}
	return quota.Current, quota.Allowed, nil
}

func Get(ctx context.Context, qq pgxscan.Querier, external string) (*Row, error) {
	var row Row
	err := pgxscan.Get(ctx, qq, &row, `
		SELECT 
			v2.organizations.id,
			v2.organizations.external,
			v2.organizations.features::TEXT[],
			v2.organizations.devices,
			v2.organizations.updated_at
		FROM v2.organizations
		WHERE v2.organizations.external = $1
	`, external)

	return &row, database.Error(err)
}

func GetExternalById(ctx context.Context, qq pgxscan.Querier, id int64) (*string, error) {
	var external string
	err := pgxscan.Get(ctx, qq, &external, "SELECT external FROM v2.organizations WHERE id = $1", id)
	if err != nil {
		return nil, database.Error(err)
	}
	if external == "" {
		return nil, database.ErrNotFound
	}
	return &external, nil
}
