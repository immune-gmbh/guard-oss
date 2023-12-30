// Copyright 2020 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package intelme

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrFormat = errors.New("wrong payload")
	ErrHeader = errors.New("wrong header")
)

// MKHI command groups and commands
const (
	MKHIGroupCBM         = 0x00
	MKHIGroupPWD         = 0x02
	MKHIGroupFWCaps      = 0x03
	MKHIGroupHMRFPO      = 0x05
	MKHIGroupMCA         = 0x0a
	MKHIGroupSecureBoot  = 0x0c
	MKHIGroupNM          = 0x11
	MKHIGroupOSBUPCommon = 0xf0
	BUPGroupICC          = 0xf1
	BUPGroupMPHY         = 0xf2
	BUPGroupBIOSAR       = 0xf4
	MKHIGroupGen         = 0xff
)

// MKHI command and reponse codes
const (
	// MKHI Group CBM
	GlobalReset = 0x0b

	// MKHI Group PWD
	PWGMGRIsModified = 0x03

	// MKHI Group FWCaps
	GetFWCaps                    = 0x02
	SetFWCaps                    = 0x03
	FWCapsRule                   = 0x00000000
	MEDisable                    = 0x00000006
	LocalFWUpdate                = 0x00000007
	UserCapsState                = 0x00000009
	SetOEMSKURule                = 0x0000001c
	PlatformType                 = 0x0000001d
	FWFeatureState               = 0x00000020
	OEMTag                       = 0x0000002b
	ACMTPMData                   = 0x0000002f
	UnconfigureOnRTCClearDisable = 0x00000030
	AMTBIOSSyncInfo              = 0x00030005

	// MKHI Group HMRFPO
	HMRFPOEnable    = 0x01
	HMRFPOLock      = 0x02
	HMRFPOGetStatus = 0x03

	// MKHI Group MCA (called MCHI since Ice Lake)
	ReadFile       = 0x02
	SetFile        = 0x03
	CommitFile     = 0x04
	CoreBIOSDone   = 0x05
	GetRPMCStatus  = 0x08
	ReadFileEx     = 0x0a
	SetFileEx      = 0x0b
	ARBHSVNCommit  = 0x1b
	ARBHSVNGetInfo = 0x1c

	// MKHI Group Secure Boot Commands
	VerifyManifest = 0x01

	// MKHI Group Node Manager
	HostConfiguration = 0x00

	// MKHI Group OS BUP Common
	DRAMInitDone                          = 0x01
	MBPRequest                            = 0x02
	MESoftEnable                          = 0x03
	HMRFPODisable                         = 0x04
	GetIMRSize                            = 0x0c
	SetBringupManufacturingMEResetAndHalt = 0x0e
	GetERLog                              = 0x1b
	SetEDebugModeState                    = 0x1e
	DataClear                             = 0x20
	GetDebugTokenData                     = 0x22

	// MKHI Group MPHY
	ReadFromMPHY = 0x02

	// MKHI Group Gen
	GetFWVersion                       = 0x02
	EndOfPOST                          = 0x0c
	GetMEUnconfigState                 = 0x0e
	SetManufacturingMEResetAndHalt     = 0x10
	FWFeatureShipmentTimeStateOverride = 0x14
	GetImageFirmwareVersion            = 0x1c
	GetFIPSData                        = 0x21
	SetMeasuredBootState               = 0x22
	GetMeasuredBootState               = 0x23
)

func encodeMKHI(group uint8, command uint8, payload []byte) []byte {
	var msg MkhiHdr
	msg.SetGroupID(group)
	msg.SetCommand(command)

	if len(payload) == 0 {
		return msg[:]
	}

	buf := make([]byte, len(msg)+len(payload))
	copy(buf, msg[:])
	copy(buf[len(msg):], payload)

	return buf
}

