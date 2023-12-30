package check

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/windows"
)

func TestBootDebug(t *testing.T) {
	testPure(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		windowsKernelConfig{},
		issuesv1.WindowsBootConfigId,
		func(subj *Subject) {
			subj.WindowsLogs = []*eventlog.WinEvents{{BootDebuggingEnabled: true}}
		})
}

func TestKernelDebug(t *testing.T) {
	testPure(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		windowsKernelConfig{},
		issuesv1.WindowsBootConfigId,
		func(subj *Subject) {
			subj.WindowsLogs = []*eventlog.WinEvents{{KernelDebugEnabled: true}}
		})
}

func TestNoDEP(t *testing.T) {
	testPure(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		windowsKernelConfig{},
		issuesv1.WindowsBootConfigId,
		func(subj *Subject) {
			subj.WindowsLogs = []*eventlog.WinEvents{{DEPEnabled: eventlog.TernaryFalse}}
		})
}

func TestNotCodeIntegrity(t *testing.T) {
	testPure(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		windowsKernelConfig{},
		issuesv1.WindowsBootConfigId,
		func(subj *Subject) {
			subj.WindowsLogs = []*eventlog.WinEvents{{CodeIntegrityEnabled: eventlog.TernaryFalse}}
		})
}

func TestTestSignign(t *testing.T) {
	testPure(t,
		"../../test/sr630.evidence.json",
		"../../test/sr630.evidence.json",
		windowsKernelConfig{},
		issuesv1.WindowsBootConfigId,
		func(subj *Subject) {
			subj.WindowsLogs = []*eventlog.WinEvents{{TestSigningEnabled: true}}
		})
}

func TestWindowsDefenderELAM(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/r340.evidence.json")
	assert.True(t, hasEndpointProtection(ctx, subj))
}

func TestESETELAM(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/latitude-7520.evidence.json")
	assert.True(t, hasEndpointProtection(ctx, subj))
}

func TestNoELAM(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-auditmode.json")
	assert.False(t, hasEndpointProtection(ctx, subj))
}

func TestNoCertInfo(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/latitude-7520.evidence.json")

	buf := subj.EarlyLaunchDrivers["C:\\WINDOWS\\system32\\DRIVERS\\eelam.sys"]
	buf = bytes.ReplaceAll(buf,
		[]byte("\x00M\x00S\x00E\x00L\x00A\x00M\x00C\x00E\x00R\x00T\x00I\x00N\x00F\x00O\x00I\x00D"),
		[]byte("\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X\x00X"))
	subj.EarlyLaunchDrivers["C:\\WINDOWS\\system32\\DRIVERS\\eelam.sys"] = buf

	sys, err := windows.Parse(buf)
	assert.NoError(t, err)

	mod := subj.WindowsLogs[0].LoadedModules["C:\\WINDOWS\\system32\\DRIVERS\\eelam.sys"]
	mod.AuthenticodeHash = sys.Authentihash()
	subj.WindowsLogs[0].LoadedModules["C:\\WINDOWS\\system32\\DRIVERS\\eelam.sys"] = mod

	assert.False(t, hasGenericELAMPPL(ctx, subj))
}

func TestNoPPL(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/latitude-7520.evidence.json")

	subj.AntiMalwareProcesses = make(map[string][]byte)
	assert.False(t, hasGenericELAMPPL(ctx, subj))
}

func TestPPLNotSigned(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/latitude-7520.evidence.json")

	buf := subj.AntiMalwareProcesses[`C:\Program Files\ESET\ESET Security\ekrn.exe`]
	// certificate subject
	buf = bytes.ReplaceAll(buf, []byte("ESET"), []byte("XXXX"))
	subj.AntiMalwareProcesses[`C:\Program Files\ESET\ESET Security\ekrn.exe`] = buf

	assert.False(t, hasGenericELAMPPL(ctx, subj))
}

func TestWBCLQuote(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootLogQuotesId, string(ann.Id()))
	}

	subj = parseEvidence(t, "../../test/dell-notebook-5eventlogs-WBCL.evidence.json")
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootLogQuotesId, string(ann.Id()))
	}
}

func TestWBCLQuoteNoEventlogs(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	tmp := subj.EventLogs
	subj.EventLogs = nil
	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootLogQuotesId, string(ann.Id()))
	}

	subj.EventLogs = tmp
	tmp2 := subj.WindowsLogs
	subj.WindowsLogs = nil
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootLogQuotesId, string(ann.Id()))
	}

	subj.WindowsLogs = tmp2
	tmp3 := subj.Values.PCPQuoteKeys
	subj.Values.PCPQuoteKeys = nil
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootLogQuotesId, string(ann.Id()))
	}

	subj.Values.PCPQuoteKeys = tmp3
	subj.WindowsLogs[0].TrustPointQuote = nil
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootLogQuotesId, string(ann.Id()))
	}
}

func TestWBCLQuoteSigInvalid(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	tmp := subj.WindowsLogs[0].TrustPointQuote["bla"].QuoteSignature[0]
	subj.WindowsLogs[0].TrustPointQuote["bla"].QuoteSignature[0] = 0x42
	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	var hit bool
	for _, ann := range res.Issues {
		hit = hit || issuesv1.WindowsBootLogQuotesId == ann.Id()
	}
	assert.True(t, hit)

	subj.WindowsLogs[0].TrustPointQuote["bla"].QuoteSignature[0] = tmp
	subj.WindowsLogs[0].TrustPointQuote["bla"].Quote[0] = 0x42
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		hit = hit || issuesv1.WindowsBootLogQuotesId == ann.Id()
	}
	assert.True(t, hit)
}

func TestWBCLQuoteInvalid(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")

	var err error
	subj.Values.TPM2EventLogs[0].Data[400] = 0x42
	subj.EventLogs[0], err = eventlog.ParseEventLog(subj.Values.TPM2EventLogs[0].Data)
	assert.NoError(t, err, "error parsing event logs")

	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	var hit bool
	for _, ann := range res.Issues {
		hit = hit || issuesv1.WindowsBootLogQuotesId == ann.Id()
	}
	assert.True(t, hit)
}

func TestWinBootReplay(t *testing.T) {
	ctx := context.Background()
	subj := parseEvidence(t, "../../test/dell-notebook-5eventlogs-WBCL.evidence.json")
	res, err := Run(ctx, subj)
	assert.NoError(t, err)
	for _, ann := range res.Issues {
		assert.NotEqual(t, issuesv1.WindowsBootCounterReplayId, string(ann.Id()))
	}

	Override(ctx, nil, subj)
	bc, err := strconv.ParseUint(subj.Baseline.BootCount, 16, 64)
	if !assert.NoError(t, err, "error decoding boot count") {
		t.FailNow()
	}
	assert.Equal(t, subj.WindowsLogs[4].BootCount, bc)

	subj.WindowsLogs[3].BootCount = 23
	res, err = Run(ctx, subj)
	assert.NoError(t, err)
	var hit bool
	for _, ann := range res.Issues {
		hit = hit || issuesv1.WindowsBootCounterReplayId == ann.Id()
	}
	assert.True(t, hit)
}
