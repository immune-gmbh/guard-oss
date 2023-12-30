package tag

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func Device(ctx context.Context, qq pgxscan.Querier, dev int64, org int64, tags []string) (bool, error) {
	var changes struct {
		Added   int
		Deleted int
	}

	metadata, err := NewMetadata().Serialize()
	if err != nil {
		return false, err
	}
	err = pgxscan.Get(ctx, qq, &changes, `
    with tags as (
      insert into v2.tags (key, value, metadata, organization_id) 
      select key, '', $2::jsonb, $3::bigint from unnest($1::text[]) key
      on conflict (key, organization_id) do update set key = v2.tags.key
      returning id
    ), added as (
      insert into v2.devices_tags (device_id, tag_id)
      select $4::bigint as device_id, tags.id as tag_id from tags
      on conflict do nothing
      returning tag_id
    ), deleted as (
      delete from v2.devices_tags
      where device_id = $4::bigint and tag_id not in (select id from tags)
      returning 1
    )
    select
      (select count(*) from added) as added,
      (select count(*) from deleted) as deleted
`, tags, metadata, org, dev)
	if err != nil {
		return false, database.Error(err)
	}
	return changes.Added > 0 || changes.Deleted > 0, nil
}
