package check

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/binarly"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func Test_fwHunt_Verify(t *testing.T) {
	buf, err := ioutil.ReadFile("../../test/fwhunt_report_linear_00.json")
	assert.NoError(t, err)

	var br binarly.FWHuntReport
	err = json.Unmarshal(buf, &br)
	assert.NoError(t, err)

	ctx := context.Background()
	subj := Subject{BinarlyReport: &binarly.Report{FWHunt: &br}, Baseline: baseline.New()}

	hit := false
	for _, check := range ListFWHunt {
		ann := check.Verify(ctx, &subj)
		if ann != nil {
			hit = assert.Equal(t, issuesv1.Brly2021011, ann.Id(), "fwhunt check didn't return expected annotation")
		}
	}

	assert.True(t, hit, "fwhunt check didn't produce expected annotation")

	for _, check := range ListFWHunt {
		check.Update(ctx, nil, &subj)
		assert.False(t, subj.BaselineModified, check.String(), " subject BaselineModified unexpectedly true")
	}

	assert.Empty(t, subj.Baseline.AllowBinarlyVulnerabilityIDs, "fwhunt check updated baseline unexpectedly")
}
