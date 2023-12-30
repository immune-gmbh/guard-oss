package check

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/immune-gmbh/agent/v3/pkg/tcg"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func TestNoEventlog(t *testing.T) {
	subj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	subj.Boot = *evidence.EmptyBoot()
	res, err := Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.True(t, hasIssueId(res.Issues, issuesv1.TpmNoEventlogId))

	// construct a dummy TPM and see if the annotation goes away when using dummy TPM
	a, err := tcg.NewSoftwareAnchor()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	cert, err := a.ReadEKCertificate()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	subj.Baseline.EndorsementCertificate = (*api.Certificate)(cert)

	res, err = Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.False(t, hasIssueId(res.Issues, issuesv1.TpmNoEventlogId))
}

func TestInvalidEventlog(t *testing.T) {
	testPure(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		tpmEventLog{},
		issuesv1.TpmInvalidEventlogId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			raw := (*ev).RawEvent()
			if len(raw.Digest) > 0 {
				raw.Digest[0] ^= 1
			}
		}))
}

func TestDummyTPM(t *testing.T) {
	subj := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")

	a, err := tcg.NewSoftwareAnchor()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	cert, err := a.ReadEKCertificate()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	subj.Baseline.EndorsementCertificate = (*api.Certificate)(cert)

	// test if annotation returns
	res, err := Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.True(t, hasIssueId(res.Issues, issuesv1.TpmDummyId))

	// test with nil certificate
	subj.Baseline.EndorsementCertificate = nil
	res, err = Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.False(t, hasIssueId(res.Issues, issuesv1.TpmDummyId))
}

func TestEK(t *testing.T) {
	beforesubj := parseEvidence(t, "../../test/DESKTOP-2AG9807.json")
	aftersubj := parseEvidence(t, "../../test/DESKTOP-2AG9807.json")
	blk, _ := pem.Decode([]byte(h12sslEK))
	assert.NotNil(t, blk)
	beforeek, err := x509.ParseCertificate(blk.Bytes)
	assert.NoError(t, err)
	blk, _ = pem.Decode([]byte(h12sslEK))
	assert.NotNil(t, blk)
	afterek, err := x509.ParseCertificate(blk.Bytes)
	assert.NoError(t, err)
	beforesubj.Baseline.EndorsementCertificate = (*api.Certificate)(beforeek)
	aftersubj.Baseline.EndorsementCertificate = (*api.Certificate)(afterek)

	testCheckImpl(t,
		beforesubj,
		aftersubj,
		tpmEndorsementCertificate{},
		issuesv1.TpmEndorsementCertUnverifiedId,
		true,
		func(subj *Subject) {
			subj.Baseline.EndorsementCertificate.Signature[10] = 1
		})
}

const philippEK string = `
-----BEGIN CERTIFICATE-----
MIIEBTCCAu2gAwIBAgIUAnFDGiuU6fV4aZXW6SvTdTl3bvwwDQYJKoZIhvcNAQEL
BQAwVTELMAkGA1UEBhMCQ0gxHjAcBgNVBAoTFVNUTWljcm9lbGVjdHJvbmljcyBO
VjEmMCQGA1UEAxMdU1RNIFRQTSBFSyBJbnRlcm1lZGlhdGUgQ0EgMDYwHhcNMjEw
NDIxMDAwMDAwWhcNNDkxMjMxMDAwMDAwWjAAMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEA6Q7tGw+qyfJjJ4X9J4kzL1w827FjTaMMCSvyD5+OA4VDIG3w
cpXEh/nip8ubwUyOi62lgu/d60Ax5JKSn9vUugUlQcAl8OU8EvnIOB1qN3AIqjla
5R0Jkr/k7N6EBxarkEHWnoYAxiiKJq4S3JnoublO+npqJjoxN2Bb9+8JsLXiEWII
q8B1hreb3/2+3S6eekUJl1AUdrhCoE57PkqkWjbgNNWtZuz7HRkySXPulg0VtQww
ZFAhXoHPTkH3oDOic54Oq/YUqvbF4k2lynghFHwzkV0DR/I4h2Og/L3Gt2sKVuqf
hJxIx+Or/cT9SqlsaNVQqsf9I0hdEhT7Nw9gjQIDAQABo4IBIDCCARwwHwYDVR0j
BBgwFoAU+xfXDXNIcOkZxOjmA5deZk4OQ94wWQYDVR0RAQH/BE8wTaRLMEkxFjAU
BgVngQUCAQwLaWQ6NTM1NDREMjAxFzAVBgVngQUCAgwMU1QzM0hUUEhBSEQ4MRYw
FAYFZ4EFAgMMC2lkOjAwMDEwMTAyMCIGA1UdCQQbMBkwFwYFZ4EFAhAxDjAMDAMy
LjACAQACAgCKMAwGA1UdEwEB/wQCMAAwEAYDVR0lBAkwBwYFZ4EFCAEwDgYDVR0P
AQH/BAQDAgUgMEoGCCsGAQUFBwEBBD4wPDA6BggrBgEFBQcwAoYuaHR0cDovL3Nl
Y3VyZS5nbG9iYWxzaWduLmNvbS9zdG10cG1la2ludDA2LmNydDANBgkqhkiG9w0B
AQsFAAOCAQEAEneDKXFu3OZ9p3ECgHgIFz06AtxOjyjov2bvA3pz7YxlXtLTSow2
6TWfs1s0d08Dq6OMJXRmcGES3ZVfgzBs4FvriFnIscCJbW5fDcj7rt+EYTBGIiA+
K0q3kbPJWtfDZtjApfaXSmVEFJKF+qDY60Vg8uD52jIcBZn2qM403T2KuGPADHQJ
TilFC50DgPRe3XuspUws94mq/7/qppzgf22Jdxexr8rXJhdeOo/S5xpaOmvfCMhP
GA1T9bUUF7wM6Nj55snrxyASnzXXJ9SpAzZHVH+YFcqehQmHFFYSJWagxgZdD5fn
0/lH30eMAos5sP45Z8D7f3W4hBs7kUGgtQ==
-----END CERTIFICATE-----
`

func TestEKCertPhilipp(t *testing.T) {
	subj := parseEvidence(t, "../../test/issue926-x1carbon-before.evidence.json")
	blk, _ := pem.Decode([]byte(philippEK))
	assert.NotNil(t, blk)
	ek, err := x509.ParseCertificate(blk.Bytes)
	assert.NoError(t, err)
	subj.Baseline.EndorsementCertificate = (*api.Certificate)(ek)

	res, err := Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.False(t, hasIssueId(res.Issues, issuesv1.TpmEndorsementCertUnverifiedId))
}

func TestEventLog(t *testing.T) {
	testPure(t,
		"../../test/test-before.json",
		"../../test/test-update.json",
		tpmEventLog{},
		"",
		func(subj *Subject) {})
}

func TestHahuriEventlog(t *testing.T) {
	subj := parseValues(t, "../../test/hahuri-invalid-eventlog.values.json")
	res, err := Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.False(t, hasIssueId(res.Issues, issuesv1.TpmInvalidEventlogId))
}

func TestInvalidLenovoEventlog(t *testing.T) {
	subj := parseValues(t, "../../test/t480s.values.json")
	res, err := Run(context.Background(), subj)
	assert.NoError(t, err)
	assert.False(t, hasIssueId(res.Issues, issuesv1.TpmInvalidEventlogId))
}
