package check

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"fmt"
	"regexp"
	"strconv"

	"github.com/google/go-tpm/tpm2"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/windows"
)

type windowsKernelConfig struct{}

func (windowsKernelConfig) String() string {
	return "WBCL kernel configuration"
}

func (windowsKernelConfig) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.WindowsLogs == nil {
		return nil
	}

	for _, winLog := range subj.WindowsLogs {
		if winLog == nil {
			continue
		}
		var unsecure bool
		var iss issuesv1.WindowsBootConfig

		iss.Common.Id = issuesv1.WindowsBootConfigId
		iss.Common.Aspect = issuesv1.WindowsBootConfigAspect
		iss.Common.Incident = false
		iss.Args.BootDebugging = winLog.BootDebuggingEnabled
		iss.Args.KernelDebugging = winLog.KernelDebugEnabled
		iss.Args.DepDisabled = winLog.DEPEnabled == eventlog.TernaryFalse
		iss.Args.CodeIntegrityDisabled = winLog.CodeIntegrityEnabled == eventlog.TernaryFalse
		iss.Args.TestSigning = winLog.TestSigningEnabled

		if winLog.BootDebuggingEnabled && !subj.Baseline.AllowUnsecureWindowsBoot {
			tel.Log(ctx).Info("boot debugging enabled")
			unsecure = true
		}
		if winLog.KernelDebugEnabled && !subj.Baseline.AllowUnsecureWindowsBoot {
			tel.Log(ctx).Info("kernel debugging enabled")
			unsecure = true
		}
		if winLog.DEPEnabled == eventlog.TernaryFalse && !subj.Baseline.AllowUnsecureWindowsBoot {
			tel.Log(ctx).Info("data exec pervention disabled")
			unsecure = true
		}
		if winLog.CodeIntegrityEnabled == eventlog.TernaryFalse && !subj.Baseline.AllowUnsecureWindowsBoot {
			tel.Log(ctx).Info("code integrity disabled")
			unsecure = true
		}
		if winLog.TestSigningEnabled && !subj.Baseline.AllowUnsecureWindowsBoot {
			tel.Log(ctx).Info("test signing enabled")
			unsecure = true
		}

		if unsecure {
			return &iss
		}
	}

	return nil
}

func (windowsKernelConfig) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.WindowsBootConfigId)

	if allow && !subj.Baseline.AllowUnsecureWindowsBoot {
		subj.Baseline.AllowUnsecureWindowsBoot = true
		subj.BaselineModified = true
	}
}

var volumeRe *regexp.Regexp = regexp.MustCompile(`^\w:`)

type windowsBootLogQuotes struct{}

func (windowsBootLogQuotes) String() string {
	return "Multiple WBCL quote verification"
}

func verifyQuoteSignature(attestBlob []byte, signature []byte, quoteKey crypto.PublicKey) bool {
	attestHasher := sha1.New()
	attestHasher.Write(attestBlob)
	attestHash := attestHasher.Sum([]byte{})

	return rsa.VerifyPKCS1v15(quoteKey.(*rsa.PublicKey), crypto.SHA1, attestHash, signature) == nil
}

