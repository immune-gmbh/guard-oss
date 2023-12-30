package check

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/windows"
)

var (
	List = []Check{
		csmeDowngrade{},
		csmeNoUpdate{},
		//csmeVulnerableVersion{},
		intelBootGuard{},
		//lvfs{},
		esetModuleEnabled{},
		esetExcluded{},
		//esetFilesManipulated{},
		imaLog{},
		imaBootAggregate{},
		imaFiles{},
		intelTSCPlatformRegs{},
		intelTSCEndorsementKey{},
		grub{},
		tpmEventLog{},
		dummyTpm{},
		tpmEndorsementCertificate{},
		uefiBootConfig{},
		uefiBootApp{},
		uefiPartitionTable{},
		uefiSecureBootDisabled{},
		uefiSecureBootKeys{},
		uefiDbx{},
		uefiExitBootServices{},
		uefiSeparators{},
		uefiOfficialDbx{},
		uefiEmbeddedFirmware{},
		windowsKernelConfig{},
		lvfsFirmwareUpdateCheck{},
		windowsBootLogQuotes{},
		windowsBootCounter{},
		policyEndpointProtection{},
		policyIntelTSC{},
	}

	ErrNotApplicable = errors.New("not applicable")
)

type Check interface {
	String() string
	Verify(ctx context.Context, subj *Subject) issuesv1.Issue
	Update(ctx context.Context, overrides []string, subj *Subject)
}

type Result struct {
	Issues             []issuesv1.Issue
	SupplyChain        bool
	EndpointProtection bool
}

func init() {
	List = append(List, ListFWHunt...)
}

func Run(ctx context.Context, subj *Subject) (*Result, error) {
	ctx, span := tel.Start(ctx, "check.Run")
	defer span.End()

	issues := []issuesv1.Issue{}

	for _, check := range List {
		if iss := check.Verify(ctx, subj); iss != nil {
			issues = append(issues, iss)
		}
	}

	for _, check := range List {
		check.Update(ctx, nil, subj)
	}

	res := Result{
		Issues:             issues,
		EndpointProtection: hasEndpointProtection(ctx, subj),
		SupplyChain:        hasSupplyChain(ctx, subj),
	}
	return &res, nil
}

func Override(ctx context.Context, overrides []string, subj *Subject) {
	for _, check := range List {
		check.Update(ctx, overrides, subj)
	}
}

func hasEndpointProtection(ctx context.Context, subj *Subject) bool {
	return hasGenericELAMPPL(ctx, subj) || hasLinuxESET(ctx, subj)
}

func hasSupplyChain(ctx context.Context, subj *Subject) bool {
	return hasIntelTSC(ctx, subj)
}

func hasLinuxESET(ctx context.Context, subj *Subject) bool {
	if len(subj.IMALog) == 0 {
		return false
	}

	var moduleHit bool
	var criticalHit = make(map[string]bool)

	for _, path := range esetCriticalBinaries {
		criticalHit[path] = false
	}

outer:
	for _, ev := range subj.IMALog {
		switch ev := ev.(type) {
		case eventlog.ImaNgEvent:
			for _, re := range esetAllBinaries {
				if re.MatchString(ev.Path) {
					if _, ok := criticalHit[ev.Path]; ok {
						criticalHit[ev.Path] = true
					}
					continue outer
				}
			}
			if esetModule.MatchString(ev.Path) {
				moduleHit = true
				continue
			}
		}
	}

	missing := !moduleHit
	for path, hit := range criticalHit {
		if !hit {
			tel.Log(ctx).WithField("path", path).Info("missing eset binary")
			missing = true
		}
	}
	return !missing
}

func hasIntelTSC(ctx context.Context, subj *Subject) bool {
	return subj.IntelTSCData != nil && subj.PlatformCertificates != nil
}

func hasBinarly(ctx context.Context, subj *Subject) bool {
	return subj.BinarlyReport != nil
}

