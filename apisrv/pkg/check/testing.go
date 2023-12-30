package check

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"sort"
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
)

func hasIssueId(issues []issuesv1.Issue, id string) bool {
	for _, iss := range issues {
		if iss.Id() == id {
			return true
		}
	}
	return false
}

func parseValues(t *testing.T, f string, opts ...interface{}) *Subject {
	buf, err := ioutil.ReadFile(f)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var vals evidence.Values
	err = json.Unmarshal(buf, &vals)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	//opts = append(opts, WithBlobs{Blobs: blobs})
	subj, err := NewSubject(context.Background(), &vals, baseline.New(), policy.New(), opts...)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return subj
}

func parseEvidence(t *testing.T, f string, opts ...interface{}) *Subject {
	buf, err := ioutil.ReadFile(f)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]api.Buffer{
			"11": ev.PCRs,
		}
	}

	blobs := make(map[string][]byte)
	if ev.Firmware.EPPInfo != nil {
		for _, v := range ev.Firmware.EPPInfo.EarlyLaunchDrivers {
			if len(v.Data) == 0 {
				cmpr, err := zstd.NewReader(nil)
				assert.NoError(t, err)
				v.Data, err = cmpr.DecodeAll(v.ZData, nil)
				assert.NoError(t, err)
			}
			assert.NotEmpty(t, v.Data)
			sum := sha256.Sum256(v.Data)
			blobs[hex.EncodeToString(sum[:])] = v.Data
		}
		for _, v := range ev.Firmware.EPPInfo.AntimalwareProcesses {
			if len(v.Data) == 0 {
				cmpr, err := zstd.NewReader(nil)
				assert.NoError(t, err)
				v.Data, err = cmpr.DecodeAll(v.ZData, nil)
				assert.NoError(t, err)
			}
			assert.NotEmpty(t, v.Data)
			sum := sha256.Sum256(v.Data)
			blobs[hex.EncodeToString(sum[:])] = v.Data
		}
	}
	if ev.Firmware.BootApps != nil {
		for _, v := range ev.Firmware.BootApps.Images {
			if len(v.Data) == 0 && len(v.ZData) != 0 {
				cmpr, err := zstd.NewReader(nil)
				assert.NoError(t, err)
				v.Data, err = cmpr.DecodeAll(v.ZData, nil)
				assert.NoError(t, err)
			}
			assert.NotEmpty(t, v.Data)
			sum := sha256.Sum256(v.Data)
			blobs[hex.EncodeToString(sum[:])] = v.Data
		}
	}

	val, err := evidence.WrapInsecure(&ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	opts = append(opts, WithBlobs{Blobs: blobs})
	subj, err := NewSubject(context.Background(), val, baseline.New(), policy.New(), opts...)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return subj
}

type modFunc func(subject *Subject)

func modEventlog(t *testing.T, mod func(*eventlog.TPMEvent)) modFunc {
	return func(subj *Subject) {
		ctx := context.Background()
		boot := evidence.EmptyBoot()
		events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
		assert.NoError(t, err)
		events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
		assert.NoError(t, err)
		events := append(events1, events256...)
		sort.Slice(events, func(x, y int) bool {
			return events[x].RawEvent().Sequence < events[y].RawEvent().Sequence
		})

		for _, ev := range events {
			mod(&ev)
			boot.Consume(ctx, ev)
		}
		subj.Boot = *boot
	}
}

func testCheck(t *testing.T, pathBefore, pathAfter string, c Check, id string, mod modFunc) {
	beforesubj := parseEvidence(t, pathBefore)
	aftersubj := parseEvidence(t, pathAfter)
	testCheckImpl(t, beforesubj, aftersubj, c, id, false, mod)
}

func testPure(t *testing.T, pathBefore, pathAfter string, c Check, id string, mod modFunc) {
	beforesubj := parseEvidence(t, pathBefore)
	aftersubj := parseEvidence(t, pathAfter)
	testCheckImpl(t, beforesubj, aftersubj, c, id, true, mod)
}

func testCheckImpl(t *testing.T, beforesubj, aftersubj *Subject, c Check, id string, pure bool, mod modFunc) {
	ctx := context.Background()

	// compare baseline against "before" state => expect no annotations
	iss := c.Verify(ctx, beforesubj)
	assert.Nil(t, iss, "Expected no annotations when checking against fresh baseline")

	// update baseline with "before" state and w/o overrides => expect change if not pure
	c.Update(ctx, []string{}, beforesubj)
	assert.Equal(t, !pure, beforesubj.BaselineModified, "Expected Update()ing the baseline w/o any overrides set *not* changing BaselineModified to true")
	if beforesubj.BaselineModified {
		aftersubj.Baseline = beforesubj.Baseline
	}

	mod(aftersubj)

	// compare "before" against "after" state => expect annotation
	iss = c.Verify(ctx, aftersubj)
	if id != "" {
		assert.NotNil(t, iss, "Expected issue from Verify()")
		assert.Equal(t, id, iss.Id())
	} else {
		assert.Nil(t, iss)
	}

	// update "before" with "after" state, but w/o any overrides
	c.Update(ctx, []string{}, aftersubj)

	if id != "" {
		// verify unchanged baseline => expect same annotation
		iss = c.Verify(ctx, aftersubj)
		assert.NotNil(t, iss, "Expected unchanged baseline to yield same annotation")
		assert.Equal(t, id, iss.Id(), "Expected unchanged baseline to yield same annotation")

		// update baseline with "after" state and override => expect change
		c.Update(ctx, []string{iss.Id()}, aftersubj)
		assert.True(t, aftersubj.BaselineModified)
	}

	// verify updated baseline state => expect no annotation
	iss = c.Verify(ctx, aftersubj)
	assert.Nil(t, iss)
}
