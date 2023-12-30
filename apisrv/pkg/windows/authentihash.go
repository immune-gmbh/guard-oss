package windows

import (
	"bytes"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"

	"go.mozilla.org/pkcs7"
)

var OIDSpcPeImageDataObj = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 15}

var ErrUnsupportedAttributeType = errors.New("authenticode: cannot parse data: unimplemented attribute type")

type SpcAttributeTypeAndOptionalValue struct {
	Type  asn1.ObjectIdentifier
	Value asn1.RawValue `asn1:"optional"`
}

type DigestInfo struct {
	DigestAlgorithm pkix.AlgorithmIdentifier
	Digest          []byte
}

// ParseSpcIndirectDataContent parses the _content_ of the SpcIndirectDataContent structure without it's header
// (to be used with the partially unpacked content sequence of the PKCS7 structure as we get it from "go.mozilla.org/pkcs7")
func ParseSpcIndirectDataContent(content []byte) (attrType SpcAttributeTypeAndOptionalValue, di DigestInfo, err error) {
	rest, err := asn1.Unmarshal(content, &attrType)
	if err != nil {
		return
	}

	_, err = asn1.Unmarshal(rest, &di)
	return
}

// CheckPEAuthentiHashSha256 checks if the PE file's authenticode hash matches the given hash
func CheckPEAuthentiHashSha256(content *pkcs7.PKCS7, hash []byte) (bool, error) {
	attrType, di, err := ParseSpcIndirectDataContent(content.Content)
	if err != nil {
		return false, err
	}

	if !attrType.Type.Equal(OIDSpcPeImageDataObj) {
		return false, ErrUnsupportedAttributeType
	}

	if len(di.Digest) != len(hash) {
		return false, nil
	}

	if !di.DigestAlgorithm.Algorithm.Equal(pkcs7.OIDDigestAlgorithmSHA256) {
		return false, nil
	}

	return bytes.Equal(di.Digest, hash), nil
}
