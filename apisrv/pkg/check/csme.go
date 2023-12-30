package check

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	ev "github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/intelme"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type csmeDowngrade struct{}

func (csmeDowngrade) String() string {
	return "CSME downgrade attack"
}

func csmeVersionCheck(before []int, after []int) (unchanged, downgraded bool) {
	unchanged = reflect.DeepEqual(before, after)
	downgraded = len(before) > 0 && compareVersions(after, before)
	return
}

func (csmeDowngrade) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	var iss issuesv1.CsmeDowngrade
	iss.Common.Id = issuesv1.CsmeDowngradeId
	iss.Common.Aspect = issuesv1.CsmeDowngradeAspect
	iss.Common.Incident = true

	ver, rec, fitc, err := subj.Values.CSMEVersions(ctx)
	if err == nil {
		_, verDowngrade := csmeVersionCheck(subj.Baseline.CSMEVersion, ver)
		_, fitcDowngrade := csmeVersionCheck(subj.Baseline.CSMEFITC, fitc)
		_, recDowngrade := csmeVersionCheck(subj.Baseline.CSMERecovery, rec)

		if verDowngrade || fitcDowngrade || recDowngrade {
			tel.Log(ctx).Info("csme downgraded")
			iss.Args.Combined = &issuesv1.CsmeDowngradeComponent{
				Before: printVersion(subj.Baseline.CSMEVersion),
				After:  printVersion(ver),
			}
		}
	}

	for key, val := range subj.Boot.CSMEComponentVersions {
		if subj.Baseline.CSMEComponentVersion == nil {
			continue
		}
		prev, prevok := subj.Baseline.CSMEComponentVersion[int(key)]
		if !prevok {
			continue
		}

		val := []int{int(val.Version[0]), int(val.Version[1]), int(val.Version[2]), int(val.Version[3])}
		_, downgraded := csmeVersionCheck(prev, val)
		if downgraded {
			tel.Log(ctx).WithFields(log.Fields{
				"component": intelme.MeasuredEntityToString(0, key),
				"before":    printVersion(prev),
				"now":       printVersion(ver),
			}).Info("csme component downgraded")
			iss.Args.Components = append(iss.Args.Components, &issuesv1.CsmeDowngradeComponent{
				Name:   intelme.MeasuredEntityToString(0, key),
				Before: printVersion(prev),
				After:  printVersion(ver),
			})
			continue
		}
	}

	if len(iss.Args.Components) > 0 || iss.Args.Combined != nil {
		return &iss
	} else {
		return nil
	}
}

func (csmeDowngrade) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowDowngrade := hasIssue(overrides, issuesv1.CsmeDowngradeId)
	runtimeChange, change := false, false

	// CSME runtime reported version changed
	ver, rec, fitc, err := subj.Values.CSMEVersions(ctx)
	if err == nil {
		runtimeChange = mergeVersion(&subj.Baseline.CSMEVersion, ver, true, allowDowngrade)
		runtimeChange = mergeVersion(&subj.Baseline.CSMEFITC, fitc, true, allowDowngrade) || runtimeChange
		runtimeChange = mergeVersion(&subj.Baseline.CSMERecovery, rec, true, allowDowngrade) || runtimeChange
	}

	// component version and measurement
	for component, manifest := range subj.Boot.CSMEComponentVersions {
		if subj.Baseline.CSMEComponentVersion == nil {
			subj.Baseline.CSMEComponentVersion = make(map[int][]int)
			subj.Baseline.CSMEComponentHash = make(map[int]api.Buffer)
			change = true
		}

		vEvidence := []int{
			int(manifest.Version[0]),
			int(manifest.Version[1]),
			int(manifest.Version[2]),
			int(manifest.Version[3]),
		}

		// test if version changed in allowed ways and if so, update baseline as is appropriate
		vBase, ok := subj.Baseline.CSMEComponentVersion[int(component)]
		if !ok {
			vBase = make([]int, 0)
		}
		if versionChange := mergeVersion(&vBase, vEvidence, true, allowDowngrade); versionChange {
			subj.Baseline.CSMEComponentVersion[int(component)] = vBase
			change = true
		}
	}

	subj.BaselineModified = subj.BaselineModified || change || runtimeChange
}

