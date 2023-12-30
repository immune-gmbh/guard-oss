package check

import (
	"context"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"embed"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-tpm/tpm2"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	y509 "github.com/immune-gmbh/guard/apisrv/v2/pkg/x509"
	log "github.com/sirupsen/logrus"
)

var (
	//go:embed certs
	certificateFiles embed.FS
	certificatePool  *x509.CertPool
)

func init() {
	certificatePool = x509.NewCertPool()
	certs := make([]*x509.Certificate, 0)
	files, err := certificateFiles.ReadDir("certs")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !file.IsDir() {
			buf, err := certificateFiles.ReadFile(path.Join("certs", file.Name()))
			if err != nil {
				panic(fmt.Sprint("read: ", file.Name(), ": ", err))
			}
			if strings.Contains(string(buf), "BEGIN CERTIFICATE") {
				blk, err := pem.Decode(buf)
				if err != nil {
					buf = blk.Bytes
				}
			}
			cert, err := x509.ParseCertificate(buf)
			if err != nil {
				panic(fmt.Sprint("parse: ", file.Name(), ": ", err))
			}
			certificatePool.AddCert(cert)
			certs = append(certs, cert)
		}
	}

	// verify certificates
	knownKp := []asn1.ObjectIdentifier{
		tcgKpEKCertificate,
		tcgKpAIKCertificate,
		tcgKpPlatformCertificate,
		tcgKpPlatformKeyCertificate,
		msCertSrv,
		msPlutonUnknown,
	}
	opts := x509.VerifyOptions{Roots: certificatePool}
	for _, cert := range certs {
		// accept TCG ExtKeyUsage
		flt := cert.UnknownExtKeyUsage[:0]
		for _, ku := range cert.UnknownExtKeyUsage {
			var isTcgKp bool
			for _, kp := range knownKp {
				isTcgKp = isTcgKp || ku.Equal(kp)
			}
			if !isTcgKp {
				flt = append(flt, ku)
			}
		}
		cert.UnknownExtKeyUsage = flt

		_, err := cert.Verify(opts)
		var certerr x509.CertificateInvalidError
		// accept expired certificates
		if errors.As(err, &certerr) && certerr.Reason == x509.Expired {
			continue
		}
		if err != nil {
			panic(fmt.Sprint("verify: ", cert.Subject.String(), ": ", err))
		}
	}
}

type tpmEventLog struct{}

func (tpmEventLog) String() string {
	return "TPM event log"
}

func verifyBank(ctx context.Context, quoted [][]byte, log *eventlog.EventLog, bank crypto.Hash) []*issuesv1.TpmInvalidEventlogPcr {
	var pcrs []eventlog.PCR
	for idx, val := range quoted {
		pcrs = append(pcrs, eventlog.PCR{
			Index:     idx,
			Digest:    val,
			DigestAlg: bank,
		})
	}
	_, err := log.Verify(pcrs)
	if err == nil {
		return nil
	}

	var ret []*issuesv1.TpmInvalidEventlogPcr
	var replayErr eventlog.ReplayError
	if errors.As(err, &replayErr) {
		tel.Log(ctx).WithError(err).WithField("bank", bank).Info("replay eventlog")
		replay, _ := log.Replay(pcrs)

		for _, idx := range replayErr.InvalidPCRs {
			computed := replay[idx]

			fmt.Printf("PCR %d: quoted %x, computed %x\n", idx, quoted[idx], computed)

			// pcr does not show up in event log -> must have it's reset state
			if len(computed) == 0 && len(quoted[idx]) > 0 {
				// TCG PC Client Platform TPM Profile Specification for TPM 2.0
				// Section 4.6.2 PCR Initial and Reset Values
				var fill byte
				if idx <= 16 {
					fill = 0
				} else if idx <= 22 {
					fill = 0xff
				} else {
					fill = 0
				}

				match := true
				for _, b := range quoted[idx] {
					match = match && b == fill
				}
				if match {
					continue
				}
				fmt.Printf("pcr %d should be initilized to %x\n", idx, fill)
			}
			args := issuesv1.TpmInvalidEventlogPcr{
				Number: fmt.Sprint(idx),
			}
			args.Computed = fmt.Sprintf("%x", replay[idx])
			args.Quoted = fmt.Sprintf("%x", quoted[idx])
			ret = append(ret, &args)
		}
	}
	return ret
}

