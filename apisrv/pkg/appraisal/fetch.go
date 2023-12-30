package appraisal

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

// be sure to also sort by id descending to get the proper order in case there are
// identical timestamps; appraised_at should not be null to not get in-flight appraisals
const SQL_SELECT_LATEST_APPRAISALS_BY_DEVICE = `
SELECT id,
    device_id,
    appraised_at,
    expires,
    verdict,
    report
FROM v2.appraisals
WHERE device_id = $1
    AND appraised_at IS NOT NULL
ORDER BY appraised_at DESC,
    id DESC
LIMIT $2;
`

// this seems to be only relevant for test cases
const SQL_SELECT_APPRAISAL_BY_ID = `
SELECT id,
    device_id,
    appraised_at,
    expires,
    verdict,
    report
FROM v2.appraisals
WHERE id = $1;
`

const SQL_SELECT_ISSUES_BY_APPRAISALS = `
SELECT payload
FROM v2.issues_appraisals
WHERE appraisal_id = $1;
`

const SQL_SELECT_OPEN_INCIDENT_STATS = `
SELECT COUNT(*) AS incident_count,
    COUNT(DISTINCT d.id) AS devs_affected
FROM v2.devices d
    CROSS JOIN lateral (
        SELECT a.id,
            a.appraised_at
        FROM v2.appraisals a
        WHERE d.id = a.device_id
        ORDER BY a.appraised_at DESC
        LIMIT 1
    ) a
    INNER JOIN v2.issues_appraisals ia ON a.id = ia.appraisal_id
WHERE d.retired = false
    AND d.organization_id = $1
    AND ia.incident;
`

const SQL_SELECT_DASHBOARD_STATS = `
SELECT COUNT(*) AS num_devices,
    CASE
        WHEN ia.devstate = 1 THEN 'unresponsive'
        WHEN ia.devstate = 2 THEN 'untrusted'
        WHEN ia.devstate = 3 THEN 'atrisk'
        ELSE 'trusted'
    END AS device_state
FROM v2.devices d
    CROSS JOIN lateral (
        SELECT a.id,
            a.appraised_at,
            a.expires
        FROM v2.appraisals a
        WHERE d.id = a.device_id
        ORDER BY a.appraised_at DESC
        LIMIT 1
    ) a
    CROSS JOIN lateral (
        (
            SELECT CASE
                    WHEN ia.incident THEN 2
					WHEN a.expires <= NOW() THEN 1
                    ELSE 3
                END AS devstate
            FROM v2.issues_appraisals ia
            WHERE a.id = ia.appraisal_id
            ORDER BY ia.incident DESC
            LIMIT 1
        )
        UNION
        SELECT a
        FROM (
                VALUES(0)
            ) s(a)
        WHERE a.id NOT IN (
                SELECT appraisal_id
                FROM v2.issues_appraisals
            )
            OR d.id NOT IN (
                SELECT device_id
                FROM v2.appraisals
            )
    ) ia
WHERE d.retired = false
    AND d.organization_id = $1
GROUP BY ia.devstate;
`

const SQL_SELECT_LATEST_ACTIVE_INCIDENTS = `
SELECT MAX(a.appraised_at) AS latest_date,
    -- this works b/c UNIQUE(ia.issue_type, ia.appraisal_id)
    -- and each device returns max one appraisal here
    COUNT(ia.issue_type) AS num_affected,
    ia.issue_type
FROM v2.devices d
    CROSS JOIN lateral (
        SELECT a.id,
            a.appraised_at
        FROM v2.appraisals a
        WHERE d.id = a.device_id
        ORDER BY a.appraised_at DESC
        LIMIT 1
    ) a
    INNER JOIN v2.issues_appraisals ia ON a.id = ia.appraisal_id
WHERE d.retired = false
    AND d.organization_id = $1
    AND ia.incident
GROUP BY ia.issue_type
ORDER BY latest_date DESC;
`

const SQL_SELECT_RISKS_PER_TYPE = `
SELECT COUNT(ia.issue_type) AS occurences_count,
    ia.issue_type
FROM v2.devices d
    CROSS JOIN lateral (
        SELECT a.id,
            a.appraised_at
        FROM v2.appraisals a
        WHERE d.id = a.device_id
            AND a.appraised_at >= $1
        ORDER BY a.appraised_at DESC
        LIMIT 1
    ) a
    INNER JOIN v2.issues_appraisals ia ON a.id = ia.appraisal_id
WHERE d.retired = false
    AND d.organization_id = $2
    AND NOT ia.incident
GROUP BY ia.issue_type
ORDER BY occurences_count DESC;
`

