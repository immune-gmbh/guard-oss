// Signed result of a remote attestation. Determines the state (trusted,
// vulnerable, etc...) of a device at a specific point in time.
package appraisal

import (
	"time"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

// Appraisal structure in database. Creating an appraisal is a two step
// process. First the evidence of the device's state is submitted and the
// cryptographic signatures on the evidence is checked. and whether all
// information expected are present. Second, the evidence is exaimed more
// deeply. This involves long running tasks like API calls and us done
// asynchronously. After the second step is done the appraisal is finalized and
// becomes the last know state of the device. Before finalization, the
// appraisal is not valid.
type Row struct {
	Id          int64             `db:"id"`
	AppraisedAt time.Time         `db:"appraised_at"`
	ExpiresAt   time.Time         `db:"expires"`
	Verdict     database.Document `db:"verdict"` // api.Verdict
	EvidenceId  *string           `db:"evidence_id"`
	DeviceId    int64             `db:"device_id"`

	Report database.Document `db:"report"` // deprecated
}

type RowOpenIncidentsStats struct {
	IncidentCount   int `db:"incident_count"`
	DevicesAffected int `db:"devs_affected"`
}

type DashboardStats struct {
	NumDevices  int    `db:"num_devices"`
	DeviceState string `db:"device_state"`
}

type RowLatestIncidents struct {
	IssueTypeId     string    `db:"issue_type"`
	LatestOccurence time.Time `db:"latest_date"`
	DevicesAffected int       `db:"num_affected"`
}

type RowRisksByType struct {
	IssueTypeId   string `db:"issue_type"`
	NumOccurences int    `db:"occurences_count"`
}

func (r *RowLatestIncidents) ToApiType() *api.IncidentStatsEntry {
	return &api.IncidentStatsEntry{
		IssueTypeId:     r.IssueTypeId,
		LatestOccurence: r.LatestOccurence,
		DevicesAffected: r.DevicesAffected,
	}
}

func (r *RowRisksByType) ToApiType() *api.RiskStatsEntry {
	return &api.RiskStatsEntry{
		IssueTypeId:   r.IssueTypeId,
		NumOccurences: r.NumOccurences,
	}
}