func (tpmEventLog) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subjectHasDummyTPM(subj) {
		return nil
	}

	if (subj.Boot.IsEmpty || subj.EventLogs == nil) && !subj.Baseline.AllowNoEventlog {
		var iss issuesv1.TpmNoEventlog
		iss.Common.Id = issuesv1.TpmNoEventlogId
		iss.Common.Aspect = issuesv1.TpmNoEventlogAspect
		iss.Common.Incident = false
		return &iss
	}

	// verify log
	log := subj.EventLogs[subj.CurrentEventLogIdx]

	// hack: see below
	var sha256passed bool

	// iterate thu all banks
	var failures []*issuesv1.TpmInvalidEventlogPcr
	for alg, bank := range subj.Values.PCR {
		var hash crypto.Hash
		switch alg {
		case "4":
			hash = crypto.SHA1
		case "11":
			hash = crypto.SHA256
		default:
			continue
		}

		var quoted [][]byte

		for stridx, strval := range bank {
			idx, err := strconv.Atoi(stridx)
			if err != nil {
				var iss issuesv1.TpmInvalidEventlog
				iss.Common.Id = issuesv1.TpmInvalidEventlogId
				iss.Common.Aspect = issuesv1.TpmInvalidEventlogAspect
				iss.Common.Incident = true
				iss.Args.Error = issuesv1.FormatInvalid
				return &iss
			}
			if len(quoted) <= idx {
				quoted = append(quoted, make([][]byte, idx+1-len(quoted))...)
			}
			val, err := hex.DecodeString(strval)
			if err != nil {
				var iss issuesv1.TpmInvalidEventlog
				iss.Common.Id = issuesv1.TpmInvalidEventlogId
				iss.Common.Aspect = issuesv1.TpmInvalidEventlogAspect
				iss.Common.Incident = true
				iss.Args.Error = issuesv1.FormatInvalid
				return &iss
			}

			quoted[idx] = val
		}

		newfailures := verifyBank(ctx, quoted, log, hash)
		sha256passed = sha256passed || len(newfailures) == 0 && hash == crypto.SHA256
		failures = append(failures, newfailures...)
	}

	// hack: we shoud have a unified event log abstraction
	if len(subj.IMALog) > 0 {
		var filtered []*issuesv1.TpmInvalidEventlogPcr
		for _, f := range failures {
			if f.Number == "10" || f.Number == "11" {
				continue
			}
			filtered = append(filtered, f)
		}
		failures = filtered
	}

	if len(failures) == 0 || subj.Baseline.AllowInvalidEventlog {
		return nil
	}

	// hack: on the t480s the sha1 event log is completely wrong while the sha256
	// still works.
	manu, serial, err := subj.Values.PlatformSerial()
	if sha256passed && err == nil && manu == "LENOVO" && serial == "PF0TP4UD" {
		return nil
	}

	var iss issuesv1.TpmInvalidEventlog
	iss.Common.Id = issuesv1.TpmInvalidEventlogId
	iss.Common.Aspect = issuesv1.TpmInvalidEventlogAspect
	iss.Common.Incident = true
	iss.Args.Pcr = failures
	iss.Args.Error = issuesv1.PcrMismatch

	return &iss
}

func (tpmEventLog) Update(ctx context.Context, overrides []string, subj *Subject) {
	if hasIssue(overrides, issuesv1.TpmNoEventlogId) && !subj.Baseline.AllowNoEventlog {
		subj.Baseline.AllowNoEventlog = true
		subj.BaselineModified = true
	}
	if hasIssue(overrides, issuesv1.TpmInvalidEventlogId) && !subj.Baseline.AllowInvalidEventlog {
		subj.Baseline.AllowInvalidEventlog = true
		subj.BaselineModified = true
	}
}

// check EK cert against https://trustedcomputinggroup.org/wp-content/uploads/Credential_Profile_EK_V2.0_R14_published.pdf
type tpmEndorsementCertificate struct{}