func decodeMKHI(ctx context.Context, group uint8, command uint8, buf []byte) ([]byte, error) {
	var hdr MkhiHdr

	if len(buf) < len(hdr) {
		tel.Log(ctx).WithField("response", len(buf)).WithField("header", len(hdr)).Error("mkhi response too short")
		return nil, ErrHeader
	}
	copy(hdr[:], buf[:len(hdr)])

	if hdr.GroupID() != group {
		tel.Log(ctx).WithField("response", hdr.GroupID()).WithField("expected", group).Error("mkhi wrong group")
		return nil, ErrHeader
	}
	if hdr.Command() != command {
		tel.Log(ctx).WithField("response", hdr.Command()).WithField("expected", command).Error("mkhi wrong command")
		return nil, ErrHeader
	}
	if !hdr.IsResponse() {
		tel.Log(ctx).Error("not a response message")
		return nil, ErrHeader
	}
	if hdr.Result() != 0 {
		tel.Log(ctx).WithField("code", hdr.Result()).Error("command failed")
		return nil, ErrHeader
	}

	return buf[len(hdr):], nil
}

type GetFWCapsResponse struct {
	Header     MkhiHdr
	Rule       uint32
	DataLength uint8
}

const GetFWCapsResponseLength = 4 + 4 + 1

func encodeGetFWCapsRule(rule uint32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, rule)
	return encodeMKHI(MKHIGroupFWCaps, GetFWCaps, buf.Bytes())
}

func decodeGetFWCapsRule(ctx context.Context, rule int, resp interface{}, resplen int, b []byte) error {
	data, err := decodeMKHI(ctx, MKHIGroupFWCaps, GetFWCaps, b)
	if err != nil {
		return err
	}

	var r uint32
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &r)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read FWCapsRule")
		return ErrFormat
	}
	if uint32(rule) != r {
		tel.Log(ctx).WithField("expected", uint32(rule)).WithField("response", r).Error("wrong rule")
		return ErrFormat
	}

	ruleLen := int(data[4])
	if ruleLen == 0 || len(data) < ruleLen+5 {
		tel.Log(ctx).WithField("ruleLen", ruleLen).WithField("dataLen", len(data)).Error("wrong ruleLen")
		return ErrFormat
	}
	err = binary.Read(bytes.NewReader(data[5:]), binary.LittleEndian, resp)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read FWCapsRule data")
		return ErrFormat
	}

	return nil
}

type GetFWCapsDWordResponse struct {
	Value uint32
}

func decodeGetFWCapsDWordRule(ctx context.Context, rule int, b []byte) (*GetFWCapsDWordResponse, error) {
	var resp GetFWCapsDWordResponse
	err := decodeGetFWCapsRule(ctx, rule, &resp, 4, b)
	return &resp, err
}

const (
	FeatureFullNetwork     = 1 << 0
	FeatureStandardNetwork = 1 << 1
	FeatureAMT             = 1 << 2
	// Reserved
	FeatureIIT = 1 << 4
	// Reserved
	FeatureCLS = 1 << 6
	// Reserved
	// Reserved
	// Reserved
	FeatureISH = 1 << 10
	// Reserved
	FeaturePAVP = 1 << 12
	// Reserved
	// Reserved
	// Reserved
	// Reserved
	FeatureIPv6 = 1 << 17
	FeatureKVM  = 1 << 18
	// Reserved
	FeatureDAL = 1 << 20
	FeatureTLS = 1 << 21
	// Reserved
	FeatureWLAN = 1 << 23
	// Reserved
	// Reserved
	// Reserved
	// Reserved
	// Reserved
	FeaturePTT = 1 << 29
	// Reserved
	// Reserved
)

func EncodeGetFeatureCaps() []byte {
	return encodeGetFWCapsRule(FWCapsRule)
}

func DecodeGetFeatureCaps(ctx context.Context, b []byte) (*GetFWCapsDWordResponse, error) {
	return decodeGetFWCapsDWordRule(ctx, FWCapsRule, b)
}

func EncodeGetLocalFirmwareUpdateCap() []byte {
	return encodeGetFWCapsRule(LocalFWUpdate)
}

