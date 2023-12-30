package check

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/inteltsc"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

var sr630EK = `
-----BEGIN CERTIFICATE-----
MIIDUTCCAvagAwIBAgIKGfVgNEmG2V1/BTAKBggqhkjOPQQDAjBVMVMwHwYDVQQD
ExhOdXZvdG9uIFRQTSBSb290IENBIDIxMTEwJQYDVQQKEx5OdXZvdG9uIFRlY2hu
b2xvZ3kgQ29ycG9yYXRpb24wCQYDVQQGEwJUVzAeFw0yMDEyMDIyMjAzMTZaFw00
MDExMjgyMjAzMTZaMAAwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDC
ldQn4cljF9fC4i6riyv6CsxyF3gOeKb6yEjJfzx1iwjuI7q7WUQ8rnIFvb/4n6Ob
k6hEXKTtm0kFaLYppoOQSCsSjkkJ3EEIgaBG48Xt+EvNZVJmFRGzlQVvtzRUrSE0
M4tFD4a4/nAzrIxYIPYP1ikGgIgtG+dSrAjLAiEGvt/SEFG2oh2SOeQB1R1WLo2T
3sHyo5KSA2MbB65yGUClq2F30ywzGy7BBIuW061wo0OlZckpdwYp7AiITEWRHnPq
HwPe2Sja4qXeMWifCNukUySQEywguYN3nL0uG7eTxgnivMt8mLEoYzERpCfIvZo/
I+otp6C042wLjdAl13qnAgMBAAGjggE2MIIBMjBQBgNVHREBAf8ERjBEpEIwQDE+
MBQGBWeBBQIBEwtpZDo0RTU0NDMwMDAQBgVngQUCAhMHTlBDVDc1eDAUBgVngQUC
AxMLaWQ6MDAwNzAwMDIwDAYDVR0TAQH/BAIwADAQBgNVHSUECTAHBgVngQUIATAf
BgNVHSMEGDAWgBQj9OIq0743SkSXcpVKooOu11JXLjAOBgNVHQ8BAf8EBAMCBSAw
IgYDVR0JBBswGTAXBgVngQUCEDEOMAwMAzIuMAIBAAICAIowaQYIKwYBBQUHAQEE
XTBbMFkGCCsGAQUFBzAChk1odHRwczovL3d3dy5udXZvdG9uLmNvbS9zZWN1cml0
eS9OVEMtVFBNLUVLLUNlcnQvTnV2b3RvbiBUUE0gUm9vdCBDQSAyMTExLmNlcjAK
BggqhkjOPQQDAgNJADBGAiEAygGbYg8t+eejnAE8f97WlaVRRD595m1hCpNRvIQn
rhQCIQCJwiAj/l1M5hx8YjRl5Y51vyK6rbAuGXPYdba0BVh2nw==
-----END CERTIFICATE-----
`

