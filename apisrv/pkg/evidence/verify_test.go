package evidence

import (
	"context"
	"testing"

	"github.com/google/go-tpm/tpm2"
)

func Test_verifyEvidenceData(t *testing.T) {
	ctx := context.Background()
	lastTimestamp := uint64(0)
	lastFirmware := uint64(0)

	ev := parseEvidence(t, "../../test/ludmilla.evidence.json")
	quoteKeyQName := tpm2.Name(ev.Quote.QualifiedSigner)

	if err := verifyEvidenceData(ctx, ev, quoteKeyQName, lastTimestamp, lastFirmware); err != nil {
		t.Errorf("verifyEvidenceData() error = %v", err)
	}

	ev = parseEvidence(t, "../../test/IMN-DELL.evidence.json")
	quoteKeyQName = tpm2.Name(ev.Quote.QualifiedSigner)

	if err := verifyEvidenceData(ctx, ev, quoteKeyQName, lastTimestamp, lastFirmware); err != nil {
		t.Errorf("verifyEvidenceData() error = %v", err)
	}
}