func DecodeGetLocalFirmwareUpdateCap(ctx context.Context, b []byte) (*GetFWCapsDWordResponse, error) {
	return decodeGetFWCapsDWordRule(ctx, LocalFWUpdate, b)
}

func EncodeGetAMTCaps() []byte {
	return encodeGetFWCapsRule(UserCapsState)
}

func DecodeGetAMTCaps(ctx context.Context, b []byte) (*GetFWCapsDWordResponse, error) {
	return decodeGetFWCapsDWordRule(ctx, UserCapsState, b)
}

const (
	// UsageType
	MobileUsageType      = 1
	DesktopUsageType     = 2
	ServerUsageType      = 4
	WorkstationUsageType = 8

	// ImageType
	NoSKU        = 0
	ConsumerSKU  = 3
	CorporateSKU = 4

	// PlatformBrand
	AMTBrand                  = 1
	StandardManagabilityBrand = 2
)

type AMTType struct {
	UsageType     int
	IsSuperSKU    bool
	ImageType     int
	PlatformBrand int
}

func EncodeGetAMTType() []byte {
	return encodeGetFWCapsRule(PlatformType)
}

func DecodeGetAMTType(ctx context.Context, b []byte) (*AMTType, error) {
	var bits uint32
	err := decodeGetFWCapsRule(ctx, PlatformType, &bits, 4+3, b)
	if err != nil {
		return nil, err
	}

	resp := AMTType{
		UsageType:     int(bits) & 0b1111,
		IsSuperSKU:    (bits>>6)&1 != 0,
		ImageType:     (int(bits) >> 8) & 0b1111,
		PlatformBrand: (int(bits) >> 12) & 0b1111,
	}
	return &resp, nil
}

func EncodeGetOEMTag() []byte {
	return encodeGetFWCapsRule(OEMTag)
}

func DecodeGetOEMTag(ctx context.Context, b []byte) ([]byte, error) {
	if len(b) < GetFWCapsResponseLength {
		tel.Log(ctx).Errorf("payload too short, got %v expect min %v", len(b), GetFWCapsResponseLength)
		return nil, ErrFormat
	}
	taglen := len(b) - GetFWCapsResponseLength
	tag := make([]byte, taglen)
	err := decodeGetFWCapsRule(ctx, OEMTag, &tag, taglen, b)
	return tag, err
}

func EncodeGetUnconfigureOnRTCClearDisabled() []byte {
	return encodeGetFWCapsRule(UnconfigureOnRTCClearDisable)
}

func DecodeGetUnconfigureOnRTCClearDisbaled(ctx context.Context, b []byte) (*GetFWCapsDWordResponse, error) {
	return decodeGetFWCapsDWordRule(ctx, UnconfigureOnRTCClearDisable, b)
}

func EncodeGetFeatureState() []byte {
	return encodeGetFWCapsRule(FWFeatureState)
}

func DecodeGetFeatureState(ctx context.Context, b []byte) (*GetFWCapsDWordResponse, error) {
	return decodeGetFWCapsDWordRule(ctx, FWFeatureState, b)
}

type HMRFPOEnableRequest struct {
	Mkhi  MkhiHdr
	Nonce uint64
}

const (
	EnableStatusSuccess        = 0x00
	EnableStatusLocked         = 0x01
	EnableStatusNvarFailure    = 0x02
	EnableStatusUnknownFailure = 0x05
)

type HMRFPOEnableResponse struct {
	Mkhi         MkhiHdr
	FactoryBase  uint32
	FactoryLimit uint32
	Status       uint8
	Reserved     [3]byte
}

func EncodeHMRFPOEnable() []byte {
	var msg HMRFPOEnableRequest
	msg.Mkhi.SetGroupID(MKHIGroupHMRFPO)
	msg.Mkhi.SetCommand(HMRFPOEnable)

	writer := new(bytes.Buffer)
	_ = binary.Write(writer, binary.LittleEndian, msg)
	return writer.Bytes()
}

