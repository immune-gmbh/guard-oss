package check

import (
	"context"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

// a & b, a - b
func intersectionAndDifference(a, b []string) ([]string, []string) {
	m := make(map[string]bool)

	for _, aa := range a {
		m[aa] = false
	}
	var i, d []string
	for _, bb := range b {
		if hit, ok := m[bb]; ok && !hit {
			i = append(i, bb)
			m[bb] = true
		}
	}
	for aa, hit := range m {
		if !hit {
			d = append(d, aa)
		}
	}

	sort.Strings(i)
	sort.Strings(d)

	return i, d
}

type esetModuleEnabled struct{}

func (esetModuleEnabled) String() string {
	return "Linux ESET kernel module"
}

func (esetModuleEnabled) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Values.ESET == nil {
		return nil
	}

	if !subj.Baseline.AllowDisabledESET && string(subj.Values.ESET.Enabled.Data) != "1\n" {
		var iss issuesv1.EsetDisabled
		iss.Common.Id = issuesv1.EsetDisabledId
		iss.Common.Aspect = issuesv1.EsetDisabledAspect
		iss.Common.Incident = true
		return &iss
	}

	return nil
}

func (esetModuleEnabled) Update(ctx context.Context, overrides []string, subj *Subject) {
	if subj.Values.ESET == nil {
		return
	}

	allowDisabled := hasIssue(overrides, issuesv1.EsetDisabledId)
	if allowDisabled && !subj.Baseline.AllowDisabledESET {
		subj.Baseline.AllowDisabledESET = true
		subj.BaselineModified = true
	}
}

type esetExcluded struct{}

func (esetExcluded) String() string {
	return "Linux ESET excluded list"
}

func (esetExcluded) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Values.ESET == nil {
		return nil
	}

	var iss issuesv1.EsetExcludedSet
	iss.Common.Id = issuesv1.EsetExcludedSetId
	iss.Common.Aspect = issuesv1.EsetExcludedSetAspect
	iss.Common.Incident = true

	if subj.Baseline.ESETExcludedFiles != nil {
		exclFiles := strings.Split(string(subj.Values.ESET.ExcludedFiles.Data), "\n\u0000")
		sort.Strings(exclFiles)
		same, diff := intersectionAndDifference(exclFiles, subj.Baseline.ESETExcludedFiles)
		if !reflect.DeepEqual(exclFiles, same) {
			iss.Args.Files = append(iss.Args.Files, diff...)
		}
	}

	if subj.Baseline.ESETExcludedProcesses != nil {
		exclProcs := strings.Split(string(subj.Values.ESET.ExcludedProcesses.Data), "\n\u0000")
		sort.Strings(exclProcs)
		same, diff := intersectionAndDifference(exclProcs, subj.Baseline.ESETExcludedProcesses)
		if !reflect.DeepEqual(exclProcs, same) {
			iss.Args.Processes = append(iss.Args.Processes, diff...)
		}
	}

	if len(iss.Args.Files) > 0 || len(iss.Args.Processes) > 0 {
		return &iss
	} else {
		return nil
	}
}

func (esetExcluded) Update(ctx context.Context, overrides []string, subj *Subject) {
	if subj.Values.ESET == nil {
		return
	}

	allowExcluded := hasIssue(overrides, issuesv1.EsetExcludedSetId)

	exclFiles := strings.Split(string(subj.Values.ESET.ExcludedFiles.Data), "\n\u0000")
	sort.Strings(exclFiles)
	sameFiles, _ := intersectionAndDifference(exclFiles, subj.Baseline.ESETExcludedFiles)
	if allowExcluded || reflect.DeepEqual(sameFiles, exclFiles) || subj.Baseline.ESETExcludedFiles == nil {
		subj.Baseline.ESETExcludedFiles = exclFiles
		subj.BaselineModified = true
	}

	exclProcs := strings.Split(string(subj.Values.ESET.ExcludedProcesses.Data), "\n\u0000")
	sort.Strings(exclProcs)
	sameProcs, _ := intersectionAndDifference(exclProcs, subj.Baseline.ESETExcludedProcesses)
	if allowExcluded || reflect.DeepEqual(sameProcs, exclProcs) || subj.Baseline.ESETExcludedProcesses == nil {
		subj.Baseline.ESETExcludedProcesses = exclProcs
		subj.BaselineModified = true
	}
}

