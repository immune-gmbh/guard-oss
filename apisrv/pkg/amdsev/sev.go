package amdsev

import (
	"bytes"
	"encoding/binary"
)

const (
	Success uint32 = iota
	InvalidPlatformState
	InvalidGuestState
	InvalidConfig
	InvalidLength
	AlreadyOwned
	InvalidCertificate
	PolicyFailure
	Inactive
	InvalidAddress
	BadSignature
	BadMeasurement
	ASIDOwned
	InvalidASID
	WBINVDRequired
	DFlushRequired
	InvalidGuest
	InvalidCommand
	Active
	HWSEVRetPlatform
	HWSEVRetUnsafe
	Unsupported
)

const (
	FactoryReset uint32 = iota
	PlatformStatus
	PEKGen
	PEKCsr
	PDHGen
	PDHCertExport
	PEKCertImport
	GetID
	GetID2
)

const (
	PlatformStatusReadLength = 12
)

const (
	ghcb            = 0xc0010130
	smm             = 0xc0010015
	sevCPUID        = 0x8000001f
	sevIOCType      = 'S'
	sevIssueCmdSize = 16
)

type PlatformStatusData struct {
	ApiMajor byte
	ApiMinor byte
	State    byte
	Owner    byte
	ConfigES byte
	Rsvd1    byte
	Rsvd2    byte
	Build    byte
	Guests   uint32
}

func DecodePlatformStatus(data []byte) (*PlatformStatusData, error) {
	var ps PlatformStatusData
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.LittleEndian, &ps); err != nil {
		return nil, err
	}
	return &ps, nil
}
