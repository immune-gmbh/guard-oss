package device

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"reflect"
	"strconv"
	"time"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

// Get a single device and associated AIK by AIK fingerprint
const SQL_SELECT_DEV_AIK_BY_FPR = `
SELECT d.id,
    d.fpr,
    d.name,
    d.retired,
    d.baseline,
    d.policy,
    d.organization_id,
    k.id AS "aik.id",
    k.public AS "aik.public",
    k.name AS "aik.name",
    k.fpr AS "aik.fpr",
    k.credential AS "aik.credential",
    k.device_id AS "aik.device_id"
FROM v2.devices d
    LEFT JOIN v2.keys k ON k.device_id = d.id
WHERE k.fpr = $1
ORDER BY d.id DESC -- if there are many devices with the same AIK then the newest one is the right one; it should however be ruled out on insert so we can save the effort of sorting here
LIMIT 1;
`

type query interface {
	FirstDeviceId() *int64
	DeviceIdSet() []int64
	Limit() int64
	Organization() int64
}

type rangeQuery struct {
	start int64
	max   int64
	orgId int64
}

func (q rangeQuery) FirstDeviceId() *int64 { return &q.start }
func (q rangeQuery) DeviceIdSet() []int64  { return nil }
func (q rangeQuery) Limit() int64          { return q.max }
func (q rangeQuery) Organization() int64   { return q.orgId }

func ListRow(ctx context.Context, qq pgxscan.Querier, iter *string, max int, orgId int64, now time.Time) ([]Row, *string, error) {
	if iter != nil {
		if i, err := strconv.ParseInt(*iter, 10, 32); err == nil {
			q := rangeQuery{start: i, max: int64(max), orgId: orgId}
			return fetch(ctx, qq, q, now)
		}
	}
	q := rangeQuery{start: math.MaxInt64, max: int64(max), orgId: orgId}
	return fetch(ctx, qq, q, now)
}

type setQuery struct {
	set   []int64
	orgId int64
}

func (q setQuery) FirstDeviceId() *int64 { return nil }
func (q setQuery) DeviceIdSet() []int64  { return q.set }
func (q setQuery) Limit() int64          { return int64(len(q.set)) }
func (q setQuery) Organization() int64   { return q.orgId }

func SetRow(ctx context.Context, qq pgxscan.Querier, ids []int64, orgId int64, now time.Time) ([]Row, error) {
	q := setQuery{set: ids, orgId: orgId}
	rows, _, err := fetch(ctx, qq, q, now)
	return rows, err
}

type pointQuery struct {
	device int64
	orgId  int64
}

func (q pointQuery) FirstDeviceId() *int64 { return nil }
func (q pointQuery) DeviceIdSet() []int64  { return []int64{q.device} }
func (q pointQuery) Limit() int64          { return 1 }
func (q pointQuery) Organization() int64   { return q.orgId }

func Get(ctx context.Context, qq pgxscan.Querier, id int64, orgId int64, now time.Time) (*Row, error) {
	q := pointQuery{device: id, orgId: orgId}
	rows, _, err := fetch(ctx, qq, q, now)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || rows[0].Id != id {
		return nil, database.ErrNotFound
	}
	return &rows[0], nil
}

// GetByFingerprint is an optimized query with a reduced field set used by attest workflows
func GetByFingerprint(ctx context.Context, qq pgxscan.Querier, aikName *api.Name) (*DevAikRow, error) {
	var row DevAikRow
	err := pgxscan.Get(ctx, qq, &row, SQL_SELECT_DEV_AIK_BY_FPR, aikName)
	if err != nil {
		return nil, database.Error(err)
	}
	if row.AIK == nil || !reflect.DeepEqual(row.AIK.QName, *aikName) {
		return nil, database.ErrNotFound
	}

	return &row, nil
}

func fetch(ctx context.Context, qq pgxscan.Querier, q query, now time.Time) ([]Row, *string, error) {
	var rows []Row

	sql := `
    select distinct on (devs.id)
      devs.*,
      v2.device_state(
					devs.retired,
					exists (select 1 from v2.devices AS d
						where d.hwid = devs.hwid
							AND d.id <> devs.id
							AND d.fpr = devs.fpr
							AND d.organization_id = devs.organization_id
							AND d.retired = false),
					42, -- circumvent device state 'new' to avoid joins with key table (instead of 'new', 'unseen' would show)
					last_appraisal.verdict,
					last_appraisal.expires,
	        $4::timestamptz at time zone 'UTC'
				) AS state,

        (case
         when last_evidence.received_at < ($4 - interval '5 minutes') then
           null
         when last_appraisal.id is not null then
           null
         else last_evidence.received_at
         end) as attestation_in_progress
		
    from v2.devices devs

    left join lateral (
      select
        a2.id,
        a2.verdict,
        a2.expires
      from v2.appraisals a2
      where a2.device_id = devs.id
      order by a2.appraised_at desc
      limit 1
    ) last_appraisal on true

    left join lateral (
      select
        ev.received_at
      from v2.evidence ev
      where ev.device_id = devs.id
        and not exists (
          select 1 from v2.appraisals where v2.appraisals.evidence_id = ev.id)
	    order by ev.received_at desc
	    limit 1
    ) last_evidence on true

    where ($2::bigint[] is null or devs.id = any($2))
      %[1]s
      and ($1::bigint = 0 or $1 = devs.organization_id)
    order by devs.id desc
    limit $3
  `

	if q.FirstDeviceId() != nil {
		sql = fmt.Sprintf(sql, fmt.Sprintf("and devs.id <= %d", *q.FirstDeviceId()))
	} else {
		sql = fmt.Sprintf(sql, "")
	}

	err := pgxscan.Select(ctx, qq, &rows, sql,
		q.Organization(), // $1
		q.DeviceIdSet(),  // $2
		q.Limit(),        // $3
		now,              // $4
	)
	if err != nil {
		return nil, nil, database.Error(err)
	}

	var next *int64
	for i := range rows {
		if next == nil || *next > rows[i].Id {
			next = &rows[i].Id
		}
	}

	if next != nil && *next > 0 {
		nextstr := fmt.Sprintf("%d", *next-1)
		return rows, &nextstr, nil
	} else {
		return rows, nil, nil
	}
}

func (row *Row) ToApiStruct() (*api.Device, error) {
	// cookie
	var cookie string
	if row.Cookie != nil {
		cookie = *row.Cookie
	}

	// hwid
	hwid, err := json.Marshal(row.HardwareFingerprint)
	if err != nil {
		return nil, err
	}

	// v2 policy
	policy, err := row.GetPolicy()
	if err != nil {
		return nil, err
	}

	dev := api.Device{
		Id:     fmt.Sprintf("%d", row.Id),
		Cookie: cookie,
		Name:   row.Name,
		Hwid:   string(hwid[1 : len(hwid)-2]),
		Policy: map[string]interface{}{
			"endpoint_protection": interface{}(policy.EndpointProtection),
			"intel_tsc":           interface{}(policy.IntelTSC),
		},
		State:                 row.State,
		AttestationInProgress: row.AttestationInProgress,
	}
	return &dev, nil
}

func FormatDevLinkSelfWeb(devId int64, organizationExternal string, webAppBaseURL string) string {
	return fmt.Sprintf(path.Join(webAppBaseURL, "devices/%s?organisation=%s"), strconv.FormatInt(devId, 10), organizationExternal)
}
