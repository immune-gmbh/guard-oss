package windows

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/stretchr/testify/assert"
)

func parseEvidence(t *testing.T, f string) *evidence.Values {
	buf, err := ioutil.ReadFile(f)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]api.Buffer{
			"11": ev.PCRs,
		}
	}

	val, err := evidence.WrapInsecure(&ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return val
}

const imn_supermicro_aik_name = "bla"

func Test_detectPcpKeyBlobSubType(t *testing.T) {
	ev := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	keyBlob := ev.PCPQuoteKeys[imn_supermicro_aik_name]

	subType, err := detectPcpKeyBlobSubType(keyBlob)
	if assert.NoError(t, err, "error detecting key blob sub type") {
		assert.Equal(t, PCPKeyBlobType20, subType, "unexpected key blob sub type")
	}

	// change len to len of Win 8 key blob
	keyBlob[4] = 0x36
	subType, err = detectPcpKeyBlobSubType(keyBlob)
	if assert.NoError(t, err, "error detecting key blob sub type") {
		assert.Equal(t, PCPKeyBlobTypeWin8, subType, "unexpected key blob sub type")
	}

	// change type to TPM 1.2
	keyBlob[4] = 0x30
	keyBlob[8] = 1
	subType, err = detectPcpKeyBlobSubType(keyBlob)
	if assert.NoError(t, err, "error detecting key blob sub type") {
		assert.Equal(t, PCPKeyBlobType12, subType, "unexpected key blob sub type")
	}

	// corrupt magic
	keyBlob[0] = 23
	subType, err = detectPcpKeyBlobSubType(keyBlob)
	if assert.NoError(t, err, "error detecting key blob sub type") {
		assert.Equal(t, PCPKeyBlobTypeUnknown, subType, "unexpected key blob sub type")
	}
}

func testPublicAIK(t *testing.T, keyBlob []byte, aik *pcpKeyBlobWin8) {
	// TPM2B_PUBLIC is expected and has uint16 size prefixed to TPMT_PUBLIC, DecodePublic wants TPMT_PUBLIC
	pub, err := tpm2.DecodePublic(keyBlob[aik.CbHeader+2 : aik.CbHeader+aik.CbPublic])
	if !assert.NoError(t, err, "error reading TPMT_PUBLIC from TPM 2.0 key blob") {
		t.FailNow()
	}
	assert.Equal(t, tpm2.AlgSHA256, pub.NameAlg)
	assert.NotNil(t, pub.RSAParameters)
	assert.Equal(t, tpm2.AlgRSASSA, pub.RSAParameters.Sign.Alg)
	assert.Equal(t, tpm2.AlgSHA1, pub.RSAParameters.Sign.Hash)
}

func Test_to20KeyBlob(t *testing.T) {
	ev := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	keyBlob := ev.PCPQuoteKeys[imn_supermicro_aik_name]

	aik, err := to20KeyBlob(keyBlob)
	if assert.NoError(t, err, "error decoding TPM 2.0 key blob") {
		testPublicAIK(t, keyBlob, &aik.pcpKeyBlobWin8)
	}
}

func Test_toKeyBlobWin8(t *testing.T) {
	ev := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	keyBlob := ev.PCPQuoteKeys[imn_supermicro_aik_name]

	// the evidence really contains a TPM 2.0 key blob which really is just a superset of Win8 key blobs
	aik, err := toKeyBlobWin8(keyBlob)
	if assert.NoError(t, err, "error decoding Win 8 key blob") {
		testPublicAIK(t, keyBlob, aik)
	}
}

func Test_ExtractTPMTPublic(t *testing.T) {
	ev := parseEvidence(t, "../../test/IMN-SUPERMICRO-WBCL-Quote-AIKs.evidence.json")
	keyBlob := ev.PCPQuoteKeys[imn_supermicro_aik_name]

	_, err := ExtractTPMTPublic(keyBlob)
	assert.NoError(t, err, "error decoding Win 8 key blob")
}
