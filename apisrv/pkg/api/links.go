package api

import (
	"fmt"
	"path"
	"strconv"

	"github.com/google/jsonapi"
)

var (
	BaseURL = "https://api.immu.ne/v2"
)

func (a *Appraisal) SetLinkSelfWeb(link string) {
	a.linkSelfWeb = link
}

func (a Appraisal) JSONAPILinks() *jsonapi.Links {
	if a.linkSelfWeb != "" {
		return &jsonapi.Links{
			"self-web": a.linkSelfWeb,
		}
	}
	return &jsonapi.Links{}
}

func (dev *Device) SetLinkSelfWeb(link string) {
	dev.linkSelfWeb = link
}

func (dev Device) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self":       fmt.Sprintf(path.Join(BaseURL, "devices/%s"), dev.Id),
		"self-web":   dev.linkSelfWeb,
		"changes":    fmt.Sprintf(path.Join(BaseURL, "devices/%s/changes"), dev.Id),
		"appraisals": fmt.Sprintf(path.Join(BaseURL, "devices/%s/appraisals"), dev.Id),
	}
}

func (dev Device) JSONAPIRelationshipLinks(relation string) *jsonapi.Links {
	if relation == "appraisals" {
		links := jsonapi.Links{
			"self": fmt.Sprintf(path.Join(BaseURL, "devices/%s/appraisals"), dev.Id),
		}

		var maxId *int64 = nil
		for _, appr := range dev.Appraisals {
			if id, err := strconv.ParseInt(appr.Id, 10, 64); err == nil {
				if maxId == nil || *maxId < id {
					maxId = &id
				}
			}
		}
		if maxId != nil {
			links["next"] = fmt.Sprintf(path.Join(BaseURL, "devices/%s/appraisals?i=%d"), dev.Id, *maxId+1)
		}

		return &links
	}
	return nil
}