type HMRFPOGetStatusResponse struct {
	Header   MkhiHdr
	Status   uint8
	Reserved [3]byte
}

func EncodeHMRFPOGetStatus() []byte {
	var mkhi MkhiHdr
	mkhi.SetGroupID(MKHIGroupHMRFPO)
	mkhi.SetCommand(HMRFPOGetStatus)

	writer := new(bytes.Buffer)
	_ = binary.Write(writer, binary.LittleEndian, mkhi)
	return writer.Bytes()
}

func DecodeHMRFPOGetStatus(ctx context.Context, b []byte) (*HMRFPOGetStatusResponse, error) {
	var resp HMRFPOGetStatusResponse
	err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &resp)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read HMRFPOGetStatusResponse")
		return nil, ErrFormat
	}
	if resp.Header.GroupID() != MKHIGroupHMRFPO {
		tel.Log(ctx).Errorf("wrong group ID, want %d, got %d", MKHIGroupHMRFPO, resp.Header.GroupID())
		return nil, ErrHeader
	}
	if resp.Header.Command() != HMRFPOGetStatus {
		tel.Log(ctx).Errorf("wrong command ID, want %d, got %d", HMRFPOGetStatus, resp.Header.Command())
		return nil, ErrHeader
	}
	if !resp.Header.IsResponse() {
		tel.Log(ctx).Errorf("not a response message")
		return nil, ErrHeader
	}

	return &resp, nil
}

type ReadFileRequest struct {
	Mkhi     MkhiHdr
	Filepath [64]byte
	Offset   uint32
	Size     uint32
	Flags    uint8
}

type ReadFileResponse struct {
	Mkhi MkhiHdr
	Size uint32
	Data []byte
}

func EncodeReadFile(filepath string, offset uint32, size uint32) ([]byte, error) {
	var mkhi MkhiHdr
	mkhi.SetGroupID(MKHIGroupMCA)
	mkhi.SetCommand(ReadFile)

	if size > 4096 {
		return nil, fmt.Errorf("max read len is 4096, got: %v", size)
	}

	msg := ReadFileRequest{
		Mkhi:   mkhi,
		Offset: offset,
		Size:   size,
		Flags:  0,
	}
	copy(msg.Filepath[:], []byte(filepath))
	writer := new(bytes.Buffer)
	if err := binary.Write(writer, binary.LittleEndian, msg); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func DecodeReadFile(b []byte) (*ReadFileResponse, error) {
	var resp ReadFileResponse
	reader := bytes.NewReader(b)
	if err := binary.Read(reader, binary.LittleEndian, &resp.Mkhi); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &resp.Size); err != nil {
		return nil, err
	}
	resp.Data = make([]byte, resp.Size)
	if err := binary.Read(reader, binary.LittleEndian, &resp.Data); err != nil {
		return nil, err
	}
	return &resp, nil
}

const (
	UsagePMC   = 2
	UsageRBE   = 3
	UsageROTKM = 5
	UsageBSMM  = 32
	UsageOEMKM = 45
	UsageDnX   = 53
	UsageBGKM  = 54
	UsageACM   = 56

	SVNEnabled = 0b00001000
	SVNValid   = 0b00010000
)

type ARBHSVN struct {
	Usage           string
	BootPartitionID int
	Enabled         bool
	Valid           bool
	ExecutingSVN    int
	MinAllowedSVN   int
}

func EncodeARBHSVNGetInfo() []byte {
	return encodeMKHI(MKHIGroupMCA, ARBHSVNGetInfo, nil)
}

