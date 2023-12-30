package check

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func TestFwupdCheck(t *testing.T) {
	type eviAnno struct {
		f string
		a bool
	}
	evidences := []eviAnno{
		{"../../test/dl160-fwupd-report.evidence.json", false},
		{"../../test/h12ssl-fwupd-report.evidence.json", true},
	}
	var subj *Subject
	for _, evidence := range evidences {
		subj = parseEvidence(t, evidence.f)
		iss := lvfsFirmwareUpdateCheck{}.Verify(context.Background(), subj)

		if evidence.a {
			assert.Equalf(t, issuesv1.FirmwareUpdateId, iss.Id(), "evidence %s", evidence.f)
		} else {
			assert.Nil(t, iss, "evidence %s", evidence.f)
		}
	}

	Run(context.Background(), subj)
	assert.False(t, subj.Baseline.AllowOutdatedFirmware)
	overrides := []string{issuesv1.FirmwareUpdateId}
	Override(context.Background(), overrides, subj)
	assert.True(t, subj.BaselineModified)
	assert.True(t, subj.Baseline.AllowOutdatedFirmware)
}

func filterNonFatal(res *Result) []string {
	var ret []string
	for _, iss := range res.Issues {
		fmt.Println(iss)
		if iss.Incident() {
			ret = append(ret, iss.Id())
		}
	}
	return ret
}

func TestNoChange(t *testing.T) {
	beforesubj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	aftersubj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-bootorder-reboot.json")

	res, err := Run(context.Background(), beforesubj)
	assert.NoError(t, err)
	assert.Empty(t, filterNonFatal(res))
	assert.True(t, beforesubj.BaselineModified)

	aftersubj.Baseline = beforesubj.Baseline
	res, err = Run(context.Background(), aftersubj)
	assert.NoError(t, err)
	assert.Empty(t, filterNonFatal(res))
}

func TestBadChange(t *testing.T) {
	beforesubj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	aftersubj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-no-secureboot.json")

	res, err := Run(context.Background(), beforesubj)
	assert.NoError(t, err)
	assert.Empty(t, filterNonFatal(res))
	assert.True(t, beforesubj.BaselineModified)
	aftersubj.Baseline = beforesubj.Baseline

	res, err = Run(context.Background(), aftersubj)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)

	var overrides []string
	for _, iss := range res.Issues {
		overrides = append(overrides, string(iss.Id()))
	}
	Override(context.Background(), overrides, aftersubj)
	assert.True(t, aftersubj.BaselineModified)
	aftersubj.BaselineModified = false

	res, err = Run(context.Background(), aftersubj)
	assert.NoError(t, err)
	assert.Empty(t, res.Issues)
	assert.False(t, aftersubj.BaselineModified)
}

func TestSHA1EventLog(t *testing.T) {
	subject := parseEvidence(t, "../../test/imn-dell-sha1.evidence.json")
	evLog := subject.Values.TPM2EventLogs[len(subject.Values.TPM2EventLogs)-1]
	assert.NotEmpty(t, evLog)

	issues, err := Run(context.Background(), subject)
	assert.NoError(t, err)
	assert.Empty(t, filterNonFatal(issues))
}

func TestMultipleEventLogsCompressed(t *testing.T) {
	subject := parseEvidence(t, "../../test/dell-notebook-5eventlogs-WBCL.evidence.json")
	if assert.NotNil(t, subject.Values.TPM2EventLogs, "no event logs uncompressed") {
		assert.Len(t, subject.Values.TPM2EventLogs, 5, "wrong number of event logs")
		for _, evLog := range subject.Values.TPM2EventLogs {
			assert.NotEmpty(t, evLog)
		}
	}
}

