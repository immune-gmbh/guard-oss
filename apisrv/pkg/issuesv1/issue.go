package issuesv1

import (
	"encoding/json"
)

type Issue interface {
	Id() string
	Incident() bool
	Aspect() string
}

type GenericIssue struct {
	Common
	Args map[string]interface{} `json:"args"`
}

func (i *GenericIssue) Id() string {
	return i.Common.Id
}

func (i *GenericIssue) Incident() bool {
	return i.Common.Incident
}

func (i *GenericIssue) Aspect() string {
	return i.Common.Aspect
}

func GenericIssueFromRawMessage(rawMessage json.RawMessage) (*GenericIssue, error) {
	var gi GenericIssue
	err := json.Unmarshal(rawMessage, &gi)
	if err != nil {
		return nil, err
	}

	return &gi, nil
}

func New(issues []Issue) (*Issues, error) {
	msg := make([]json.RawMessage, len(issues))
	for i, ii := range issues {
		m, err := json.Marshal(ii)
		if err != nil {
			return nil, err
		}
		msg[i] = json.RawMessage(m)
	}
	iss := Issues{
		Type:   "issues/v1", // keep in sync w/ issuesv1.schema.yaml
		Issues: msg,
	}

	return &iss, nil
}
