// Copyright 2020 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package intelme

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	FWUpdateSuccess     = 0x0
	blackListMaxSize    = 10
	ipuSupportedMaxSize = 4
)

type fwupVersion struct {
	Major  uint16
	Minor  uint16
	Hotfix uint16
	Build  uint16
}

const (
	GetVersionCommand uint8 = iota
	GetVersionCommandReply
	StartUpdateCommand
	StartUpdateCommandReply
	SendUpdateDataCommand
	SendUpdateDataCommandReply
	EndUpdateCommand
	EndUpdateCommandReply
	GetInfoCommand
	GetInfoCommandReply
	GetFeatureStateCommand
	GetFeatureStateCommandReply
	GetFeatureCapabilityCommand
	GetFeatureCapabilityCommandReply
	GetPlatformTypeCommand
	GetPlatformTypeCommandReply
	VerifyOemIDCommand
	VerifyOemIDCommandReply
	GetOemIDCommand
	GetOemIDCommandReply
	ImageCompatabilityCheckCommand
	ImageCompatabilityCheckCommandReply
	GetUpdateDataExtensionCommand
	GetUpdateDataExtensionCommandReply
	GetRestorePointImageCommand
	GetRestorePointImageCommandReply
	GetIPUPTAttributeCommand
	GetIPUPTAttributeCommandReply
	GetInfoStatusCommand
	GetInfoStatusCommandReply
	GetMEInfoCommand
	GetMEInfoCommandReply
)

type fwupOEMID struct {
	Data1 uint64
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type fwupBlackListEntry struct {
	ExpressionType uint16
	MinorVersion   uint16
	HotfixVersion1 uint16
	BuildVersion1  uint16
	HotfixVersion2 uint16
	BuildVersion2  uint16
}

type fwupFlags struct {
	RecoveryMode      uint32 // 0 = No recovery; 1 = Full Recovery Mode,2 = Partial Recovery Mode (unused at present)
	IpuNeeded         uint32 // IPU_NEEDED bit, if set we are in IPU_NEEDED state.
	FWInitDone        uint32 // If set indicate FW is done initialized
	FWUInProgress     uint32 // If set FWU is in progress, this will be set for IFU update as well
	SafeFWUInprogress uint32 // If set IFU Safe FW update is in progress.
	NewFWTestState    uint32 // If set indicate that the new FT image is in Test Needed state (Stage 2 Boot)
	SafeBootCount     uint32 // Boot count before the operation is success
	ForceSafeBoot     uint32 // Force Safe Boot Flag, when this bit is set, we'll boot kernel only and go into recovery mode
	LivePingNeeded    uint32 // Use for IFU only, See Below
	// FWU tool needs to send Live-Ping or perform querying to confirm update successful.
	// With the current implementation when LivePingNeeded is set,
	// Kernel had already confirmed it. No action from the tool is needed.
	ResumeUpdateNeeded uint32 // Use for IFU only, If set FWU tool needs to resend update image
	RollbackNeededMode uint32 // FWU_ROLLBACK_NONE = 0, FWU_ROLLBACK_1, FWU_ROLLBACK_2
	// If not FWU_ROLLBACK_NONE, FWU tool needs to send restore_point image.
	ResetNeeded uint32 // When this field is set to ME_RESET_REQUIRED, FW Kernel will
	// perform ME_RESET after this message. No action from the tool is needed.
	Reserved uint32
}

type fwupPTAttributes struct {
	PTName             uint32
	LoadAddress        uint32
	FirmwareVersion    fwupVersion
	CurrentInstID      uint32
	CurrentUPVVersion  uint32
	ExpectedInstID     uint32
	ExpectedUPVVersion uint32
	Reserved           [16]byte
}

const (
	UpdateDisabled uint32 = iota
	UpdateEnabled
	UpdatePassword
)

type FWUpMEInfo struct {
	StructSize         uint32
	APIVersion         uint32
	FTPVersion         fwupVersion
	NFTPVersion        fwupVersion
	ChipsetVersion     uint32
	GlobalChipID       uint32
	SystemManufacturer [32]byte
	UpdateConfig       uint32 // 0= Disable, 1 = enable , 2 = PW protected
	HWFeatures         uint32
	FWFeatures         uint32
	LastFwUpdateStatus uint32 // Last FW update status
	DataFormatVersion  uint32 // Data format version Major(31:16), Minor (15:0). Only Major is used
	SVN                uint32 // Security version: Major (31:16), Minor (15:0),Only Major is used
	VCN                uint32 // Version Control Number: Major (31:16), Minor (15:0),Only Major is used
	MEBXVersion        fwupVersion
	FWUFlags           fwupFlags
	PlatformType       uint32
	OEMUUID            fwupOEMID
	FirmwareSize       uint16  // Size of FW image in multiple of .5MB
	History            [4]byte // Keep track of version tree history
	// Minor0Predecessor, Minor1Predecessor, Minor2Predecessor, Minor3Predecessor
	// FWU will check to see if the update image has the same predecessor with the
	// one that is already in the flash before allow the update.
	CVEDescriptor uint32
	BlackList     [blackListMaxSize]fwupBlackListEntry
	NumberOfIPUs  uint16
	//IPUEntry      [ipuSupportedMaxSize]fwupPTAttributes
}

type FWUpGetMEInfoResponse struct {
	Header FWupHdr
	Status uint32
	Info   FWUpMEInfo
}

func DecodeFWUPGetMEInfo(buf []byte) (*FWUpGetMEInfoResponse, error) {
	var resp FWUpGetMEInfoResponse
	minlen := len(resp.Header)
	minlen += 4
	if len(buf) < minlen {
		return nil, fmt.Errorf("size mismatch, want a minimum of %d bytes, got %d", minlen, len(buf))
	}
	copy(resp.Header[:], buf[:4])
	if len(buf) == minlen {
		// don't parse the rest, we got a partial response
		return &resp, nil
	}
	reader := bytes.NewReader(buf)
	reader.Seek(4, 0)
	if err := binary.Read(reader, binary.LittleEndian, &resp.Status); err != nil {
		return nil, fmt.Errorf("couldn't parse command status")
	}
	if resp.Status != 0 {
		return nil, fmt.Errorf("GetMEInfoResponseFromBytes failed")
	}
	if err := binary.Read(reader, binary.LittleEndian, &resp.Info); err != nil {
		return nil, fmt.Errorf("couldn't parse firmware version")
	}
	return &resp, nil
}

func EncodeFWUpGetMEInfo() []byte {
	var hdr FWupHdr
	hdr.SetMessageID(GetMEInfoCommand)

	return hdr[:]
}
