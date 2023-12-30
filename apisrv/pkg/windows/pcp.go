package windows

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unsafe"

	"github.com/google/go-tpm/tpm2"
)

// this file contains functions related to the Platform Crypto Provider which manages TPM keys under windows

var (
	ErrTpm12KeyBlob   = errors.New("TPM 1.2 key blob structure not supported")
	ErrInvalidKeyBlob = errors.New("key blob invalid")
)

var bcrypt_pcp_key_magic = binary.LittleEndian.Uint32([]byte{'P', 'C', 'P', 'M'})

type PCPKeyBlobType int

const (
	PCPKeyBlobTypeUnknown PCPKeyBlobType = iota - 1
	PCPKeyBlobType12
	PCPKeyBlobTypeWin8
	PCPKeyBlobType20
)

var keyBlobTypeStings = []string{"unkown", "TPM 1.2 key blob", "Win8 TPM 2.0 key blob", "TPM 2.0 key blob"}

type pcpKeyBlobCommonHeader struct {
	Magic    uint32
	CbHeader uint32
	PcpType  uint32
}

const (
	pcptype_tpm12 uint32 = 0x00000001
	pcptype_tpm20 uint32 = 0x00000002
)

type pcpKeyBlobWin8 struct {
	Magic              uint32
	CbHeader           uint32
	PcpType            uint32
	Flags              uint32
	CbPublic           uint32
	CbPrivate          uint32
	CbMigrationPublic  uint32
	CbMigrationPrivate uint32
	CbPolicyDigestList uint32
	CbPCRBinding       uint32
	CbPCRDigest        uint32
	CbEncryptedSecret  uint32
	CbTpm12HostageBlob uint32
}

type pcp20KeyBlob struct {
	pcpKeyBlobWin8
	PcrAlgId uint16
}

func (v PCPKeyBlobType) String() string {
	return keyBlobTypeStings[v+1]
}

func detectPcpKeyBlobSubType(blob []byte) (PCPKeyBlobType, error) {
	var hdr pcpKeyBlobCommonHeader
	if err := binary.Read(bytes.NewReader(blob), binary.LittleEndian, &hdr); err != nil {
		return PCPKeyBlobTypeUnknown, err
	}

	if hdr.Magic != bcrypt_pcp_key_magic {
		return PCPKeyBlobTypeUnknown, nil
	}

	if hdr.PcpType == pcptype_tpm12 && hdr.CbHeader < uint32(unsafe.Sizeof(pcpKeyBlobWin8{})) {
		return PCPKeyBlobType12, nil
	}

	if hdr.PcpType == pcptype_tpm20 && hdr.CbHeader >= uint32(unsafe.Sizeof(pcp20KeyBlob{})) {
		return PCPKeyBlobType20, nil
	}

	if hdr.PcpType == pcptype_tpm20 && hdr.CbHeader >= uint32(unsafe.Sizeof(pcpKeyBlobWin8{})) {
		return PCPKeyBlobTypeWin8, nil
	}

	return PCPKeyBlobTypeUnknown, nil
}

func toKeyBlobWin8(blob []byte) (*pcpKeyBlobWin8, error) {
	out := pcpKeyBlobWin8{}
	if err := binary.Read(bytes.NewReader(blob), binary.LittleEndian, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func to20KeyBlob(blob []byte) (*pcp20KeyBlob, error) {
	out := pcp20KeyBlob{}
	if err := binary.Read(bytes.NewReader(blob), binary.LittleEndian, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func ExtractTPMTPublic(blob []byte) (*tpm2.Public, error) {
	subType, err := detectPcpKeyBlobSubType(blob)
	if err != nil {
		return nil, err
	}

	var keyBlob *pcpKeyBlobWin8
	switch subType {
	case PCPKeyBlobType12:
		return nil, ErrTpm12KeyBlob
	case PCPKeyBlobTypeWin8:
		kb, err := toKeyBlobWin8(blob)
		if err != nil {
			return nil, err
		}
		keyBlob = kb
	case PCPKeyBlobType20:
		kb, err := to20KeyBlob(blob)
		if err != nil {
			return nil, err
		}
		keyBlob = &kb.pcpKeyBlobWin8
	default:
		return nil, ErrInvalidKeyBlob

	}

	// TPM2B_PUBLIC is expected and has uint16 size prefixed to TPMT_PUBLIC, DecodePublic wants TPMT_PUBLIC
	pub, err := tpm2.DecodePublic(blob[keyBlob.CbHeader+2 : keyBlob.CbHeader+keyBlob.CbPublic])
	if err != nil {
		return nil, err
	}
	return &pub, nil
}