var (
	esetAllBinaries = []*regexp.Regexp{
		regexp.MustCompile(`^/opt/eset/efs/sbin/.*$`),
		regexp.MustCompile(`^/opt/eset/efs/lib/*\.so(\.\d+)*$`),
		regexp.MustCompile(`^/opt/eset/efs/lib/\w+$`),
		regexp.MustCompile(`^/opt/eset/RemoteAdministrator/Agent/ERAAgent$`),
		regexp.MustCompile(`^/opt/eset/RemoteAdministrator/Agent/\w+\.so(\.\d+)*$`),
	}
	esetCriticalBinaries = []string{
		"/opt/eset/RemoteAdministrator/Agent/ERAAgent",
		"/opt/eset/efs/sbin/startd",
		"/opt/eset/efs/lib/sysinfod",
		"/opt/eset/efs/lib/utild",
		"/opt/eset/efs/lib/oaeventd",
	}
	esetModule *regexp.Regexp = regexp.MustCompile(`^/.*/eset/efs/eset_rtp.ko$`)
)

type esetFilesManipulated struct{}

func (esetFilesManipulated) String() string {
	return "Linux ESET endpoint protection -- files"
}

func (esetFilesManipulated) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if len(subj.IMALog) == 0 {
		return nil
	}

	var moduleHit bool
	var criticalHit = make(map[string]bool)
	var iss issuesv1.EsetManipulated

	iss.Common.Id = issuesv1.EsetManipulatedId
	iss.Common.Aspect = issuesv1.EsetManipulatedAspect
	iss.Common.Incident = true

	for _, path := range esetCriticalBinaries {
		criticalHit[path] = false
	}

outer:
	for _, ev := range subj.IMALog {
		switch ev := ev.(type) {
		case eventlog.ImaNgEvent:
			for _, re := range esetAllBinaries {
				if re.MatchString(ev.Path) {
					h, err := baseline.NewHash(ev.FileDigest)
					if err != nil {
						tel.Log(ctx).WithError(err).Error("parse hash")
						continue outer
					}
					if hh, ok := subj.Baseline.ESETFiles[ev.Path]; ok && !h.IntersectsWith(&hh) {
						before, after := baseline.BeforeAfter(&hh, &h)
						iss.Args.Components = append(iss.Args.Components, issuesv1.EsetManipulatedFile{
							Before: before,
							After:  after,
							Path:   ev.Path,
						})
					}
					if _, ok := criticalHit[ev.Path]; ok {
						criticalHit[ev.Path] = true
					}
					continue outer
				}
			}
			if esetModule.MatchString(ev.Path) {
				h, err := baseline.NewHash(ev.FileDigest)
				if err != nil {
					tel.Log(ctx).WithError(err).Error("parse hash")
					continue
				}
				if !subj.Baseline.ESETKernelModule.IntersectsWith(&h) {
					before, after := baseline.BeforeAfter(&subj.Baseline.ESETKernelModule, &h)
					iss.Args.Components = append(iss.Args.Components, issuesv1.EsetManipulatedFile{
						Before: before,
						After:  after,
						Path:   ev.Path,
					})
				}
				moduleHit = true
				continue
			}
		}
	}

	var missing bool
	if !moduleHit {
		tel.Log(ctx).Error("no eset_rtp.ko")
		missing = true
	}
	for path, hit := range criticalHit {
		if !hit {
			tel.Log(ctx).WithField("path", path).Info("missing eset binary")
			missing = true
		}
	}

	if missing {
		return nil
	} else if len(iss.Args.Components) > 0 {
		return &iss
	} else {
		return nil
	}
}

func (esetFilesManipulated) Update(ctx context.Context, overrides []string, subj *Subject) {
	if len(subj.IMALog) == 0 {
		return
	}

	allowChange := hasIssue(overrides, issuesv1.EsetManipulatedId)
	change := subj.BaselineModified

outer:
	for _, ev := range subj.IMALog {
		switch ev := ev.(type) {
		case eventlog.ImaNgEvent:
			for _, re := range esetAllBinaries {
				if re.MatchString(ev.Path) {
					found, err := baseline.NewHash(ev.FileDigest)
					if err != nil {
						tel.Log(ctx).WithError(err).Error("parse hash")
						continue outer
					}
					if subj.Baseline.ESETFiles == nil {
						subj.Baseline.ESETFiles = make(map[string]baseline.Hash)
					}
					expected := subj.Baseline.ESETFiles[ev.Path]
					if allowChange {
						change = change || !reflect.DeepEqual(found, expected)
						subj.Baseline.ESETFiles[ev.Path] = found
					} else {
						change = expected.UnionWith(&found) || change
						subj.Baseline.ESETFiles[ev.Path] = expected
					}
					continue outer
				}
			}
			if esetModule.MatchString(ev.Path) {
				h, err := baseline.NewHash(ev.FileDigest)
				if err != nil {
					tel.Log(ctx).WithError(err).Error("parse hash")
					continue
				}
				if allowChange {
					change = change || !reflect.DeepEqual(subj.Baseline.ESETKernelModule, h)
					subj.Baseline.ESETKernelModule = h
				} else {
					change = subj.Baseline.ESETKernelModule.UnionWith(&h) || change
				}
				continue
			}
		}
	}

	subj.BaselineModified = change
}
