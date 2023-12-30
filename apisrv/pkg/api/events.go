// Keep in sync with agent/pkg/api/types.go
package api

import "time"

var (
	FeatureReport       = "report"
	FeatureDeepAnalysis = "deep"
)

// /v2/events
// dataschema: "https://immu.ne/specs/v2/events/quota-update.schema.json"
// subject: organization.id
var QuotaUpdateEventType = "ne.immu.v2.quote-update"

type QuotaUpdateEvent struct {
	Devices  uint     `json:"devices,string"`
	Features []string `json:"features"` // Feature*
}

// /v2/events
// dataschema "https://immu.ne/specs/v2/events/billing-update.schema.json"
// no subject
var BillingUpdateEventType = "ne.immu.v2.billing-update"

type BillingUpdateEvent struct {
	Timestamp time.Time                  `json:"timestamp"`
	Records   []BillingUpdateUsageRecord `json:"records"`
}
type BillingUpdateUsageRecord struct {
	Organisation string `json:"organisation"`
	Devices      int    `json:"devices,string"`
}

// /v2/events (authsrv)
// dataschema "https://immu.ne/specs/v2/events/new-appraisal.schema.json"
// subject device.id
var NewAppraisalEventType = "ne.immu.v2.new-appraisal"

type NewAppraisalEvent struct {
	Device   Device     `json:"device"`
	Previous *Appraisal `json:"previous,omitempty"`
	Next     Appraisal  `json:"next"`
}

// /v2/events (authsrv)
// dataschema "https://immu.ne/specs/v2/events/appraisal-expired.schema.json"
// subject device.id
var AppraisalExpiredEventType = "ne.immu.v2.appraisal-expired"

type AppraisalExpiredEvent struct {
	Device   Device    `json:"device"`
	Previous Appraisal `json:"previous"`
}

// /v2/events (apisrv)
// dataschema "https://immu.ne/specs/v2/events/heartbeat.schema.json"
// no subject
var HeartbeatEventType = "ne.immu.v2.heartbeat"

type HeartbeatEvent struct {
	ExpireAppraisals    bool `json:"expire_appraisals"`
	CompressRevocations bool `json:"compress_revokations"`
	ReportUsage         bool `json:"report_usage"`
}

// /v2/events (apisrv)
// dataschema "https://immu.ne/specs/v2/events/revoke-credentials.schema.json"
// no subject
var RevokeCredentialsEventType = "ne.immu.v2.revoke-credentials"

type RevokeCredentialsEvent struct {
	Expires  time.Time `json:"expires,"`
	TokenIDs []string  `json:"token_ids"`
}
