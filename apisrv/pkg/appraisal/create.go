package appraisal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	ev "github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

const SQL_CREATE_APPRAISAL = `
INSERT INTO v2.appraisals (
        received_at,
        appraised_at,
        expires,
        verdict,
        evidence_id,
        report,
        key_id,
        device_id
    )
VALUES (
        $1::timestamptz,
        $2::timestamptz,
        $3::timestamptz,
        $4,
        $5,
        $6,
        $7,
        $8
    )
RETURNING id;
`

var (
	appraisalLifetime = time.Hour * 24
)

// Create and persist an appraisal appraising a given evidence and issues (analysis results).
func Create(ctx context.Context, tx pgx.Tx, evidence *ev.Row, report *api.Report, result *check.Result, key *device.KeysRow, deviceId int64, now time.Time, actor string) (int64, error) {
	// verdict
	verdict := determineVerdict(result)
	verdictDoc, err := database.NewDocument(verdict)
	if err != nil {
		return 0, err
	}
	// report
	reportDoc, err := database.NewDocument(*report)
	if err != nil {
		return 0, err
	}

	var id int64
	err = pgxscan.Get(ctx, tx, &id, SQL_CREATE_APPRAISAL,
		evidence.ReceivedAt, // $1
		now,                 // $2
		evidence.ReceivedAt.Add(appraisalLifetime), // $3
		verdictDoc,  // $4
		evidence.Id, // $6
		reportDoc,   // $7
		key.Id,      // $8
		deviceId)    // $9
	if err != nil {
		return 0, fmt.Errorf("appraisal ins: %w", err)
	}

	// dedup issue list
	var issues []issuesv1.Issue
	seenIssues := make(map[string]bool)
	for _, iss := range result.Issues {
		if _, ok := seenIssues[iss.Id()]; ok {
			continue
		}
		issues = append(issues, iss)
		seenIssues[iss.Id()] = true
	}
	// bulk insert if issues are given (check here b/c CopyFrom will realize pretty late that the len would be 0)
	if len(issues) > 0 {
		count, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"v2", "issues_appraisals"},
			[]string{"appraisal_id", "issue_type", "incident", "payload"},
			pgx.CopyFromSlice(len(issues), func(i int) ([]interface{}, error) {
				issue := issues[i]
				// store complete issue here; would be more efficient to just store the args
				issueDoc, err := database.NewDocument(issue)
				if err != nil {
					err = fmt.Errorf("issue_appraisal ins: %w", err)
				}
				return []interface{}{id, issue.Id(), issue.Incident(), issueDoc}, err
			}))
		if err != nil {
			return 0, err
		}

		if count != int64(len(issues)) {
			return 0, errors.New("issues appraisals insert len mismatch")
		}
	}

	return id, nil
}

// Find all appraisals that have expired since the last call until 'now'.
/*
func Expire(ctx context.Context, tx pgx.Tx, now time.Time) ([]*api.Appraisal, []string, error) {
	var rows []struct {
		Device       int64  `db:"device"`
		Appraisal    int64  `db:"appraisal"`
		Organization string `db:"organization"`
	}

	err := pgxscan.Select(ctx, tx, &rows, `
		--
		-- Expire-Appraisals
		--
		WITH expired AS (
			SELECT
				appraisals.id,
				appraisals.device_id,
				organizations.external
			FROM v2.appraisals
			INNER JOIN v2.devices ON devices.id = appraisals.device_id
			INNER JOIN v2.organizations ON v2.organizations.id = v2.devices.organization_id
			WHERE appraisals.expires <= $1
			  AND NOT EXISTS (
					SELECT 1 FROM v2.expired_appraisals
					WHERE v2.expired_appraisals.id = appraisals.id)
		), new_expired_rows AS (
			INSERT INTO v2.expired_appraisals
				(id)
			SELECT expired.id FROM expired
			RETURNING id
		)
		SELECT
			expired.id        AS appraisal,
			expired.device_id AS device,
			expired.external  AS organization
		FROM
			expired
	`, now)
	if err != nil {
		return nil, nil, err
	}

	devices := make([]int64, len(rows))
	for i, pair := range rows {
		devices[i] = pair.Device
	}
	// ORDER BY devices.id DESC
	devs, err := device.SetNoOrganization(ctx, tx, devices, now)
	if err != nil {
		return nil, nil, err
	}

	appraisals := make([]int64, len(rows))
	for i, pair := range rows {
		appraisals[i] = pair.Appraisal
	}
	// ORDER BY appraisals.id DESC
	apprRows, err := fetchRows(ctx, tx, appraisals)
	if err != nil {
		return nil, nil, err
	}

	retApprs := make([]*api.Appraisal, len(rows))
	retOrgs := make([]string, len(rows))
	for i := 0; i < len(rows); i += 1 {
		devId := devices[i]
		apprId := appraisals[i]
		org := rows[i].Organization

		deviceIdx := sort.Search(len(devs), func(j int) bool {
			id, _ := strconv.ParseInt(devs[j].Id, 10, 64)
			return id <= devId
		})
		if id, err := strconv.ParseInt(devs[deviceIdx].Id, 10, 64); err != nil || id != devId {
			return nil, nil, errors.New("no missing device")
		}
		apprRowIdx := sort.Search(len(apprRows), func(j int) bool { return apprRows[j].Id <= apprId })
		if apprRows[apprRowIdx].Id != apprId {
			return nil, nil, errors.New("no matching appraisal")
		}
		appr, err := FromRow(&apprRows[apprRowIdx])
		if err != nil {
			return nil, nil, err
		}

		appr.Device = devs[deviceIdx]
		retApprs[i] = appr
		retOrgs[i] = org
	}

	return retApprs, retOrgs, nil
}
*/

func determineVerdict(result *check.Result) api.Verdict {
	var verdict = api.Verdict{
		Type:               api.VerdictType,
		Result:             api.Trusted,
		SupplyChain:        api.Trusted,
		Configuration:      api.Trusted,
		Firmware:           api.Trusted,
		Bootloader:         api.Trusted,
		OperatingSystem:    api.Trusted,
		EndpointProtection: api.Trusted,
	}

	if !result.SupplyChain {
		verdict.SupplyChain = api.Unsupported
	}
	if !result.EndpointProtection {
		verdict.EndpointProtection = api.Unsupported
	}
	for _, iss := range result.Issues {
		if !iss.Incident() {
			continue
		}

		switch iss.Aspect() {
		case issuesv1.SupplyChain:
			verdict.SupplyChain = api.Vulnerable
		case issuesv1.Configuration:
			verdict.Configuration = api.Vulnerable
		case issuesv1.Firmware:
			verdict.Firmware = api.Vulnerable
		case issuesv1.Bootloader:
			verdict.Bootloader = api.Vulnerable
		case issuesv1.OperatingSystem:
			verdict.OperatingSystem = api.Vulnerable
		case issuesv1.EndpointProtection:
			verdict.EndpointProtection = api.Vulnerable
		}
		verdict.Result = api.Vulnerable
	}

	return verdict
}