func (tpmEndorsementCertificate) String() string {
	return "TPM endorsement key certificate"
}

var (
	ErrinvalidName = errors.New("invalid ASN.1 name")
	ErrdupName     = errors.New("duplicate ASN.1 name")

	tcgAttributeTpmManufacturer = asn1.ObjectIdentifier{2, 23, 133, 2, 1}
	tcgAttributeTpmModel        = asn1.ObjectIdentifier{2, 23, 133, 2, 2}
	tcgAttributeTpmVersion      = asn1.ObjectIdentifier{2, 23, 133, 2, 3}

	tcgKpEKCertificate          = asn1.ObjectIdentifier{2, 23, 133, 8, 1}
	tcgKpPlatformCertificate    = asn1.ObjectIdentifier{2, 23, 133, 8, 2}
	tcgKpAIKCertificate         = asn1.ObjectIdentifier{2, 23, 133, 8, 3}
	tcgKpPlatformKeyCertificate = asn1.ObjectIdentifier{2, 23, 133, 8, 4}
	msCertSrv                   = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 21, 36}
	msPlutonUnknown             = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 122, 2}

	x509SubjectAltNameExt = asn1.ObjectIdentifier{2, 5, 29, 17}
)

func parseIntName(value interface{}) (uint32, error) {
	if strval, ok := value.(string); ok {
		if strings.HasPrefix(strval, "id:") {
			if intval, err := strconv.ParseUint(strings.TrimPrefix(strval, "id:"), 16, 32); err == nil {
				return uint32(intval), nil
			}
		}
	}
	return 0, ErrinvalidName
}

func parseTCGName(name *pkix.Name) (uint32, string, uint32, error) {
	var (
		manuf, ver *uint32
		model      *string
	)

	for _, n := range name.Names {
		if n.Type.Equal(tcgAttributeTpmManufacturer) {
			if manuf != nil {
				return 0, "", 0, ErrdupName
			}
			m, err := parseIntName(n.Value)
			if err != nil {
				return 0, "", 0, err
			}
			manuf = &m
		} else if n.Type.Equal(tcgAttributeTpmVersion) {
			if ver != nil {
				return 0, "", 0, ErrdupName
			}
			v, err := parseIntName(n.Value)
			if err != nil {
				return 0, "", 0, err
			}
			ver = &v
		} else if n.Type.Equal(tcgAttributeTpmModel) {
			if model != nil {
				return 0, "", 0, ErrdupName
			}
			strval, ok := n.Value.(string)
			if !ok {
				return 0, "", 0, ErrinvalidName
			}
			model = &strval
		}
	}

	if manuf != nil && ver != nil && model != nil {
		return *manuf, *model, *ver, nil
	}

	return 0, "", 0, ErrinvalidName
}

