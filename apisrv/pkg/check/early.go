package check

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"fmt"
	"io"
	"reflect"

	"github.com/antchfx/xmlquery"
	"github.com/google/go-tpm/tpm2"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	ev "github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

// https://cdn.fwupd.org/downloads/firmware.xml.gz
// Last update: 25th Mar 2022
var (
	//go:embed firmware.xml.gz
	lvfsFirmware []byte
	lvfsPCR0     map[string]string
)

func init() {
	lvfsPCR0 = make(map[string]string)
	rd, err := gzip.NewReader(bytes.NewReader(lvfsFirmware))
	if err != nil {
		panic(err)
	}
	strm, err := xmlquery.CreateStreamParser(rd, "/components/component")
	if err != nil {
		panic(err)
	}
	for {
		c, err := strm.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		nameNode := xmlquery.FindOne(c, "name")
		if nameNode != nil {
			name := nameNode.InnerText()
			for _, pcr0 := range xmlquery.Find(c, "//checksum[@target='device']") {
				lvfsPCR0[pcr0.InnerText()] = name
			}
		}
	}
	rd.Close()
}

type intelBootGuard struct{}

func (intelBootGuard) String() string {
	return "Boot Guard"
}

func UEFIUpdated(ctx context.Context, bline *baseline.Values, evidence *ev.Values) bool {
	ver, date, err := evidence.SMBIOSPlatformVersion(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("decode smbios")
		return false
	}
	verUnchanged := bline.BIOSVersion == "" || reflect.DeepEqual(bline.BIOSVersion, ver)
	dateUnchanged := bline.BIOSReleaseDate == "" || reflect.DeepEqual(bline.BIOSReleaseDate, date)
	return !verUnchanged || !dateUnchanged
}

func (intelBootGuard) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	// IBB measurements changed, expect reported version/date to change
	if subj.Baseline.BootGuardIBB.IntersectsWith(&subj.Boot.BootGuardIBB) {
		return nil
	}
	ver, date, err := subj.Values.SMBIOSPlatformVersion(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("decode smbios -- version")
		return nil
	}
	ven, err := subj.Values.SMBIOSPlatformVendor(ctx)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("decode smbios -- vendor")
		return nil
	}

	if !UEFIUpdated(ctx, subj.Baseline, subj.Values) {
		before, after := baseline.BeforeAfter(&subj.Baseline.BootGuardIBB, &subj.Boot.BootGuardIBB)
		var iss issuesv1.UefiIbbNoUpdate
		iss.Common.Id = issuesv1.UefiIbbNoUpdateId
		iss.Common.Aspect = issuesv1.UefiIbbNoUpdateAspect
		iss.Common.Incident = true
		iss.Args.Before = before
		iss.Args.After = after
		iss.Args.Vendor = ven
		iss.Args.ReleaseDate = date
		iss.Args.Version = ver

		return &iss
	}

	return nil
}

func (intelBootGuard) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.UefiIbbNoUpdateId)
	change := false

	if allowChange {
		change = change || !reflect.DeepEqual(subj.Baseline.BootGuardIBB, subj.Boot.BootGuardIBB)
		subj.Baseline.BootGuardIBB = subj.Boot.BootGuardIBB
	} else {
		change = subj.Baseline.BootGuardIBB.UnionWith(&subj.Boot.BootGuardIBB) || change
	}

	ver, date, err := subj.Values.SMBIOSPlatformVersion(ctx)
	if err == nil {
		if subj.Baseline.BIOSVersion == "" || allowChange {
			change = change || subj.Baseline.BIOSVersion != ver
			subj.Baseline.BIOSVersion = ver
		}
		if subj.Baseline.BIOSReleaseDate == "" || allowChange {
			change = change || subj.Baseline.BIOSReleaseDate != date
			subj.Baseline.BIOSReleaseDate = date
		}
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type lvfs struct{}

func (lvfs) String() string {
	return "LVFS PCR0"
}

func (lvfs) Verify(ctx context.Context, subj *Subject) *api.Annotation {
	var hasSha256 = false
	var hasSha1 = false

	if pcrs, ok := subj.Values.PCR[fmt.Sprintf("%d", tpm2.AlgSHA256)]; ok {
		if pcr0, ok := pcrs["0"]; ok {
			_, ok := lvfsPCR0[fmt.Sprintf("%x", pcr0)]
			hasSha256 = ok
		}
	} else if pcrs, ok := subj.Values.PCR[fmt.Sprintf("%d", tpm2.AlgSHA1)]; ok {
		if pcr0, ok := pcrs["0"]; ok {
			_, ok := lvfsPCR0[fmt.Sprintf("%x", pcr0)]
			hasSha1 = ok
		}
	}

	if !hasSha1 && !hasSha256 && !subj.Baseline.AllowMissingLVFS {
		return api.NewAnnotation(api.AnnTCNotInLVFS)
	}
	return nil
}

func (lvfs) Update(ctx context.Context, overrides []api.AnnotationID, subj *Subject) {
	//if hasAnnotation(api.AnnTCNotInLVFS, overrides) && !subj.Baseline.AllowMissingLVFS {
	//	subj.Baseline.AllowMissingLVFS = true
	//	subj.BaselineModified = true
	//}
}