func DecodeARBHSVNGetInfo(ctx context.Context, b []byte) ([]ARBHSVN, error) {
	data, err := decodeMKHI(ctx, MKHIGroupMCA, ARBHSVNGetInfo, b)
	if err != nil {
		return nil, err
	}

	var num uint32
	rd := bytes.NewReader(data)
	err = binary.Read(rd, binary.LittleEndian, &num)
	if err != nil || num > 128 {
		return nil, ErrFormat
	}

	parts := make([]ARBHSVN, num)
	for i := 0; uint32(i) < num; i += 1 {
		var bs [4]byte
		err = binary.Read(rd, binary.LittleEndian, &bs)
		if err != nil {
			return nil, ErrFormat
		}

		switch bs[0] {
		case UsagePMC:
			parts[i].Usage = "pmc"
		case UsageRBE:
			parts[i].Usage = "rbe"
		case UsageROTKM:
			parts[i].Usage = "rot-km"
		case UsageBSMM:
			parts[i].Usage = "bsmm"
		case UsageOEMKM:
			parts[i].Usage = "oem-km"
		case UsageDnX:
			parts[i].Usage = "dnx"
		case UsageBGKM:
			parts[i].Usage = "bg-km"
		case UsageACM:
			parts[i].Usage = "acm"
		default:
			parts[i].Usage = "unknown"
		}

		parts[i].Enabled = (bs[1]>>3)&1 == 0
		parts[i].Valid = (bs[1]>>4)&1 == 0
		parts[i].BootPartitionID = int(bs[1]) & 0b111
		parts[i].ExecutingSVN = int(bs[2])
		parts[i].MinAllowedSVN = int(bs[3])
	}

	return parts, nil
}

type RPMCStatus struct {
	Device          string
	Support         bool
	Bound           bool
	AllowsRebinding bool
	Rebindings      int
	MaxRebinds      int
	Counters        int
	Chipselect      int
	FatalError      int
}

func EncodeRPMCStatus() []byte {
	return encodeMKHI(MKHIGroupMCA, GetRPMCStatus, nil)
}

func DecodeRPMCStatus(ctx context.Context, buf []byte) (*RPMCStatus, error) {
	data, err := decodeMKHI(ctx, MKHIGroupMCA, GetRPMCStatus, buf)
	if err != nil {
		return nil, err
	}

	var status uint32
	rd := bytes.NewReader(data)
	err = binary.Read(rd, binary.LittleEndian, &status)
	if err != nil {
		return nil, ErrFormat
	}

	var ret RPMCStatus
	switch status & 0b11 {
	case 0b10:
		ret.Device = "spi"
	case 0b11:
		ret.Device = "none"
	default:
		ret.Device = "unknown"
	}

	ret.Support = ((status >> 2) & 1) != 0
	ret.Bound = ((status >> 3) & 1) != 0
	ret.AllowsRebinding = ((status >> 4) & 1) != 0
	ret.Rebindings = int((status >> 5) & 0b11111)
	ret.MaxRebinds = int((status >> 10) & 0b1111)
	ret.Counters = int((status >> 14) & 0b111)
	ret.Chipselect = int((status >> 17) & 0b1)
	ret.FatalError = int((status >> 18) & 0b11111111)

	return &ret, nil
}

type FITCVersion struct {
	MinorFITC       uint16
	MajorFITC       uint16
	BuildNumberFITC uint16
	HotFixFITC      uint16
}
type FirmwareVersion struct {
	MinorCode           uint16
	MajorCode           uint16
	BuildNumberCode     uint16
	HotFixCode          uint16
	MinorRecovery       uint16
	MajorRecovery       uint16
	BuildNumberRecovery uint16
	HotFixRecovery      uint16
}

func EncodeGetFirmwareVersion() []byte {
	return encodeMKHI(MKHIGroupGen, GetFWVersion, nil)
}

func DecodeGetFirmwareVersion(ctx context.Context, b []byte) (*FirmwareVersion, *FITCVersion, error) {
	data, err := decodeMKHI(ctx, MKHIGroupGen, GetFWVersion, b)
	if err != nil {
		return nil, nil, err
	}
	if (len(data) != 24) && (len(data) != 16) {
		return nil, nil, ErrFormat
	}

	var fwver FirmwareVersion
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &fwver)
	if err != nil {
		return nil, nil, ErrFormat
	}

	if len(data) == 16 {
		return &fwver, nil, nil
	}

	var fitc FITCVersion
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &fitc)
	if err != nil {
		return nil, nil, ErrFormat
	}

	return &fwver, &fitc, nil
}