func (tpmEndorsementCertificate) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Baseline.EndorsementCertificate == nil || subj.Baseline.AllowEKCertificateUnverified || subjectHasDummyTPM(subj) {
		return nil
	}

	cert := (*x509.Certificate)(subj.Baseline.EndorsementCertificate)

	var iss issuesv1.TpmEndorsementCertUnverified
	iss.Common.Id = issuesv1.TpmEndorsementCertUnverifiedId
	iss.Common.Aspect = issuesv1.TpmEndorsementCertUnverifiedAspect
	iss.Common.Incident = true
	iss.Args.EkIssuer = cert.Issuer.String()

	// remove SAN from unhandle crit ext
	filteredExt := cert.UnhandledCriticalExtensions[:0]
	for _, extid := range cert.UnhandledCriticalExtensions {
		if !extid.Equal(x509SubjectAltNameExt) {
			filteredExt = append(filteredExt, extid)
		}
	}
	cert.UnhandledCriticalExtensions = filteredExt

	// check extended key usage extension
	var hasEkKeyUsage bool
	for _, ku := range cert.UnknownExtKeyUsage {
		hasEkKeyUsage = hasEkKeyUsage || ku.Equal(tcgKpEKCertificate)
	}
	if !hasEkKeyUsage {
		tel.Log(ctx).Info("no tcg extKeyUsage")
		iss.Args.Error = issuesv1.NoEku
		return &iss
	}

	// verify signature
	opts := x509.VerifyOptions{
		Roots: certificatePool,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageAny,
		},
	}
	_, err := cert.Verify(opts)
	var certerr x509.CertificateInvalidError
	// accept expired certificates
	if errors.As(err, &certerr) && certerr.Reason == x509.Expired {
		tel.Log(ctx).Info("ek expired")
	} else if err != nil {
		tel.Log(ctx).WithError(err).Info("verify ek certificate")
		iss.Args.Error = issuesv1.InvalidCertificate
		return &iss
	}

	// handle TCG SAN
	type sanType struct {
		Manufacturer uint32
		Model        string
		Version      uint32
	}
	var san *sanType
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(x509SubjectAltNameExt) {
			if san != nil {
				tel.Log(ctx).Warn("dup SAN extension")
				break
			}
			_, _, _, _, names, err := y509.ParseAllSANExtension(ext.Value)
			if err != nil {
				tel.Log(ctx).WithError(err).Info("parse SAN extension")
				break
			}
			if len(names) != 1 {
				tel.Log(ctx).WithError(err).Info("expected one directoryName")
				break
			}
			manuf, model, ver, err := parseTCGName(&names[0])
			if err != nil {
				tel.Log(ctx).WithError(err).Info("parse directoryName")
				break
			}
			san = &sanType{Manufacturer: manuf, Model: model, Version: ver}
		}
	}
	if san == nil {
		tel.Log(ctx).Info("no tcg directoryName")
		iss.Args.Error = issuesv1.SanInvalid
		return &iss
	}

	// read TPM manufacturer and vendor ID
	manuf, err := subj.Values.TPMProperty(tpm2.Manufacturer)
	if err != nil {
		tel.Log(ctx).WithError(err).Info("get manufacturer")
		return nil
	}

	iss.Args.EkVendor = fmt.Sprintf("%d", san.Manufacturer)
	iss.Args.EkVersion = fmt.Sprintf("%d", san.Version)
	iss.Args.Vendor = fmt.Sprintf("%d", manuf)

	if manuf != san.Manufacturer {
		tel.Log(ctx).WithFields(log.Fields{"ek": san.Manufacturer, "tpm": manuf}).Info("manufacturer wrong")
		iss.Args.Error = issuesv1.SanMismatch
		return &iss
	}

	return nil
}

func (tpmEndorsementCertificate) Update(ctx context.Context, overrides []string, subj *Subject) {
	if hasIssue(overrides, issuesv1.TpmEndorsementCertUnverifiedId) && !subj.Baseline.AllowEKCertificateUnverified {
		subj.Baseline.AllowEKCertificateUnverified = true
		subj.BaselineModified = true
	}
}

type dummyTpm struct{}

func (dummyTpm) String() string {
	return "Dummy TPM"
}

func subjectHasDummyTPM(subj *Subject) bool {
	// the EK may be empty during tests and if it is we treat it as a real TPM
	// so that we do not have to use the dummy TPM and un-dummy its cert for these tests
	if subj.Baseline.EndorsementCertificate == nil {
		return false
	}

	return strings.HasPrefix(strings.ToLower(subj.Baseline.EndorsementCertificate.Subject.CommonName), "immune gmbh software")
}

func (dummyTpm) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Baseline.EndorsementCertificate == nil {
		return nil
	}

	dummyTPM := subjectHasDummyTPM(subj)

	if dummyTPM && !subj.Baseline.AllowDummyTPM {
		var iss issuesv1.TpmDummy
		iss.Common.Id = issuesv1.TpmDummyId
		iss.Common.Aspect = issuesv1.TpmDummyAspect
		iss.Common.Incident = true
		return &iss
	}

	return nil
}

func (dummyTpm) Update(ctx context.Context, overrides []string, subj *Subject) {
	if hasIssue(overrides, issuesv1.TpmDummyId) && !subj.Baseline.AllowDummyTPM {
		subj.Baseline.AllowDummyTPM = true
		subj.BaselineModified = true
	}
}
