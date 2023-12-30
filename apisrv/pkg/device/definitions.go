package device

import (
	"time"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/jackc/pgtype"
)

type KeysRow struct {
	Id         int64         `db:"id"`
	Public     api.PublicKey `db:"public"`
	Name       string        `db:"name"`
	QName      api.Name      `db:"fpr"`
	Credential string        `db:"credential"`
	DeviceId   int64         `db:"device_id"`
}

type Row struct {
	Id                    int64             `db:"id"`
	HardwareFingerprint   api.Name          `db:"hwid"` // EK qname
	Fingerprint           api.Name          `db:"fpr"`  // Root qname
	Name                  string            `db:"name"`
	Cookie                *string           `db:"cookie"`
	Retired               bool              `db:"retired"`
	Baseline              database.Document `db:"baseline"`
	Policy                database.Document `db:"policy"`
	OrganizationId        int64             `db:"organization_id"`
	ReplacedBy            *int64            `db:"replaced_by"`
	AttestationInProgress *time.Time        `db:"attestation_in_progress"` // virtual
	State                 string            `db:"state"`                   // virtual v2.devices_state

	parsedPolicy *policy.Values

	// deprecated
	Attributes pgtype.Hstore `db:"attributes"`
}

func (row *Row) GetPolicy() (*policy.Values, error) {
	if row.parsedPolicy != nil {
		return row.parsedPolicy, nil
	}
	pol, err := policy.Parse(row.Policy)
	if err != nil {
		return nil, err
	}
	row.parsedPolicy = pol
	return pol, nil
}

// Specialized device row with reduced field-set to be used during attestation flows
// AIK is attest key in key table; FPR is tpm indentity; AIK is only used during attest
// FPR is used for attest auth and enroll to see if device is duped
type DevAikRow struct {
	Id             int64             `db:"id"`
	Fingerprint    api.Name          `db:"fpr"` // Root qname
	Name           string            `db:"name"`
	Retired        bool              `db:"retired"`
	Baseline       database.Document `db:"baseline"`
	Policy         database.Document `db:"policy"`
	OrganizationId int64             `db:"organization_id"`
	AIK            *KeysRow          `db:"aik"` // virtual

	parsedPolicy *policy.Values
}

func (row *DevAikRow) GetPolicy() (*policy.Values, error) {
	if row.parsedPolicy != nil {
		return row.parsedPolicy, nil
	}
	pol, err := policy.Parse(row.Policy)
	if err != nil {
		return nil, err
	}
	row.parsedPolicy = pol
	return pol, nil
}
