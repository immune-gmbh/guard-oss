package policy

import (
	"errors"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

const PolicyType = "policy/1"

type Trinary string

const (
	True      = Trinary("on")
	False     = Trinary("off")
	IfPresent = Trinary("if-present")
)

func IsTrinary(s string) bool {
	return s == "on" || s == "off" || s == "if-present"
}

type ProtectedFile struct {
	Path string `json:"path"`
}

type Values struct {
	Type               string          `json:"type"`
	EndpointProtection Trinary         `json:"endpoint_protection"` // ELAM or IAM
	IntelTSC           Trinary         `json:"intel_tsc"`
	ProtectedFiles     []ProtectedFile `json:"protected_files,omitempty"`
}

var (
	NotPresentErr     = errors.New("policy not present")
	UnknownVersionErr = errors.New("unknown policy version")
)

func New() *Values {
	return &Values{
		Type:               PolicyType,
		EndpointProtection: IfPresent,
		IntelTSC:           IfPresent,
	}
}

func (pol Values) Serialize() (database.Document, error) {
	return database.NewDocument(pol)
}

func Parse(doc database.Document) (*Values, error) {
	var pol Values

	if doc.IsNull() {
		return nil, NotPresentErr
	}
	switch doc.Type() {
	case PolicyType:
		err := doc.Decode(&pol)
		if err != nil {
			return nil, err
		}
		return &pol, nil

	default:
		return nil, UnknownVersionErr
	}
}
