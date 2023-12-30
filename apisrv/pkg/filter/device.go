package filter

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
)

// Specialized query to list all active (not retired) devices for device list view
const SQL_SELECT_DEV_LIST_BY_ORG = `
SELECT devs.*,
    ds.state AS state,
    (
        CASE
            WHEN last_evidence.received_at < ($3 - INTERVAL '5 minutes') THEN NULL
            WHEN last_appraisal.id IS NOT NULL THEN NULL
            ELSE last_evidence.received_at
        END
    ) AS attestation_in_progress
FROM v2.devices devs
    LEFT JOIN lateral (
        SELECT a2.id,
            a2.verdict,
            a2.expires
        FROM v2.appraisals a2
        WHERE a2.device_id = devs.id
        ORDER BY a2.appraised_at DESC
        LIMIT 1
    ) last_appraisal ON TRUE
    LEFT JOIN lateral (
        SELECT ev.received_at
        FROM v2.evidence ev
        WHERE ev.device_id = devs.id
            AND NOT EXISTS (
                SELECT 1
                FROM v2.appraisals
                WHERE v2.appraisals.evidence_id = ev.id
            )
        ORDER BY ev.received_at DESC
        LIMIT 1
    ) last_evidence ON TRUE
    LEFT JOIN lateral (
        SELECT v2.device_state(
                false,
                -- this query is specialized and ignores retired devices
                false,
                42,
                -- circumvent device state 'new' to avoid joins with key table (instead of 'new', 'unseen' would show)
                last_appraisal.verdict,
                last_appraisal.expires,
                $3::timestamptz at time zone 'UTC'
            ) AS state
    ) ds ON TRUE
WHERE devs.organization_id = $1
    AND devs.retired = false %[1]s
ORDER BY devs.id DESC
LIMIT $2
`

const SQL_SELECT_DEV_LIST_BY_ORG_PARNUM = 3

func sanitizeFilterState(state string) string {
	switch state {
	case "vulnerable":
		fallthrough
	case "trusted":
		fallthrough
	case "outdated":
		return state
	default:
		return ""
	}
}

// queryActiveDevicesByOrganizationFilter is an optimized query that lists the active (not retired) devices with state for the device list and uses a parameterized subquery to filter with a device ID set
func queryActiveDevicesByOrganizationFilter(ctx context.Context, qq pgxscan.Querier, orgId, limit int64, firstDeviceId *int64, now time.Time, filterState, filterSetSql string, filterSetParams []any) ([]device.Row, error) {
	var rows []device.Row
	var sql, andDevId, andSubQuery, andState string

	// hack together our conditions
	if firstDeviceId != nil {
		andDevId = fmt.Sprint(" AND devs.id <=", *firstDeviceId)
	}
	if filterSetSql != "" {
		andSubQuery = fmt.Sprint(" AND devs.id IN (", filterSetSql, ")")
	}
	filterState = sanitizeFilterState(filterState)
	if filterState != "" {
		andState = fmt.Sprintf(" AND ds.state = $%d", SQL_SELECT_DEV_LIST_BY_ORG_PARNUM+len(filterSetParams)+1)
	}
	if len(andDevId) > 0 || len(filterSetSql) > 0 || len(filterState) > 0 {
		where := fmt.Sprint(andDevId, andSubQuery, andState)
		sql = fmt.Sprintf(SQL_SELECT_DEV_LIST_BY_ORG, where)
	} else {
		sql = fmt.Sprintf(SQL_SELECT_DEV_LIST_BY_ORG, "")
	}

	// base arg list
	args := []any{
		orgId, // $1
		limit, // $2
		now,   // $3
	}

	// args for filter set subquery
	if len(filterSetParams) > 0 {
		args = append(args, filterSetParams...)
	}

	// append state condition afterwards
	if len(andState) > 0 {
		args = append(args, filterState)
	}

	err := pgxscan.Select(ctx, qq, &rows, sql, args...)
	if err != nil {
		return nil, database.Error(err)
	}

	return rows, nil
}