type csmeNoUpdate struct{}

func (csmeNoUpdate) String() string {
	return "CSME runtime measurments"
}

func (csmeNoUpdate) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	var csmeUnchanged *bool

	var iss issuesv1.CsmeNoUpdate
	iss.Common.Id = issuesv1.CsmeNoUpdateId
	iss.Common.Aspect = issuesv1.CsmeNoUpdateAspect
	iss.Common.Incident = true

	ver, rec, fitc, err := subj.Values.CSMEVersions(ctx)
	if err == nil {
		verUnchanged, _ := csmeVersionCheck(subj.Baseline.CSMEVersion, ver)
		fitcUnchanged, _ := csmeVersionCheck(subj.Baseline.CSMEFITC, fitc)
		recUnchanged, _ := csmeVersionCheck(subj.Baseline.CSMERecovery, rec)

		unchanged := verUnchanged && fitcUnchanged && recUnchanged
		csmeUnchanged = &unchanged
	}

	// test for consistency (inconsistent sets can't really occur in
	// real boots and point to programming errors)
	for key := range subj.Boot.CSMEComponentHash {
		if _, ok := subj.Boot.CSMEComponentHash[key]; !ok {
			tel.Log(ctx).Errorf("CSME component hash %d has no corresponding version", key)
		}
	}

	for key, val := range subj.Boot.CSMEComponentVersions {
		if subj.Baseline.CSMEComponentVersion == nil {
			continue
		}
		prev, prevok := subj.Baseline.CSMEComponentVersion[int(key)]
		if !prevok {
			continue
		}

		val := []int{int(val.Version[0]), int(val.Version[1]), int(val.Version[2]), int(val.Version[3])}
		unchanged, _ := csmeVersionCheck(prev, val)

		if subj.Baseline.CSMEComponentHash == nil {
			continue
		}
		prevhash, prevhashok := subj.Baseline.CSMEComponentHash[int(key)]
		if !prevhashok {
			continue
		}
		hash, hashok := subj.Boot.CSMEComponentHash[key]
		if !hashok {
			continue
		}
		if bytes.Equal(prevhash, hash) {
			continue
		}

		comp := issuesv1.CsmeNoUpdateComponent{
			Name:    intelme.MeasuredEntityToString(0, key),
			Before:  fmt.Sprintf("%x", prevhash),
			After:   fmt.Sprintf("%x", hash),
			Version: printVersion(val),
		}

		if csmeUnchanged != nil && *csmeUnchanged {
			tel.Log(ctx).WithField("component", intelme.MeasuredEntityToString(0, key)).
				Info("component changed w/o csme update")
			iss.Args.Components = append(iss.Args.Components, comp)
			continue
		}

		if unchanged {
			tel.Log(ctx).WithField("component", intelme.MeasuredEntityToString(0, key)).
				Info("component changed w/o component update")
			iss.Args.Components = append(iss.Args.Components, comp)
		}
	}

	if len(iss.Args.Components) > 0 {
		return &iss
	} else {
		return nil
	}
}

