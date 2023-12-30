package check

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func TestESETConfigGood(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/sr630.evidence.json")
	assert.True(t, hasEndpointProtection(ctx, subj))
}

func TestESETConfigDisable(t *testing.T) {
	testPure(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetModuleEnabled{},
		issuesv1.EsetDisabledId,
		func(subj *Subject) {
			subj.Values.ESET.Enabled.Data = []byte("0\n")
		})
}

func TestESETConfigExcludeMoreProcs(t *testing.T) {
	testCheck(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetExcluded{},
		issuesv1.EsetExcludedSetId,
		func(subj *Subject) {
			subj.Values.ESET.ExcludedProcesses.Data = append(subj.Values.ESET.ExcludedProcesses.Data, []byte("/blah\n\u0000/blub/\n\u0000")...)
		})
}

func TestESETConfigExcludeOtherProcs(t *testing.T) {
	testCheck(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetExcluded{},
		issuesv1.EsetExcludedSetId,
		func(subj *Subject) {
			subj.Values.ESET.ExcludedProcesses.Data = []byte("/blah\n\u0000/blub/\n\u0000")
		})
}

func TestESETConfigExcludeLessProcs(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/sr630.evidence.json")

	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	for _, a := range res.Issues {
		assert.NotContains(t, a.Id(), "eset/")
	}
	subj.Values.ESET.ExcludedProcesses.Data = []byte("/opt/eset/esets/\n\u0000/usr/bin/encfs\n\u0000")
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, a := range res.Issues {
		assert.NotContains(t, a.Id(), "eset/")
	}
}

func TestESETConfigExcludeMoreFiles(t *testing.T) {
	testCheck(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetExcluded{},
		issuesv1.EsetExcludedSetId,
		func(subj *Subject) {
			subj.Values.ESET.ExcludedFiles.Data = append(subj.Values.ESET.ExcludedFiles.Data, []byte("/blah\n\u0000/blub/\n\u0000")...)
		})
}

func TestESETConfigExcludeOtherFiles(t *testing.T) {
	testCheck(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetExcluded{},
		issuesv1.EsetExcludedSetId,
		func(subj *Subject) {
			subj.Values.ESET.ExcludedFiles.Data = []byte("/blah\n\u0000/blub/\n\u0000")
		})
}

func TestESETConfigExcludeLessFiles(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/sr630.evidence.json")

	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	for _, a := range res.Issues {
		assert.NotContains(t, a.Id(), "eset/")
	}
	subj.Values.ESET.ExcludedFiles.Data = []byte("/dev/\n\u0000/proc/\n\u0000")
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, a := range res.Issues {
		assert.NotContains(t, a.Id(), "eset/")
	}
}

func TestESETMissingModule(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/sr630.evidence.json")

	imaLog := make([]eventlog.TPMEvent, 0)
	for i := range subj.IMALog {
		switch ev := subj.IMALog[i].(type) {
		case eventlog.ImaNgEvent:
			if strings.HasSuffix(ev.Path, "eset_rtp.ko") {
				continue
			}
		}
		imaLog = append(imaLog, subj.IMALog[i])
	}
	subj.IMALog = imaLog
	assert.False(t, hasLinuxESET(ctx, subj))
}

func TestESETChangedModule(t *testing.T) {
	testCheck(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetFilesManipulated{},
		issuesv1.EsetManipulatedId,
		func(subj *Subject) {
			imaLog := make([]eventlog.TPMEvent, 0)
			for i := range subj.IMALog {
				switch ev := subj.IMALog[i].(type) {
				case eventlog.ImaNgEvent:
					if strings.HasSuffix(ev.Path, "eset_rtp.ko") {
						ev.FileDigest[0] = 0xff
						imaLog = append(imaLog, ev)
						continue
					}
				}
				imaLog = append(imaLog, subj.IMALog[i])
			}
			subj.IMALog = imaLog
		})
}

func TestESETMissingFiles(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/sr630.evidence.json")

	imaLog := make([]eventlog.TPMEvent, 0)
	for i := range subj.IMALog {
		switch ev := subj.IMALog[i].(type) {
		case eventlog.ImaNgEvent:
			if strings.HasSuffix(ev.Path, "ERAAgent") {
				continue
			}
		}
		imaLog = append(imaLog, subj.IMALog[i])
	}
	subj.IMALog = imaLog
	assert.False(t, hasLinuxESET(ctx, subj))
}

func TestESETChangedFiles(t *testing.T) {
	testCheck(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		esetFilesManipulated{},
		issuesv1.EsetManipulatedId,
		func(subj *Subject) {
			imaLog := make([]eventlog.TPMEvent, 0)
			for i := range subj.IMALog {
				switch ev := subj.IMALog[i].(type) {
				case eventlog.ImaNgEvent:
					if strings.HasSuffix(ev.Path, "ERAAgent") {
						ev.FileDigest[0] = 0xff
						imaLog = append(imaLog, ev)
						continue
					}
				}
				imaLog = append(imaLog, subj.IMALog[i])
			}
			subj.IMALog = imaLog
		})
}

func intersection(a, b []string) []string {
	r, _ := intersectionAndDifference(a, b)
	return r
}

func TestSetIntersection(t *testing.T) {
	assert.Empty(t, intersection(nil, nil))
	assert.Empty(t, intersection([]string{}, nil))
	assert.Empty(t, intersection(nil, []string{}))
	assert.Empty(t, intersection([]string{}, []string{}))

	assert.Empty(t, intersection(nil, []string{"a"}))
	assert.Empty(t, intersection([]string{"a"}, nil))

	assert.Equal(t, []string{"a"}, intersection([]string{"a"}, []string{"a"}))
	assert.Equal(t, []string{"a"}, intersection([]string{"a", "a", "a"}, []string{"a"}))
	assert.Equal(t, []string{"a"}, intersection([]string{"a"}, []string{"a", "a", "a"}))
	assert.Equal(t, []string{"a"}, intersection([]string{"a"}, []string{"a"}))

	assert.Equal(t, []string{"a", "b"}, intersection([]string{"a", "a", "b"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a", "b"}, intersection([]string{"a", "b"}, []string{"a", "a", "b"}))

	assert.Equal(t, []string{"a", "b", "c"}, intersection([]string{"a", "b", "c"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{"a", "b", "c"}, intersection([]string{"1", "a", "b", "c"}, []string{"a", "b", "c", "z"}))
	assert.Equal(t, []string{"a", "b", "c"}, intersection([]string{"1", "a", "b", "c"}, []string{"a", "b", "c", "z", "z"}))
	assert.Equal(t, []string{"a", "b", "c"}, intersection([]string{"a", "b", "c", "z", "z"}, []string{"1", "a", "b", "c"}))
	assert.Equal(t, []string{"a", "b", "z"}, intersection([]string{"a", "b", "c", "z", "z"}, []string{"1", "a", "b", "z"}))

	assert.Equal(t, []string{"a", "z"}, intersection([]string{"a", "z"}, []string{"1", "a", "b", "z"}))
	assert.Equal(t, []string{"a", "b"}, intersection([]string{"a", "b", "x"}, []string{"a", "b", "z"}))

	assert.Empty(t, intersection([]string{"1", "a", "b", "c"}, []string{"d", "e", "f", "z", "z"}))
}