// GetLatestAppraisalsByDeviceId returns appraisals for a device ordered by date descending
func GetLatestAppraisalsByDeviceId(ctx context.Context, qq pgxscan.Querier, deviceId int64, withIssues bool, limit uint) ([]*api.Appraisal, error) {
	rows, err := qq.Query(ctx, SQL_SELECT_LATEST_APPRAISALS_BY_DEVICE, deviceId, limit)
	if err != nil {
		return nil, database.Error(err)
	}
	defer rows.Close()

	// directly scan into issues here to avoid allocating possibly large arrays of rowstructs
	// only to allocate and convert into issues array, doubling ram consumption
	var appraisals []*api.Appraisal
	for rows.Next() {
		var row Row
		err = pgxscan.ScanRow(&row, rows)
		if err != nil {
			return nil, database.Error(err)
		}

		appraisal, err := FromRow(&row)
		if err != nil {
			return nil, fmt.Errorf("failed converting appraisal row: %w", err)
		}

		appraisals = append(appraisals, appraisal)
	}

	if err := rows.Err(); err != nil {
		return nil, database.Error(err)
	}

	if withIssues {
		for _, appraisal := range appraisals {
			id, err := strconv.ParseInt(appraisal.Id, 10, 64)
			if err != nil {
				// srsly converting the IDs from string and back inside the same function and needing error branches is just a waste
				return nil, err
			}

			issues, err := GetIssuesByAppraisal(ctx, qq, id)
			if err != nil {
				return nil, err
			}

			// sort to make webapp show correct incident first (with respect to trust chain graphic)
			sortIssues(issues)

			issuesJson, err := issuesv1.New(issues)
			if err != nil {
				return nil, err
			}
			appraisal.Issues = issuesJson
		}
	}

	return appraisals, nil
}

// GetAppraisalById is only used in unit test code and does not return issues to keep things simple
func GetAppraisalById(ctx context.Context, qq pgxscan.Querier, id int64) (*api.Appraisal, error) {
	var row Row
	err := pgxscan.Get(ctx, qq, &row, SQL_SELECT_APPRAISAL_BY_ID, id)
	if err != nil {
		return nil, database.Error(err)
	}

	appraisal, err := FromRow(&row)
	if err != nil {
		return nil, fmt.Errorf("failed converting appraisal row: %w", err)
	}

	return appraisal, nil
}

func GetIssuesByAppraisal(ctx context.Context, qq pgxscan.Querier, appraisalId int64) ([]issuesv1.Issue, error) {
	rows, err := qq.Query(ctx, SQL_SELECT_ISSUES_BY_APPRAISALS, appraisalId)
	if err != nil {
		return nil, database.Error(err)
	}
	defer rows.Close()

	// directly scan into issues here to avoid allocating possibly large arrays of rowstructs
	// only to allocate and convert into issues array, doubling ram consumption
	issues := []issuesv1.Issue{}
	for rows.Next() {
		var doc database.Document
		err = rows.Scan(&doc)
		if err != nil {
			return nil, database.Error(err)
		}

		var issue issuesv1.GenericIssue
		err = doc.Decode(&issue)
		if err != nil {
			return nil, err
		}
		issues = append(issues, &issue)
	}

	if err := rows.Err(); err != nil {
		return nil, database.Error(err)
	}

	return issues, nil
}

// GetIncidentStats returns the count of all open incidents and the number of devices affected
func GetIncidentStats(ctx context.Context, qq pgxscan.Querier, organizationId int64) (RowOpenIncidentsStats, error) {
	var row RowOpenIncidentsStats
	err := pgxscan.Get(ctx, qq, &row, SQL_SELECT_OPEN_INCIDENT_STATS,
		organizationId) // $1

	return row, database.Error(err)
}

func GetDashboardStats(ctx context.Context, qq pgxscan.Querier, organizationId int64) ([]DashboardStats, error) {
	var rows []DashboardStats
	err := pgxscan.Select(ctx, qq, &rows, SQL_SELECT_DASHBOARD_STATS,
		organizationId) // $1

	return rows, database.Error(err)
}

func GetLatestIncidents(ctx context.Context, qq pgxscan.Querier, organizationId int64) ([]RowLatestIncidents, error) {
	var rows []RowLatestIncidents
	err := pgxscan.Select(ctx, qq, &rows, SQL_SELECT_LATEST_ACTIVE_INCIDENTS,
		organizationId) // $1

	return rows, database.Error(err)
}

