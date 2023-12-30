package filter

import "fmt"

const SQL_SELECT_DEVICE_IDS_BY_ISSUE = `
SELECT d.id AS device_id
FROM v2.devices d
    CROSS JOIN lateral (
        SELECT a.id,
            a.appraised_at
        FROM v2.appraisals a
        WHERE d.id = a.device_id
        ORDER BY a.appraised_at DESC
        LIMIT 1
    ) a
    INNER JOIN v2.issues_appraisals b ON a.id = b.appraisal_id
WHERE d.retired = false
    AND d.organization_id = $%[1]d
    AND b.issue_type = $%[2]d
`

func buildQueryDevsByIssue(dollarOffset int, orgId int64, issueId string) (string, []any) {
	return fmt.Sprintf(SQL_SELECT_DEVICE_IDS_BY_ISSUE, dollarOffset, dollarOffset+1), []any{orgId, issueId}
}
