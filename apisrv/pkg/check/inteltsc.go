package check

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type intelTSCPlatformRegs struct{}

func (intelTSCPlatformRegs) String() string {
	return "Intel TSC PCR"
}

func (intelTSCPlatformRegs) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.IntelTSCData == nil {
		return nil
	} else if subj.Baseline.AllowTSCPlaformRegsMismatch {
		return nil
	}
	gooddoc, err := subj.IntelTSCData.Verify()
	if err != nil {
		tel.Log(ctx).WithError(err).Error("verify xml")
		return nil
	}

	// PCR values
	tscpcr := gooddoc.Root().FindElements("/DirectPlatformData/PCRs/PCR")
	if len(tscpcr) == 0 {
		tel.Log(ctx).Error("no pcrs in the xml")
		return nil
	}

	var iss issuesv1.TscPcrValues
	iss.Common.Id = issuesv1.TscPcrValuesId
	iss.Common.Aspect = issuesv1.TscPcrValuesAspect
	iss.Common.Incident = true
	hit := false

outer:
	for _, bank := range subj.Values.PCR {
		iss.Args.Values = nil

		for _, pcr := range tscpcr {
			num := pcr.SelectAttr("id")
			val := pcr.Text()

			if num != nil && val != "" {
				if have, ok := bank[num.Value]; ok {
					if !strings.EqualFold(have, val) {
						if len(have) == len(val) {
							iss.Args.Values = append(iss.Args.Values, issuesv1.TscPcrValuesPcr{
								Number: num.Value,
								Quoted: have,
								Tsc:    val,
							})
							continue outer
						}
					}
				}

				hit = true
				break
			}
		}

		if !hit {
			tel.Log(ctx).Info("no pcr bank matches xml values")
		}
	}

	if len(iss.Args.Values) > 0 {
		return &iss
	} else {
		return nil
	}
}

func (intelTSCPlatformRegs) Update(ctx context.Context, overrides []string, subj *Subject) {
	if hasIssue(overrides, issuesv1.TscPcrValuesId) && !subj.Baseline.AllowTSCPlaformRegsMismatch {
		subj.Baseline.AllowTSCPlaformRegsMismatch = true
		subj.BaselineModified = true
	}
}

type intelTSCEndorsementKey struct{}

func (intelTSCEndorsementKey) String() string {
	return "Intel TSC EK"
}

func (intelTSCEndorsementKey) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.IntelTSCData == nil {
		return nil
	} else if subj.Baseline.AllowTSCEndorsementKeyMismatch {
		return nil
	}
	gooddoc, err := subj.IntelTSCData.Verify()
	if err != nil {
		tel.Log(ctx).WithError(err).Error("verify xml")
		return nil
	}

	// XXX: cross check data with SMBIOS table

	if subj.Baseline.EndorsementCertificate == nil {
		return nil
	}

	ek := (*x509.Certificate)(subj.Baseline.EndorsementCertificate)

	var iss issuesv1.TscEndorsementCertificate
	iss.Common.Id = issuesv1.TscEndorsementCertificateId
	iss.Common.Aspect = issuesv1.TscEndorsementCertificateAspect
	iss.Common.Incident = true
	iss.Args.EkIssuer = ek.Issuer.String()
	iss.Args.EkSerial = ek.Issuer.SerialNumber

	// EK subject in Platform Certificate
	hit := false
	for _, pc := range subj.PlatformCertificates {
		iss.Args.HolderIssuer = pc.Holder.Issuer.String()
		hit = hit || pc.Holder.Issuer.String() == ek.Issuer.String()
	}
	if !hit && len(subj.PlatformCertificates) > 0 {
		tel.Log(ctx).WithField("ek", ek.Issuer.String()).Info("ek issuer does not match platform certificate")
		iss.Args.Error = issuesv1.HolderIssuer
		return &iss
	}

	if ek.SerialNumber != nil {
		myeksn := strings.ToLower(fmt.Sprintf("%x", ek.SerialNumber))
		iss.Args.EkSerial = myeksn

		// EK serial number in XML
		eksn := gooddoc.Root().FindElement("/DirectPlatformData/EKCertSerialNumber")
		iss.Args.XmlSerial = eksn.Text()
		if eksn != nil {
			if strings.ToLower(eksn.Text()) != myeksn {
				tel.Log(ctx).WithFields(log.Fields{"ek": myeksn, "xml": eksn.Text()}).
					Info("ek serial does not match xml")
				iss.Args.Error = issuesv1.XmlSerial
				return &iss
			}
		} else {
			tel.Log(ctx).Error("no ek serial in xml")
		}

		// EK serial number and issuer in Platform Certificate
		hit := false
		for _, pc := range subj.PlatformCertificates {
			iss.Args.HolderSerial = pc.Holder.Serial.String()
			hit = hit || pc.Holder.Serial.Cmp(ek.SerialNumber) == 0
		}
		if !hit && len(subj.PlatformCertificates) > 0 {
			tel.Log(ctx).WithField("ek", ek.SerialNumber.String()).Info("ek serial not in platform certificate")
			iss.Args.Error = issuesv1.HolderSerial
			return &iss
		}
	}

	// XXX: verify platform cert itself

	return nil
}

func (intelTSCEndorsementKey) Update(ctx context.Context, overrides []string, subj *Subject) {
	if hasIssue(overrides, issuesv1.TscEndorsementCertificateId) && !subj.Baseline.AllowTSCEndorsementKeyMismatch {
		subj.Baseline.AllowTSCEndorsementKeyMismatch = true
		subj.BaselineModified = true
	}
}