func GetRisksToplist(ctx context.Context, qq pgxscan.Querier, since time.Time, organizationId int64) ([]RowRisksByType, error) {
	var rows []RowRisksByType
	err := pgxscan.Select(ctx, qq, &rows, SQL_SELECT_RISKS_PER_TYPE,
		since,          // $1
		organizationId) // $2

	return rows, database.Error(err)
}

func FromRow(row *Row) (*api.Appraisal, error) {
	var report api.Report
	if rep, err := reportFromRow(row.Report); err != nil {
		return nil, err
	} else if rep != nil {
		report = *rep
	}
	verdict, err := verdictFromRow(row.Verdict)
	if err != nil {
		return nil, err
	}

	return &api.Appraisal{
		Id: fmt.Sprintf("%d", row.Id),
		// deprecated (filling this field is for backwards compatibility with API but no client uses it; using the receiced_at value from evidence would complicate the SQL without effect)
		// XXX deprecating this seems wrong, we need somehow be able to tell when the device was last seen in case we re-appraise with changed baseline due to override
		// XXX it looks like the code determining if a device is unseen for a while is buggy b/c the received field does not exist in the webapp api type and it looks like it isn't used anywhere in apisrv
		Received:  row.AppraisedAt,
		Appraised: row.AppraisedAt,
		Expires:   row.ExpiresAt,
		Verdict:   *verdict,
		// deprecated
		Report: report,
	}, nil
}

func issueAspectOrder(category string) int {
	switch category {
	default:
		return -1
	case issuesv1.SupplyChain:
		return 0
	case issuesv1.Firmware:
		return 1
	case issuesv1.Configuration:
		return 2
	case issuesv1.Bootloader:
		return 3
	case issuesv1.OperatingSystem:
		return 4
	case issuesv1.EndpointProtection:
		return 5
	}
}

func sortIssues(issues []issuesv1.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Incident() && !issues[j].Incident() {
			return true
		}
		if !issues[i].Incident() && issues[j].Incident() {
			return false
		}
		return issueAspectOrder(issues[i].Aspect()) < issueAspectOrder(issues[j].Aspect())
	})
}

func reportFromRow(doc database.Document) (*api.Report, error) {
	if doc.IsNull() {
		return nil, nil
	}

	switch doc.Type() {
	case "report/2":
		var report api.Report
		err := doc.Decode(&report)
		if err != nil {
			return nil, err
		}
		// strip down report
		report.Annotations = nil
		v := &report.Values
		v.AMT = nil
		v.ME = nil
		v.SEV = nil
		v.SGX = nil
		v.TPM = nil
		v.TXT = nil
		v.UEFI = nil
		return &report, err

	default:
		return nil, errors.New("unknown report type")
	}
}

func verdictFromRow(doc database.Document) (*api.Verdict, error) {
	if doc.IsNull() {
		return nil, errors.New("is null")
	}

	b2s := func(b bool) string {
		if b {
			return api.Trusted
		} else {
			return api.Vulnerable
		}
	}

	switch doc.Type() {
	// there are still some live "last" appraisals on production referencing verdict V1 but not atfer 04/22
	case api.VerdictTypeV1:
		var v1 api.VerdictV1
		err := doc.Decode(&v1)
		if err != nil {
			return nil, err
		}
		verdict := api.Verdict{
			Type:               api.VerdictType,
			Result:             b2s(v1.Result),
			SupplyChain:        api.Trusted,
			Configuration:      b2s(v1.Configuration),
			Firmware:           b2s(v1.Firmware),
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		}
		return &verdict, nil
	// there are still some live "last" appraisals on production referencing verdict V2 but not atfer 06/22
	case api.VerdictTypeV2:
		var v2 api.VerdictV2
		err := doc.Decode(&v2)
		if err != nil {
			return nil, err
		}
		verdict := api.Verdict{
			Type:               api.VerdictType,
			Result:             b2s(v2.Result),
			SupplyChain:        b2s(v2.SupplyChain),
			Configuration:      b2s(v2.Configuration),
			Firmware:           b2s(v2.Firmware),
			Bootloader:         b2s(v2.Bootloader),
			OperatingSystem:    b2s(v2.OperatingSystem),
			EndpointProtection: b2s(v2.EndpointProtection),
		}
		return &verdict, nil
	case api.VerdictType:
		var verdict api.Verdict
		err := doc.Decode(&verdict)
		return &verdict, err

	default:
		return nil, errors.New("unknown verdict type")
	}
}
