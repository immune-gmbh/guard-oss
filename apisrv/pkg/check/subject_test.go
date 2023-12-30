package check

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
)

func TestNoBlobs(t *testing.T) {
	ctx := context.Background()
	ss := parseEvidence(t, "../../test/r340.evidence.json")
	subj, err := NewSubject(ctx, ss.Values, baseline.New(), policy.New(), WithBlobs{})
	assert.NoError(t, err)
	assert.NotNil(t, subj)

	assert.Empty(t, subj.EarlyLaunchDrivers)
	assert.Empty(t, subj.AntiMalwareProcesses)
	assert.NotEmpty(t, ss.Values.EarlyLaunchDrivers)
	assert.NotEmpty(t, ss.Values.AntiMalwareProcesses)
}

func TestMissingBlobs(t *testing.T) {
	ctx := context.Background()
	ss := parseEvidence(t, "../../test/r340.evidence.json")
	blobs := make(map[string][]byte)

	for _, v := range ss.Values.AntiMalwareProcesses {
		blobs[hex.EncodeToString(v)] = make([]byte, 100)
		break
	}
	for _, v := range ss.Values.EarlyLaunchDrivers {
		blobs[hex.EncodeToString(v)] = make([]byte, 100)
		break
	}

	fmt.Println(blobs)
	subj, err := NewSubject(ctx, ss.Values, baseline.New(), policy.New(), WithBlobs{Blobs: blobs})
	assert.NoError(t, err)
	assert.NotNil(t, subj)

	assert.Len(t, subj.EarlyLaunchDrivers, 1)
	assert.Len(t, subj.AntiMalwareProcesses, 1)
	assert.NotEmpty(t, ss.Values.EarlyLaunchDrivers)
	assert.NotEmpty(t, ss.Values.AntiMalwareProcesses)
}
