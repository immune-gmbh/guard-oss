package evidence

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"reflect"
	"strconv"

	"github.com/google/go-tpm/tpm2"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrQuote  = errors.New("quote invalid")
	ErrFormat = errors.New("format invalid")
)

func Validate(ctx context.Context, in *api.Evidence, config *api.Configuration, rootQN api.Name, aik *api.PublicKey /*, validNonces[]api.Buffer*/) error {
	// extract AIK
	key, err := tpm2.Public(*aik).Key()
	if err != nil {
		tel.Log(ctx).WithError(err).Error("extract aik")
		return ErrFormat
	}
	// verify it matches the root QN
	qname, err := api.ComputeName(rootQN, *aik)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("compute root qn")
		return ErrFormat
	}
	// check attestation structure fields
	err = verifyEvidenceData(ctx, in, tpm2.Name(qname), 0, 0)
	if err != nil {
		return err
	}
	// check signature with AIK
	err = verifyQuoteSignature(ctx, tpm2.AttestationData(in.Quote), &in.Signature, tpm2.Name(qname), key)
	if err != nil {
		return err
	}

	return nil
}

func verifyEvidenceData(ctx context.Context, ev *api.Evidence, quoteKeyQName tpm2.Name, lastTimestamp uint64, lastFirmware uint64) error {
	ctx, span := tel.Start(ctx, "evidence.verifyEvidenceData")
	defer span.End()

	// compute fw properties hash
	fwPropsHashes, err := EvidenceHashes(ctx, ev)
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("firmware", ev.Firmware).Error("firmware props hash")
		return ErrFormat
	}
	// Verify TPMS_ATTEST and TPMS_SIGNATURE structures against quote key
	attest := tpm2.AttestationData(ev.Quote)
	alg, err := strconv.ParseUint(ev.Algorithm, 0, 16)
	if err != nil {
		tel.Log(ctx).WithError(err).WithField("algorithm", ev.Algorithm).Error("select algorithm")
		return ErrFormat
	}
	// use AllPCRs field if supplied by client, else fallback to old behaviour
	var digests map[string]map[string]api.Buffer
	if ev.AllPCRs != nil && len(ev.AllPCRs) > 0 {
		digests = ev.AllPCRs

		// sanitize digests b/c the TPM wire format doesn't encode PCR selections with zero PCRs in it
		// but agent v3.2.0 sends digests with only a hash alg specified but no actual hashes
		for k, v := range digests {
			if (v == nil) || (len(v) == 0) {
				delete(digests, k)
			}
		}
	} else {
		// if we have an empty PCR list then null digests b/c the TPM wire format doesn't encode
		// PCR selections with zero PCRs in it but agents prior to v3.2.0
		if (ev.PCRs == nil) || (len(ev.PCRs) == 0) {
			digests = nil
		} else {
			digests = make(map[string]map[string]api.Buffer)
			digests[strconv.FormatUint(alg, 10)] = ev.PCRs
		}
	}

	return verifyQuoteFields(ctx, attest, digests, fwPropsHashes, quoteKeyQName, lastTimestamp, lastFirmware)
}

