package filter

import (
	"context"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
)

// ListActiveDevicesFiltered lists the not retired devices filtered by tag ids or issue id
func ListActiveDevicesFiltered(ctx context.Context, qq pgxscan.Querier, orgId, limit int64, firstDeviceId *int64, now time.Time, filterTagIds []string, filterIssueId, filterState string) ([]device.Row, error) {
	dollarOffset := SQL_SELECT_DEV_LIST_BY_ORG_PARNUM + 1
	var sqlSnippet string
	var sqlSnippets []string
	var params, paramSubset []any

	// add a filter snippet for an issue id (only one issue supported by now)
	if filterIssueId != "" {
		sqlSnippet, paramSubset = buildQueryDevsByIssue(dollarOffset, orgId, filterIssueId)
		dollarOffset = dollarOffset + len(paramSubset)
		params = append(params, paramSubset...)
		sqlSnippets = append(sqlSnippets, sqlSnippet)
	}

	// add a filter snippet for each tag id
	for _, tagId := range filterTagIds {
		sqlSnippet, paramSubset = buildQueryDevsByTags(dollarOffset, tagId)
		dollarOffset = dollarOffset + len(paramSubset)
		params = append(params, paramSubset...)
		sqlSnippets = append(sqlSnippets, sqlSnippet)
	}

	// build a possibly big intersect query to be used as filter and return the results
	sql := strings.Join(sqlSnippets, " INTERSECT ")
	return queryActiveDevicesByOrganizationFilter(ctx, qq, orgId, limit, firstDeviceId, now, filterState, sql, params)
}