func TestIssue926(t *testing.T) {
	ctx := context.Background()
	subj1 := parseEvidence(t, "../../test/issue926-x1carbon-ancient.evidence.json")
	res, err := Run(ctx, subj1)
	assert.NoError(t, err)
	assert.True(t, subj1.BaselineModified)
	for _, a := range res.Issues {
		assert.False(t, a.Incident())
	}

	subj2 := parseEvidence(t, "../../test/issue926-x1carbon-before.evidence.json")
	subj2.Baseline = subj1.Baseline
	res, err = Run(ctx, subj2)
	assert.NoError(t, err)
	assert.False(t, subj2.BaselineModified)
	for _, a := range res.Issues {
		assert.False(t, a.Incident())
	}

	subj3 := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	subj3.Baseline = subj1.Baseline
	res, err = Run(ctx, subj3)
	assert.NoError(t, err)
	assert.True(t, subj3.BaselineModified)
	hit := false
	for _, iss := range res.Issues {
		if iss.Id() == issuesv1.UefiBootAppSetId {
			assert.True(t, iss.Incident())
			hit = true
		} else {
			assert.False(t, iss.Incident())
		}
	}
	assert.True(t, hit)
}

func TestMergeVersion(t *testing.T) {
	// mergeVersion should do nothing when we got no evidence
	vBaseline := []int{1, 2, 3, 4}
	change := mergeVersion(&vBaseline, nil, false, false)
	assert.False(t, change, "mergeVersion changes baseline without evidence")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 4}), "unexpected data")

	// mergeVersion should accept evidence when we got no baseline yet
	vEvidence := []int{1, 2, 3, 4}
	tmp := []int{}
	change = mergeVersion(&tmp, vEvidence, false, false)
	assert.True(t, change, "mergeVersion didn't seed baseline with evidence")
	assert.True(t, reflect.DeepEqual(tmp, []int{1, 2, 3, 4}), "unexpected data")

	// reset baseline after change b/c mergeVersion copies the pointer
	vBaseline = []int{1, 2, 3, 4}

	// don't change when both are the same
	change = mergeVersion(&vBaseline, vEvidence, false, false)
	assert.False(t, change, "mergeVersion changes baseline when nothing is requested and nothing changed")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 4}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 4}), "unexpected data")

	// no change when update is set but version hasn't changed
	change = mergeVersion(&vBaseline, vEvidence, true, false)
	assert.False(t, change, "mergeVersion upgraded despite unchanged values")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 4}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 4}), "unexpected data")

	// no change when downgrade is set but version hasn't changed
	change = mergeVersion(&vBaseline, vEvidence, false, true)
	assert.False(t, change, "mergeVersion downgrade despite unchanged values")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 4}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 4}), "unexpected data")

	// update when required
	vEvidence[3] = 99
	change = mergeVersion(&vBaseline, vEvidence, true, false)
	assert.True(t, change, "mergeVersion didn't upgrade baseline with evidence")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 99}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 99}), "unexpected data")
	vBaseline = []int{1, 2, 3, 4}

	// don't change when evidence is newer but upgrade isn't set
	change = mergeVersion(&vBaseline, vEvidence, false, false)
	assert.False(t, change, "mergeVersion changes baseline when no upgrade requested")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 4}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 99}), "unexpected data")

	// downgrade when required
	vEvidence[3] = 0
	change = mergeVersion(&vBaseline, vEvidence, false, true)
	assert.True(t, change, "mergeVersion didn't downgrade baseline with evidence")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 0}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 0}), "unexpected data")
	vBaseline = []int{1, 2, 3, 4}

	// don't change when evidence is older but downgrade isn't set
	change = mergeVersion(&vBaseline, vEvidence, false, false)
	assert.False(t, change, "mergeVersion changes baseline when no downgrade requested")
	assert.True(t, reflect.DeepEqual(vBaseline, []int{1, 2, 3, 4}), "unexpected data")
	assert.True(t, reflect.DeepEqual(vEvidence, []int{1, 2, 3, 0}), "unexpected data")
}

func TestIssue1498(t *testing.T) {
	ctx := context.Background()
	subj1 := parseEvidence(t, "../../test/ludmilla-2.evidence.json")
	res, err := Run(ctx, subj1)
	assert.NoError(t, err)
	assert.False(t, res.SupplyChain)
	assert.False(t, res.EndpointProtection)
}