type flashPartitionRaw struct {
	Id        uint32
	Reserved1 uint64
	Version   [4]uint16
	VendorId  uint32
	SVN       uint32
	Reserved2 [60]byte
}

type FlashPartition struct {
	Id      string
	Version []int
	Vendor  string
	SVN     int
}

func EncodeGetFirmwareImageVersion() []byte {
	return encodeMKHI(MKHIGroupGen, GetImageFirmwareVersion, make([]byte, 4))
}

func DecodeGetFirmwareImageVersion(ctx context.Context, b []byte) ([]FlashPartition, error) {
	data, err := decodeMKHI(ctx, MKHIGroupGen, GetImageFirmwareVersion, b)
	if err != nil {
		return nil, err
	}

	var num uint32
	rd := bytes.NewReader(data)
	err = binary.Read(rd, binary.LittleEndian, &num)
	if err != nil || num > 128 {
		return nil, ErrFormat
	}

	parts := make([]FlashPartition, num)
	for i := 0; uint32(i) < num; i += 1 {
		var raw flashPartitionRaw
		err = binary.Read(rd, binary.LittleEndian, &raw)
		if err != nil {
			return nil, ErrFormat
		}

		switch raw.Id {
		// FTPR
		case 0x52505446:
			parts[i].Id = "ftpr"
		// RBEP
		case 0x50454252:
			parts[i].Id = "rbe"
		// LOCL
		case 0x4c434f4c:
			parts[i].Id = "locl"
		// WCOD
		case 0x444f4357:
			parts[i].Id = "wcod"
		// OEMP
		case 0x504d454f:
			parts[i].Id = "oem"
		// PMCP
		case 0x50434d50:
			parts[i].Id = "pmc"
			// ISHC
		case 0x43485349:
			parts[i].Id = "ish"
		// IOMP
		case 0x504d4f49:
			parts[i].Id = "iom"
		// NPHY
		case 0x5948504e:
			parts[i].Id = "phy"
		// TBTP
		case 0x50544254:
			parts[i].Id = "thunderbold"
		// PCHC
		case 0x43484350:
			parts[i].Id = "pchc"
		default:
			parts[i].Id = fmt.Sprintf("0x%x", raw.Id)
		}
		parts[i].Vendor = fmt.Sprintf("%x", raw.VendorId)
		parts[i].Version = []int{
			int(raw.Version[0]),
			int(raw.Version[1]),
			int(raw.Version[2]),
			int(raw.Version[3]),
		}
		parts[i].SVN = int(raw.SVN)
	}

	return parts, nil
}

type fipsDataRaw struct {
	Mode     uint32
	Version  uint64
	Reserved uint64
}

type FIPSData struct {
	Enabled bool
	Version uint64
}

func EncodeGetFIPSData() []byte {
	return encodeMKHI(MKHIGroupGen, GetFIPSData, nil)
}

func DecodeGetFIPSData(ctx context.Context, b []byte) (*FIPSData, error) {
	data, err := decodeMKHI(ctx, MKHIGroupGen, GetFIPSData, b)
	if err != nil {
		return nil, err
	}
	if len(data) != 20 {
		return nil, ErrFormat
	}

	var payload fipsDataRaw
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &payload)
	if err != nil {
		return nil, ErrFormat
	}

	ret := FIPSData{
		Enabled: payload.Mode == 1,
		Version: payload.Version,
	}

	return &ret, nil
}

func EncodeGetMeasuredBootState() []byte {
	return encodeMKHI(MKHIGroupGen, GetMeasuredBootState, nil)
}

func DecodeGetMeasuredBootState(ctx context.Context, b []byte) (bool, error) {
	data, err := decodeMKHI(ctx, MKHIGroupGen, GetMeasuredBootState, b)
	if err != nil {
		return false, err
	}
	if len(data) != 1 {
		return false, ErrFormat
	}
	return data[0] == 1, nil
}
