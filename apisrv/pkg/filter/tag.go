package filter

import "fmt"

const SQL_SELECT_DEVICE_IDS_BY_TAG = `
SELECT dt.device_id
FROM v2.devices_tags dt
WHERE dt.tag_id = $%[1]d
`

func buildQueryDevsByTags(dollarOffset int, tagId string) (string, []any) {
	return fmt.Sprintf(SQL_SELECT_DEVICE_IDS_BY_TAG, dollarOffset), []any{tagId}
}