func verifyQuoteFields(ctx context.Context, attest tpm2.AttestationData, digests map[string]map[string]api.Buffer, fwPropsHashes [][]byte, quoteKeyQName tpm2.Name, lastTimestamp uint64, lastFirmware uint64) error {
	ctx, span := tel.Start(ctx, "evidence.verifyQuoteFields")
	defer span.End()

	if !api.EqualNames((*api.Name)(&attest.QualifiedSigner), (*api.Name)(&quoteKeyQName)) {
		tel.Log(ctx).
			WithField("expected", (*api.Name)(&quoteKeyQName).String()).
			WithField("received", (*api.Name)(&attest.QualifiedSigner).String()).
			Error("Wrong signer")
		return ErrQuote
	}

	if attest.AttestedQuoteInfo == nil {
		tel.Log(ctx).Error("not a quote")
		return ErrQuote
	}

	info := attest.AttestedQuoteInfo

	// when the quote contained empty PCR selections then the TPM library decodes the selections
	// in the TPMs answer as an empty set using tpm2.AlgUnknown; take special care to ignore this
	if (len(info.PCRSelection) != 1) || (len(info.PCRSelection[0].PCRs) != 0) || (info.PCRSelection[0].Hash != tpm2.AlgUnknown) {
		if len(info.PCRSelection) != len(digests) {
			tel.Log(ctx).
				WithField("received", len(digests)).
				WithField("quoted", len(info.PCRSelection)).
				Error("Number of PCR banks mismatch")
			return ErrQuote
		}

		for _, pcrsel := range info.PCRSelection {
			hash := strconv.FormatUint(uint64(pcrsel.Hash), 10)
			vals, ok := digests[hash]
			if !ok {
				tel.Log(ctx).
					WithField("received", reflect.ValueOf(digests).MapKeys()).
					WithField("quoted", int(pcrsel.Hash)).
					Error("Quoted PCR bank missing in evidence")
				return ErrQuote
			}

			if len(pcrsel.PCRs) != len(vals) {
				tel.Log(ctx).
					WithField("received", digests).
					WithField("quoted", pcrsel.PCRs).
					Error("Wrong PCR set")
				return ErrQuote
			}

			for _, pcr := range pcrsel.PCRs {
				if vals[strconv.Itoa(pcr)] == nil {
					tel.Log(ctx).WithField("pcr", pcr).Error("pcr missing")
					return ErrQuote
				}
			}
		}
	}

	signingHasher := sha256.New()
	for _, pcrsel := range info.PCRSelection {
		for _, pcr := range pcrsel.PCRs {
			signingHasher.Write(digests[strconv.Itoa(int(pcrsel.Hash))][strconv.Itoa(pcr)])
		}
	}

	pcrDigest := signingHasher.Sum([]byte{})
	if !bytes.Equal(info.PCRDigest, pcrDigest) {
		tel.Log(ctx).
			WithField("expected", pcrDigest).
			WithField("received", info.PCRDigest).
			Error("Wrong PCR digest")
		return ErrQuote
	}

	var hit bool
	for _, h := range fwPropsHashes {
		hit = hit || bytes.Equal(attest.ExtraData, h)
	}
	if !hit {
		tel.Log(ctx).
			WithField("expected", fwPropsHashes).
			WithField("received", attest.ExtraData).
			Error("Wrong firmware properties hash")
		return ErrQuote
	}

	if attest.Magic != 0xFF544347 {
		tel.Log(ctx).
			WithField("expected", 0xFF544347).
			WithField("received", attest.Magic).
			Error("Wrong attestation structure magic value")
		return ErrQuote
	}

	if attest.ClockInfo.Clock < lastTimestamp {
		tel.Log(ctx).
			WithField("last", lastTimestamp).
			WithField("received", attest.ClockInfo.Clock).
			Error("Clock was rolled back")
		return ErrQuote
	}

	if attest.FirmwareVersion < lastFirmware {
		tel.Log(ctx).
			WithField("last", lastFirmware).
			WithField("received", attest.FirmwareVersion).
			Error("Firmware was rolled back")
		return ErrQuote
	}

	return nil
}

func verifyQuoteSignature(ctx context.Context, attest tpm2.AttestationData, signature *api.Signature, quoteKeyQName tpm2.Name, quoteKey crypto.PublicKey) error {
	nameAlgHash, err := quoteKeyQName.Digest.Alg.Hash()
	if err != nil {
		tel.Log(ctx).WithError(err).Error("create hasher")
		return ErrFormat
	}

	attestBlob, err := attest.Encode()
	if err != nil {
		tel.Log(ctx).WithError(err).Error("reencode attestation structure")
		return ErrFormat
	}

	attestHasher := nameAlgHash.New()
	attestHasher.Write(attestBlob)
	attestHash := attestHasher.Sum([]byte{})

	var sigValid = false
	if signature.ECC != nil {
		ec, ok := (quoteKey).(*ecdsa.PublicKey)
		if !ok || ec.Curve != elliptic.P256() {
			tel.Log(ctx).Error("Quote key is not a ECDSA key over NIST P-256")
			return ErrQuote
		}
		sigValid = ecdsa.Verify(ec, attestHash, signature.ECC.R, signature.ECC.S)
	} else if signature.RSA != nil {
		pssOpt := rsa.PSSOptions{Hash: crypto.SHA256, SaltLength: rsa.PSSSaltLengthAuto}
		sigValid = rsa.VerifyPSS(quoteKey.(*rsa.PublicKey), crypto.SHA256, attestHash, signature.RSA.Signature, &pssOpt) == nil
	}

	if !sigValid {
		tel.Log(ctx).Error("Attestation signature invalid")
		return ErrQuote
	} else {
		return nil
	}
}
