package tag

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

var (
	InvalidKsuidErr = errors.New("invalid ksuid")
)

const SQL_SELECT_TAGS_BY_DEVICE = `
SELECT t.id,
    t.key
FROM v2.tags t
    INNER JOIN v2.devices_tags dt ON t.id = dt.tag_id
WHERE dt.device_id = $1
ORDER BY t.id ASC;
`

func GetTagsByDeviceId(ctx context.Context, qq pgxscan.Querier, deviceId int64) ([]Row, error) {
	var rows []Row
	err := pgxscan.Select(ctx, qq, &rows, SQL_SELECT_TAGS_BY_DEVICE,
		deviceId) // $1

	return rows, database.Error(err)
}

type Query interface {
	execute(ctx context.Context, qq pgxscan.Querier, st Statement) ([]*Row, error)
}

// the PostgreSQL query planner can't handle t.id < $x so in order to get a
// more efficient query we have to express the range boundary as a concrete
// value. See https://github.com/rails/rails/issues/22566
func rangeClause(start string) (string, error) {
	if !database.ValidKsuid(start) {
		return "", InvalidKsuidErr
	}
	return fmt.Sprintf("and t.id < '%s'", start), nil
}

// query based on the primary key
type idQuery struct {
	start *string
	set   []string
	org   *int64
	limit int
	min   int
}

func (q idQuery) execute(ctx context.Context, qq pgxscan.Querier, st Statement) ([]*Row, error) {
	var rows []*Row
	var rc string
	var err error

	cs := st.columnSet()
	if q.start != nil {
		rc, err = rangeClause(*q.start)
		if err != nil {
			return nil, err
		}
	}

	err = pgxscan.Select(ctx, qq, &rows, fmt.Sprintf(`
    --
    -- Tag fetch
    --
    select
      %[1]s,
      devices.value as devices
    from v2.tags t
    left join lateral (
      select array_agg(v2.devices_tags.device_id) as value
      from v2.devices_tags
      where v2.devices_tags.tag_id = t.id
    ) devices on $1
    where ($3::bigint is null or t.organization_id = $3)
      and ($2::v2.ksuid[] is null or t.id = any($2))
      %[2]s
    order by t.id desc
    limit $4
  `, cs, rc), st.includeDevices, q.set, q.org, q.limit)
	if err != nil {
		return nil, database.Error(err)
	}
	if len(rows) < q.min {
		return nil, database.ErrNotFound
	}
	return rows, nil
}

// retrieve a single tag by id, owned by organization org
func Point(id string, org int64) idQuery {
	return Set(&org, id)
}

// set of tags by id, skip organization check is org is nil. the latter is used
// internally when it's certain that all tags belong to the same user.
func Set(org *int64, set ...string) idQuery {
	if set == nil {
		set = []string{}
	}
	sort.Strings(set)
	deduped := []string{}
	for _, id := range set {
		if len(deduped) == 0 || deduped[len(deduped)-1] != id {
			deduped = append(deduped, id)
		}
	}
	return idQuery{nil, deduped, org, len(deduped), len(deduped)}
}

// first 50 tags with id greater than after that are owned by org
func Range(after *string, org int64) idQuery {
	start := database.MaxKsuid
	if after != nil {
		start = *after
	}
	return idQuery{&start, nil, &org, 50, 0}
}

type textQuery struct {
	fragment string
	org      int64
	limit    int
}

// query based on (partial) text match
func (q textQuery) execute(ctx context.Context, qq pgxscan.Querier, st Statement) ([]*Row, error) {
	var rows []*Row
	err := pgxscan.Select(ctx, qq, &rows, fmt.Sprintf(`
    --
    -- Tag full text search
    --
    select
      %[1]s,
      similarity(t.key, $3) as score,
      devices.value as devices
    from v2.tags t
    left join lateral (
      select array_agg(v2.devices_tags.device_id) as value
      from v2.devices_tags
      where v2.devices_tags.tag_id = t.id
    ) devices on $1
    where t.organization_id = $2
      and t.key %% $3
    order by score desc
    limit $4
  `, st.columnSet()), st.includeDevices, q.org, q.fragment, q.limit)
	if err != nil {
		return nil, database.Error(err)
	}
	return rows, nil
}

// first 10 tags with key containing a given text fragment, owed by the given
// organization, sorted by match quality.
func Text(fragment string, org int64) textQuery {
	return textQuery{fragment, org, 10}
}

type Statement struct {
	ctx            context.Context
	query          Query
	columns        []string
	includeDevices bool
}

// column set to return in its escaped form
func (st Statement) columnSet() string {
	cols := make([]string, len(st.columns))
	for i := range st.columns {
		cols[i] = pgx.Identifier{"t", st.columns[i]}.Sanitize()
	}
	return strings.Join(cols, ",")
}

// set the columns to be fetched from the database.
func (st Statement) Columns(columns ...string) Statement {
	if len(columns) > 0 {
		st.columns = make([]string, len(columns))
		for i, col := range columns {
			if _, ok := allowedColumns[col]; ok {
				st.columns[i] = col
			}
		}
	}
	return st
}

// set the relations to be fetched from the database. default is none, the only
// value relations is "devices". this will also fetch all device ids associated
// with a tag. the device columns need to be fetched seperatly.
func (st Statement) Relations(relations ...string) Statement {
	sort.Strings(relations)
	st.includeDevices = sort.SearchStrings(relations, "devices") < len(relations)
	return st
}

// execute the query
func (st Statement) Do(qq pgxscan.Querier) ([]*Row, error) {
	return st.query.execute(st.ctx, qq, st)
}

// execute the query, expect a single row
func (st Statement) DoSingle(qq pgxscan.Querier) (*Row, error) {
	rows, err := st.Do(qq)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, database.ErrNotFound
	}
	return rows[0], nil
}

// fetch tags from the database
func Fetch(ctx context.Context, q Query) Statement {
	return Statement{
		ctx:            ctx,
		query:          q,
		columns:        defaultColumns,
		includeDevices: false,
	}
}