func hasGenericELAMPPL(ctx context.Context, subj *Subject) bool {
	haveLog := len(subj.WindowsLogs) > subj.BootEventLogIdx && subj.WindowsLogs[subj.BootEventLogIdx] != nil
	havePPL := len(subj.AntiMalwareProcesses) > 0
	haveELAM := len(subj.EarlyLaunchDrivers) > 0

	if !haveLog || !havePPL || !haveELAM {
		tel.Log(ctx).WithFields(log.Fields{"log": haveLog, "ppl": havePPL, "elam": haveELAM}).Warn("missing data")
		return false
	}

	// 1st: ELAM driver was loaded
	type driverInfo struct {
		Certs   []windows.ELAMCertificateInfo
		Version *windows.Version
	}
	var loadedElam = make(map[string]driverInfo)
	for path, contents := range subj.EarlyLaunchDrivers {
		modulePath := volumeRe.ReplaceAllString(path, "")
		for _, lm := range subj.WindowsLogs[subj.BootEventLogIdx].LoadedModules {
			if strings.EqualFold(strings.ToLower(modulePath), strings.ToLower(lm.FilePath)) {
				sys, err := windows.Parse(contents)
				if err != nil {
					tel.Log(ctx).WithError(err).Error("parse sys file")
					continue
				}
				if bytes.Equal(sys.Authentihash(), lm.AuthenticodeHash) {
					certs, err := windows.GetELAMCertificateInfo(sys)
					if err != nil {
						tel.Log(ctx).WithError(err).Error("cert info")
						continue
					}
					ver, err := windows.GetVersion(sys)
					if err != nil {
						tel.Log(ctx).WithError(err).Error("pe version")
						continue
					}

					loadedElam[path] = driverInfo{certs, ver}
					break
				}
			}
		}
	}

	// 2nd: Antimalware PPL process was whitelisted by it
outer:
	for ppl, contents := range subj.AntiMalwareProcesses {
		exe, err := windows.Parse(contents)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("parse exe file")
			continue
		}
		signers := exe.Certificates.Content.Certificates

		for _, driver := range loadedElam {
			for _, cert := range driver.Certs {
				for i := range signers {
					cert.Algorithm.Reset()
					cert.Algorithm.Write(signers[i].RawTBSCertificate)
					sum := cert.Algorithm.Sum(nil)

					if bytes.Equal(sum, cert.Hash) {
						tel.Log(ctx).WithField("ver", fmt.Sprintf("%#v", driver.Version)).Info("found epp")
						continue outer
					}
				}
			}
		}

		tel.Log(ctx).WithField("path", ppl).Info("no signer")
		return false
	}

	return true
}
func printVersion(v []int) string {
	var s string
	for i, vv := range v {
		if i > 0 {
			s = fmt.Sprintf("%s.%d", s, vv)
		} else {
			s = fmt.Sprint(vv)
		}
	}
	return s
}

func hasIssue(ii []string, i string) bool {
	idx := sort.SearchStrings(ii, i)
	return idx < len(ii) && ii[idx] == i
}

// a < b
func compareVersions(a []int, b []int) bool {
	if len(a) < len(b) {
		return true
	} else if len(a) > len(b) {
		return false
	}
	for i := 0; i < len(a); i += 1 {
		if a[i] < b[i] {
			return true
		}
	}
	return false
}

// mergeVersion tests and updates baseline versions in a generic way
func mergeVersion(baseline *[]int, evidence []int, upgrade, downgrade bool) bool {
	if len(evidence) == 0 {
		return false
	}

	if len(*baseline) == 0 {
		*baseline = evidence
		return true
	}

	downgraded := compareVersions(evidence, *baseline)
	if downgrade && downgraded {
		*baseline = evidence
		return true
	}

	upgraded := !downgraded && !reflect.DeepEqual(evidence, *baseline)
	if upgraded && upgrade {
		*baseline = evidence
		return true
	}

	return false
}
