package check

import (
	"context"
	"fmt"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

// created with:
// cat fwhunt_report_linear.json | grep Name | cut -d : -f 2 | tr -d " \"," | xargs -I {} bash -c "echo fwHunt{\\\"{}\\\", api.Ann`echo {} | tr -d -`},"
var ListFWHunt = []Check{
	fwHunt{"BRLY-2021-001", issuesv1.Brly2021001},
	fwHunt{"BRLY-2021-003", issuesv1.Brly2021003},
	fwHunt{"BRLY-2021-004", issuesv1.Brly2021004},
	fwHunt{"BRLY-2021-005", issuesv1.Brly2021005},
	fwHunt{"BRLY-2021-006", issuesv1.Brly2021006},
	fwHunt{"BRLY-2021-007", issuesv1.Brly2021007},
	fwHunt{"BRLY-2021-008", issuesv1.Brly2021008},
	fwHunt{"BRLY-2021-009", issuesv1.Brly2021009},
	fwHunt{"BRLY-2021-010", issuesv1.Brly2021010},
	fwHunt{"BRLY-2021-011", issuesv1.Brly2021011},
	fwHunt{"BRLY-2021-012", issuesv1.Brly2021012},
	fwHunt{"BRLY-2021-013", issuesv1.Brly2021013},
	fwHunt{"BRLY-2021-014", issuesv1.Brly2021014},
	fwHunt{"BRLY-2021-015", issuesv1.Brly2021015},
	fwHunt{"BRLY-2021-016", issuesv1.Brly2021016},
	fwHunt{"BRLY-2021-017", issuesv1.Brly2021017},
	fwHunt{"BRLY-2021-018", issuesv1.Brly2021018},
	fwHunt{"BRLY-2021-019", issuesv1.Brly2021019},
	fwHunt{"BRLY-2021-020", issuesv1.Brly2021020},
	fwHunt{"BRLY-2021-021", issuesv1.Brly2021021},
	fwHunt{"BRLY-2021-022", issuesv1.Brly2021022},
	fwHunt{"BRLY-2021-023", issuesv1.Brly2021023},
	fwHunt{"BRLY-2021-024", issuesv1.Brly2021024},
	fwHunt{"BRLY-2021-025", issuesv1.Brly2021025},
	fwHunt{"BRLY-2021-026", issuesv1.Brly2021026},
	fwHunt{"BRLY-2021-027", issuesv1.Brly2021027},
	fwHunt{"BRLY-2021-028", issuesv1.Brly2021028},
	fwHunt{"BRLY-2021-029", issuesv1.Brly2021029},
	fwHunt{"BRLY-2021-030", issuesv1.Brly2021030},
	fwHunt{"BRLY-2021-031", issuesv1.Brly2021031},
	fwHunt{"BRLY-2021-032", issuesv1.Brly2021032},
	fwHunt{"BRLY-2021-033", issuesv1.Brly2021033},
	fwHunt{"BRLY-2021-034", issuesv1.Brly2021034},
	fwHunt{"BRLY-2021-035", issuesv1.Brly2021035},
	fwHunt{"BRLY-2021-036", issuesv1.Brly2021036},
	fwHunt{"BRLY-2021-037", issuesv1.Brly2021037},
	fwHunt{"BRLY-2021-038", issuesv1.Brly2021038},
	fwHunt{"BRLY-2021-039", issuesv1.Brly2021039},
	fwHunt{"BRLY-2021-040", issuesv1.Brly2021040},
	fwHunt{"BRLY-2021-041", issuesv1.Brly2021041},
	fwHunt{"BRLY-2021-042", issuesv1.Brly2021042},
	fwHunt{"BRLY-2021-043", issuesv1.Brly2021043},
	fwHunt{"BRLY-2021-045", issuesv1.Brly2021045},
	fwHunt{"BRLY-2021-046", issuesv1.Brly2021046},
	fwHunt{"BRLY-2021-047", issuesv1.Brly2021047},
	fwHunt{"BRLY-2021-050", issuesv1.Brly2021050},
	fwHunt{"BRLY-2021-051", issuesv1.Brly2021051},
	fwHunt{"BRLY-2021-053", issuesv1.Brly2021053},
	fwHunt{"BRLY-2022-004", issuesv1.Brly2022004},
	fwHunt{"BRLY-2022-009", issuesv1.Brly2022009},
	fwHunt{"BRLY-2022-010", issuesv1.Brly2022010},
	fwHunt{"BRLY-2022-011", issuesv1.Brly2022011},
	fwHunt{"BRLY-2022-012", issuesv1.Brly2022012},
	fwHunt{"BRLY-2022-013", issuesv1.Brly2022013},
	fwHunt{"BRLY-2022-014", issuesv1.Brly2022014},
	fwHunt{"BRLY-2022-015", issuesv1.Brly2022015},
	fwHunt{"BRLY-2022-016", issuesv1.Brly2022016},
	fwHunt{"BRLY-2022-027", issuesv1.Brly2022027},
	fwHunt{"BRLY-2022-028-RsbStuffing", issuesv1.Brly2022028Rsbstuffing},
	fwHunt{"BRLY-ESPecter", issuesv1.BrlyEspecter},
	fwHunt{"BRLY-Intel-BSSA-DFT", issuesv1.BrlyIntelBssaDft},
	fwHunt{"BRLY-Lojax-SecDxe", issuesv1.BrlyLojaxSecdxe},
	fwHunt{"BRLY-MoonBounce-CORE-DXE", issuesv1.BrlyMoonbounceCoreDxe},
	fwHunt{"BRLY-MosaicRegressor", issuesv1.BrlyMosaicregressor},
	fwHunt{"BRLY-Rkloader", issuesv1.BrlyRkloader},
	fwHunt{"BRLY-ThinkPwn", issuesv1.BrlyThinkpwn},
	fwHunt{"BRLY-UsbRt-CVE-2017-5721", issuesv1.BrlyUsbrtCve20175721},
	fwHunt{"BRLY-UsbRt-INTEL-SA-00057", issuesv1.BrlyUsbrtIntelSa00057},
	fwHunt{"BRLY-UsbRt-SwSmi-CVE-2020-12301", issuesv1.BrlyUsbrtSwsmiCve202012301},
	fwHunt{"BRLY-UsbRt-UsbSmi-CVE-2020-12301", issuesv1.BrlyUsbrtUsbsmiCve202012301},
}

type fwHunt struct {
	vulnName string
	annId    string
}

func (fwh fwHunt) String() string {
	return fmt.Sprint("Binarly fwhunt ", fwh.vulnName)
}

func (fwh fwHunt) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	report := subj.BinarlyReport

	if report == nil || report.FWHunt == nil {
		return nil
	}

	var match bool
	for _, result := range report.FWHunt.Results {
		if result.Name == fwh.vulnName && result.Value > 0 {
			match = true
			break
		}
	}

	if !match {
		return nil
	}

	for _, id := range subj.Baseline.AllowBinarlyVulnerabilityIDs {
		if id == fwh.vulnName {
			return nil
		}
	}
	var iss issuesv1.Binarly
	iss.Common.Id = fwh.annId
	iss.Common.Aspect = issuesv1.BinarlyAspect
	iss.Common.Incident = false
	return &iss
}

func (fwh fwHunt) Update(ctx context.Context, overrides []string, subj *Subject) {
	if !hasIssue(overrides, fwh.annId) {
		return
	}

	for _, id := range subj.Baseline.AllowBinarlyVulnerabilityIDs {
		if id == fwh.vulnName {
			return
		}
	}

	subj.Baseline.AllowBinarlyVulnerabilityIDs = append(subj.Baseline.AllowBinarlyVulnerabilityIDs, fwh.vulnName)
	subj.BaselineModified = true
}