var h12sslEK = `
-----BEGIN CERTIFICATE-----
MIIElTCCA32gAwIBAgIEEOoxlDANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC
REUxITAfBgNVBAoMGEluZmluZW9uIFRlY2hub2xvZ2llcyBBRzEaMBgGA1UECwwR
T1BUSUdBKFRNKSBUUE0yLjAxNTAzBgNVBAMMLEluZmluZW9uIE9QVElHQShUTSkg
UlNBIE1hbnVmYWN0dXJpbmcgQ0EgMDI5MB4XDTIxMDEyOTIzMDA0N1oXDTM2MDEy
OTIzMDA0N1owADCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAIvA7RGl
BqfsyjIqNg7UZz76qjk7hb3fzeyKh74dJOnAlPo/7CenXNOjda1UHwzFLH9j2KZJ
1TPYO9ijPBM98s4nTF+gXW5mDYzltICrZkjLf7fH61yWxPLI0FT4L/P1/9QLOktT
UGpm0Hf3VmlyUQqXRIqTr/yT1kvhhebq6T9WyvfOnzSLJDS5EXFS8o6qjx94yNUM
B0pd6DISrkk4OIjBpd8K5C2m/Hx4gidK4u0huZZEE9HLB8Uu/6OSwW8fk6tmpAgt
8PVhEb49qq81o/bJCDHDpONt4UgObRNIopbIQ3lCZDxE5NkM3xxSysgo1866wnFM
iia21yHz+MPy59kCAwEAAaOCAZEwggGNMFsGCCsGAQUFBwEBBE8wTTBLBggrBgEF
BQcwAoY/aHR0cDovL3BraS5pbmZpbmVvbi5jb20vT3B0aWdhUnNhTWZyQ0EwMjkv
T3B0aWdhUnNhTWZyQ0EwMjkuY3J0MA4GA1UdDwEB/wQEAwIAIDBRBgNVHREBAf8E
RzBFpEMwQTEWMBQGBWeBBQIBDAtpZDo0OTQ2NTgwMDETMBEGBWeBBQICDAhTTEIg
OTY2NTESMBAGBWeBBQIDDAdpZDowNTNFMAwGA1UdEwEB/wQCMAAwUAYDVR0fBEkw
RzBFoEOgQYY/aHR0cDovL3BraS5pbmZpbmVvbi5jb20vT3B0aWdhUnNhTWZyQ0Ew
MjkvT3B0aWdhUnNhTWZyQ0EwMjkuY3JsMBUGA1UdIAQOMAwwCgYIKoIUAEQBFAEw
HwYDVR0jBBgwFoAUGLGvcLk/mRly82JVapo/v0uyTg0wEAYDVR0lBAkwBwYFZ4EF
CAEwIQYDVR0JBBowGDAWBgVngQUCEDENMAsMAzIuMAIBAAIBdDANBgkqhkiG9w0B
AQsFAAOCAQEAlrbOhNOUC14GXVkc9t7bFRrmdMOsG6icmAnueNDdWuh2YnXq9koY
NxF4lIcjiTiLBIBWhN0m0cAYtqoOMf00V5wh2kGUExJ0a0PVSoQYVsdpAzKvrmak
vfIgCZMxFMxbvHu8K4vCb2oMQaQYs6Em0pMmB+JCZUzuVMD+DNXFJfcPKweZ6R9s
S/P3fD9n4buFZ3zv6hBBmTDIImgw2tWgMIYrLYLHDnkZyWkyhHJPll9ha/x+clFG
EiRSTMC0rme+7b5hFd9Ntu/xQ215yEVa32DlSZm6QmXM2V+Jvbi2dLSApqHuvRVe
8ez0MOtNZ3VX/PB83Xuvul/s3sBT/muz8g==
-----END CERTIFICATE-----
`

func TestIntelTSC(t *testing.T) {
	ctx := context.Background()
	beforesubj := parseEvidence(t, "../../test/sr630.evidence.json")
	aftersubj := parseEvidence(t, "../../test/sr630.evidence.json")
	file, err := ioutil.ReadFile("../../test/J101WYR1-2022-06-21.zip")
	assert.NoError(t, err)
	beforeblk, _ := pem.Decode([]byte(sr630EK))
	assert.NotNil(t, beforeblk)
	beforeek, err := x509.ParseCertificate(beforeblk.Bytes)
	assert.NoError(t, err)
	afterblk, _ := pem.Decode([]byte(h12sslEK))
	assert.NotNil(t, afterblk)
	afterek, err := x509.ParseCertificate(afterblk.Bytes)
	assert.NoError(t, err)
	rawxml, certs, err := inteltsc.UnpackZip(ctx, file)
	assert.NoError(t, err)
	data, err := inteltsc.ParseData(rawxml)
	assert.NoError(t, err)

	beforesubj.IntelTSCData = data
	beforesubj.PlatformCertificates = certs
	beforesubj.Baseline.EndorsementCertificate = (*api.Certificate)(beforeek)

	aftersubj.IntelTSCData = data
	aftersubj.PlatformCertificates = certs
	aftersubj.Values.TPM2Properties = []api.TPM2Property{}
	aftersubj.Baseline.EndorsementCertificate = (*api.Certificate)(afterek)

	//t.Run("PCR", func(t *testing.T) {
	//	testCheckImpl(t, beforesubj, aftersubj, intelTSCPlatformRegs{}, api.AnnTSCPCRMismatch, false, func(*Subject) {})
	//})

	t.Run("EK", func(t *testing.T) {
		testCheckImpl(t, beforesubj, aftersubj, intelTSCEndorsementKey{}, issuesv1.TscEndorsementCertificateId, true, func(*Subject) {})
	})
}