func (csmeNoUpdate) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowNoUpgrade := hasIssue(overrides, issuesv1.CsmeNoUpdateId)
	runtimeChange, change := false, false

	// CSME runtime reported version changed
	ver, rec, fitc, err := subj.Values.CSMEVersions(ctx)
	if err == nil {
		runtimeChange = mergeVersion(&subj.Baseline.CSMEVersion, ver, true, false)
		runtimeChange = mergeVersion(&subj.Baseline.CSMEFITC, fitc, true, false) || runtimeChange
		runtimeChange = mergeVersion(&subj.Baseline.CSMERecovery, rec, true, false) || runtimeChange
	}

	// component version and measurement
	for component, manifest := range subj.Boot.CSMEComponentVersions {
		if subj.Baseline.CSMEComponentVersion == nil {
			subj.Baseline.CSMEComponentVersion = make(map[int][]int)
			subj.Baseline.CSMEComponentHash = make(map[int]api.Buffer)
			change = true
		}

		vEvidence := []int{
			int(manifest.Version[0]),
			int(manifest.Version[1]),
			int(manifest.Version[2]),
			int(manifest.Version[3]),
		}

		hashBase, hashBaseFound := subj.Baseline.CSMEComponentHash[int(component)]
		hashEvidence, hashEvidenceFound := subj.Boot.CSMEComponentHash[component]

		// if we have no baseline hash then accept evidence if any
		updateHash := !hashBaseFound && hashEvidenceFound

		// when present, compare both hashes and update differing and the flag is set that says we may change hashes w/o version updates
		// really do compare hashes here to set updateHash (and thus change) only when there was a change to avoid unnecessary baseline writes
		if hashBaseFound && hashEvidenceFound && (allowNoUpgrade || runtimeChange) && !reflect.DeepEqual(hashBase, hashEvidence) {
			updateHash = true
		}

		// test if version changed in allowed ways and if so, update baseline as is appropriate
		vBase, ok := subj.Baseline.CSMEComponentVersion[int(component)]
		if !ok {
			vBase = make([]int, 0)
		}
		if versionChange := mergeVersion(&vBase, vEvidence, true, false); versionChange {
			subj.Baseline.CSMEComponentVersion[int(component)] = vBase
			change = true

			// only update hash if there is actually a hash to update to
			updateHash = hashEvidenceFound
		}

		if updateHash {
			subj.Baseline.CSMEComponentHash[int(component)] = api.Buffer(hashEvidence)
			change = true
		}
	}

	subj.BaselineModified = subj.BaselineModified || change || runtimeChange
}

type csmeVulnerableVersion struct{}

func (csmeVulnerableVersion) String() string {
	return "CSME vulnerable version"
}

func (csmeVulnerableVersion) Verify(ctx context.Context, subj *Subject) *api.Annotation {
	// Based on Intel CSME Version Detection Tool 6.0.1.0 (6/8/2021) Intel #19392
	ver, _, _, err := subj.Values.CSMEVersions(ctx)
	if err == ev.ErrNotFound {
		return nil
	}
	if err == ev.ErrNoResponse {
		tel.Log(ctx).Warn("CSME didn't answer")
		return nil
	}
	if err != nil {
		tel.Log(ctx).WithError(err).Error("csme version")
		return nil
	}

	vuln := true
	maj := ver[0]
	min := ver[1]
	patch := ver[2]

	tel.Log(ctx).WithFields(log.Fields{"maj": maj, "min": min, "patch": patch}).Info("check csme version")

	// TXE
	if maj == 3 && min == 1 && patch >= 86 {
		vuln = false
	} else if maj == 4 && min == 0 && patch >= 32 {
		vuln = false
	} else if maj == 5 {
		vuln = false
	}

	// CSME
	if maj == 11 && min == 8 && patch >= 86 {
		vuln = false
	} else if maj == 11 && min == 12 && patch >= 86 {
		vuln = false
	} else if maj == 11 && min == 22 && patch >= 86 {
		vuln = false
	} else if maj == 12 && min == 0 && patch >= 81 {
		vuln = false
	} else if maj == 13 && min == 0 && patch >= 47 {
		vuln = false
	} else if maj == 13 && min == 30 && patch >= 17 {
		vuln = false
	} else if maj == 13 && min == 50 && patch >= 11 {
		vuln = false
	} else if maj == 14 && min == 1 && patch >= 53 {
		vuln = false
	} else if maj == 14 && min == 5 && patch >= 32 {
		vuln = false
	} else if maj == 15 && min == 0 && patch >= 22 {
		vuln = false
	} else if maj == 16 {
		vuln = false
	}

	// SPS
	if maj == 1 || maj == 2 {
		vuln = false
	}

	// Unknown
	if maj == 0 {
		vuln = false
	}

	if vuln && !subj.Baseline.AllowVulnerableCSME {
		return api.NewAnnotation(api.AnnTCCSMEVersionVuln)
	} else {
		return nil
	}
}

//func (csmeVulnerableVersion) Update(ctx context.Context, overrides []string, subj *Subject) {
//	if hasIssue(overrides, issuesv1.CsmeVulnerableVersionId) && !subj.Baseline.AllowVulnerableCSME {
//		subj.Baseline.AllowVulnerableCSME = true
//		subj.BaselineModified = true
//	}
//}