func (windowsBootLogQuotes) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if len(subj.EventLogs) == 0 {
		return nil
	}
	if len(subj.WindowsLogs) == 0 {
		return nil
	}
	if subj.Values.PCPQuoteKeys == nil {
		return nil
	}
	// if we have only one log and no TrustPoint quote then the regular validation will suffice
	if len(subj.EventLogs) == 1 && (subj.WindowsLogs[0].TrustPointQuote == nil || len(subj.WindowsLogs[0].TrustPointQuote) == 0) {
		return nil
	}
	if subj.Baseline.AllowInvalidEventlog {
		return nil
	}

	var iss issuesv1.WindowsBootLogQuotes
	iss.Common.Id = issuesv1.WindowsBootLogQuotesId
	iss.Common.Aspect = issuesv1.WindowsBootLogQuotesAspect
	iss.Common.Incident = true

	// subject only decodes the boot windows log; decode more if required
	var trustPointQuotes []map[string]eventlog.WinWBCLQuote
	for _, winLog := range subj.WindowsLogs {
		if winLog == nil {
			continue
		}
		if winLog == nil || len(winLog.TrustPointQuote) == 0 {
			iss.Args.Error = "missing-trust-point"
			return &iss
		}
		trustPointQuotes = append(trustPointQuotes, winLog.TrustPointQuote)
	}

	// if we have multiple WBCLs but no TrustPoint quotes then the trust chain is broken
	if len(subj.EventLogs) != len(trustPointQuotes) {
		iss.Args.Error = "missing-trust-point"
		return &iss
	}

	// verify signatures of all quotes of all eventlogs and collect quotes
	for i, trustPointQuotesInEvLog := range trustPointQuotes {
		var quotes []*tpm2.AttestationData

		iss.Args.Log = int64(i)
		for key, quote := range trustPointQuotesInEvLog {
			attData, err := tpm2.DecodeAttestationData(quote.Quote)
			if err != nil {
				tel.Log(ctx).WithField("error", err).WithField("i", i).Warn("decode quote")
				iss.Args.Error = "wrong-format"
				return &iss
			}
			quotes = append(quotes, attData)

			pubBlob, ok := subj.Values.PCPQuoteKeys[key]
			if !ok {
				tel.Log(ctx).WithField("error", err).WithField("i", i).Warn("decode quote key")
				iss.Args.Error = "wrong-format"
				return &iss
			}

			pub, err := windows.ExtractTPMTPublic(pubBlob)
			if err != nil {
				tel.Log(ctx).WithField("error", err).WithField("i", i).Warn("extract pub key")
				iss.Args.Error = "wrong-format"
				return &iss
			}

			cryptoPub, err := pub.Key()
			if err != nil {
				tel.Log(ctx).WithField("error", err).WithField("i", i).Warn("check signature")
				iss.Args.Error = "wrong-format"
				return &iss
			}

			valid := verifyQuoteSignature(quote.Quote, quote.QuoteSignature, cryptoPub)
			if !valid {
				tel.Log(ctx).WithField("error", err).WithField("i", i).Warn("decode quote")
				iss.Args.Error = "wrong-signature"
				return &iss
			}
		}

		// verify quotes
		for i, quote := range quotes {
			iss.Args.Log = int64(i)

			var pcrSels []eventlog.PCR
			for _, pcrSel := range quote.AttestedQuoteInfo.PCRSelection {
				hash, err := pcrSel.Hash.Hash()
				if err != nil {
					tel.Log(ctx).WithError(err).Error("create hash")
					continue
				}
				for _, pcrIndex := range pcrSel.PCRs {
					pcr := eventlog.PCR{Index: pcrIndex, DigestAlg: hash, Digest: make([]byte, hash.Size())}
					pcrSels = append(pcrSels, pcr)
				}
			}

			// replay eventlog against PCRs selected in quote
			// ignore the error b/c we don't have input hashes that would verify
			replayedPcrs, _ := subj.EventLogs[i].Replay(pcrSels)

			// hash the PCR values and verify against quote
			pcrHasher := sha1.New()
			for i, replayedPcr := range replayedPcrs {
				if replayedPcr == nil {
					replayedPcr = make([]byte, pcrSels[i].DigestAlg.Size())
				}
				pcrHasher.Write(replayedPcr)
			}
			pcrHashSum := pcrHasher.Sum([]byte{})
			if !bytes.Equal(pcrHashSum, []byte(quote.AttestedQuoteInfo.PCRDigest)) {
				iss.Args.Error = "wrong-quote"
				return &iss
			}
		}
	}

	return nil
}

func (windowsBootLogQuotes) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.WindowsBootLogQuotesId)

	if allow && !subj.Baseline.AllowInvalidEventlog {
		subj.Baseline.AllowInvalidEventlog = true
		subj.BaselineModified = true
	}
}

type windowsBootCounter struct{}

func (windowsBootCounter) String() string {
	return "Ensure boot counter is monotonic"
}

func (windowsBootCounter) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.WindowsLogs == nil {
		return nil
	}
	if len(subj.Baseline.BootCount) == 0 {
		return nil
	}

	for _, winLog := range subj.WindowsLogs {
		if winLog == nil {
			continue
		}
		oldCount, err := strconv.ParseUint(subj.Baseline.BootCount, 16, 64)
		if err != nil {
			return nil
		}
		if winLog.BootCount < oldCount {
			var iss issuesv1.WindowsBootCounterReplay

			iss.Common.Id = issuesv1.WindowsBootCounterReplayId
			iss.Common.Aspect = issuesv1.WindowsBootCounterReplayAspect
			iss.Common.Incident = true
			iss.Args.Latest = fmt.Sprint(oldCount)
			iss.Args.Received = fmt.Sprint(winLog.BootCount)

			return &iss
		}
	}

	return nil
}

func (windowsBootCounter) Update(ctx context.Context, overrides []string, subj *Subject) {
	var oldCount uint64
	if len(subj.Baseline.BootCount) == 0 {
		oldCount = 0
	} else {
		var err error
		oldCount, err = strconv.ParseUint(subj.Baseline.BootCount, 16, 64)
		if err != nil {
			return
		}
	}

	var maxBoot uint64
	update := false
	for _, winLog := range subj.WindowsLogs {
		if winLog == nil {
			continue
		}
		if winLog.BootCount > oldCount {
			maxBoot = winLog.BootCount
			update = true
		} else {
			update = false
		}
	}

	if !update {
		update = hasIssue(overrides, issuesv1.WindowsBootCounterReplayId)
	}

	// boot count of 0 is not possible, as a counter starts with zero and each boot increments,
	// so the first boot must be 1
	if update && maxBoot > 0 {
		subj.Baseline.BootCount = strconv.FormatUint(maxBoot, 16)
		subj.BaselineModified = true
	}
}
