package configuration

import (
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/amdsev"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/intelme"
)

var (
	DefaultConfiguration = api.Configuration{
		Root: api.RootECC,
		Keys: map[string]api.KeyTemplate{
			"aik": api.QuoteECC,
		},
		PCRBank: uint16(tpm2.AlgSHA256),
		PCRs:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
		TPM2NVRAM: []uint32{
			uint32(0x1C10103),
			uint32(0x1800001),
			uint32(0x1C10102),
			uint32(0x1800003),
			uint32(0x1C10106),
			uint32(0x1400001),
		},
		UEFIVariables: []api.UEFIVariable{
			{Vendor: api.EFIGlobalVariable.String(), Name: "SetupMode"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "AuditMode"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "DeployedMode"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "SecureBoot"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "PK"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "PKDefault"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "KEK"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "KEKDefault"},
			{Vendor: api.EFIImageSecurityDatabase.String(), Name: "db"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "dbDefault"},
			{Vendor: api.EFIImageSecurityDatabase.String(), Name: "dbx"},
			{Vendor: api.EFIGlobalVariable.String(), Name: "dbxDefault"},
		},
		MSRs: []api.MSR{
			{MSR: api.MSRSMBase},
			{MSR: api.MSRMTRRCap},
			{MSR: api.MSRSMRRPhysBase},
			{MSR: api.MSRSMRRPhysMask},
			{MSR: api.MSRFeatureControl},
			{MSR: api.MSRPlatformID},
			{MSR: api.MSRIA32DebugInterface},
			{MSR: api.MSRK8Sys},
			{MSR: api.MSREFER},
		},
		CPUIDLeafs: []api.CPUIDLeaf{
			{LeafEAX: 0, LeafECX: 0},
			{LeafEAX: 1, LeafECX: 0},
			{LeafEAX: 0x80000002, LeafECX: 0},
			{LeafEAX: 0x80000003, LeafECX: 0},
			{LeafEAX: 0x80000004, LeafECX: 0},
			{LeafEAX: api.CPUIDExtendedFeatureFlags, LeafECX: 0},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 0},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 1},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 2},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 3},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 4},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 5},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 6},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 7},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 8},
			{LeafEAX: api.CPUIDSGXCapabilities, LeafECX: 9},
			{LeafEAX: api.CPUIDSEV, LeafECX: 0},
		},
		SEV: []api.SEVCommand{
			{Command: amdsev.PlatformStatus, ReadLength: amdsev.PlatformStatusReadLength},
		},
		ME: []api.MEClientCommands{
			{
				GUID:    &intelme.CSME_MKHIGuid,
				Address: "",
				Commands: []api.MECommand{
					{Command: intelme.EncodeGetFirmwareVersion()},
					{Command: intelme.EncodeGetFeatureState()},
					{Command: intelme.EncodeGetFeatureCaps()},
					{Command: intelme.EncodeGetLocalFirmwareUpdateCap()},
					{Command: intelme.EncodeGetAMTCaps()},
					{Command: intelme.EncodeGetAMTType()},
					{Command: intelme.EncodeGetOEMTag()},
					{Command: intelme.EncodeGetUnconfigureOnRTCClearDisabled()},
					{Command: intelme.EncodeHMRFPOEnable()},
					{Command: intelme.EncodeHMRFPOGetStatus()},
					{Command: intelme.EncodeGetFIPSData()},
					{Command: intelme.EncodeGetMeasuredBootState()},
					{Command: intelme.EncodeGetFirmwareImageVersion()},
				},
			}, {
				GUID:    &intelme.SPS_MKHIGuid,
				Address: "7",
				Commands: []api.MECommand{
					{Command: intelme.EncodeGetFirmwareVersion()},
				},
			}, {
				GUID:    &intelme.FWUpdateGuid,
				Address: "",
				Commands: []api.MECommand{
					{Command: intelme.EncodeFWUpGetMEInfo()},
				},
			}, {
				GUID:    &intelme.MCHIGuid1,
				Address: "",
				Commands: []api.MECommand{
					{Command: intelme.EncodeRPMCStatus()},
					{Command: intelme.EncodeMCHIARBHSVNGetInfo()},
					{Command: intelme.EncodeHasUPID()},
				},
			}, {
				GUID:    &intelme.AMTGuid3,
				Address: "",
				Commands: []api.MECommand{
					{Command: intelme.EncodeGetAMTState()},
					// The order of the following two commands is important
					{Command: intelme.EncodeGetAMTAuditLogRecords()},
					{Command: intelme.EncodeGetAMTAuditLogSignature()},
					{Command: intelme.EncodeGetAMTMESetupAuditRecord()},
					{Command: intelme.EncodeGetAMTProvisioningMode()},
					{Command: intelme.EncodeGetAMTSecurityParameters()},
					{Command: intelme.EncodeGetAMTState()},
					{Command: intelme.EncodeGetAMTCodeVersions()},
					{Command: intelme.EncodeGetAMTZeroTouchEnabled()},
					{Command: intelme.EncodeGetAMTRedirectionSessionState()},
					{Command: intelme.EncodeGetAMTKVMSessionState()},
					{Command: intelme.EncodeGetAMTWebUIState()},
					{Command: intelme.EncodeGetAMTRemoteAccessConnectionState()},
				},
			},
		},
		TPM2Properties: []api.TPM2Property{
			{Property: uint32(tpm2.Manufacturer)},
			{Property: uint32(tpm2.VendorString1)},
			{Property: uint32(tpm2.VendorString2)},
			{Property: uint32(tpm2.VendorString3)},
			{Property: uint32(tpm2.VendorString4)},
			{Property: uint32(tpm2.SpecLevel)},
			{Property: uint32(tpm2.SpecRevision)},
			{Property: uint32(tpm2.SpecDayOfYear)},
			{Property: uint32(tpm2.SpecYear)},
		},
		PCIConfigSpaces: []api.PCIConfigSpace{
			{Bus: 0, Device: 0, Function: 0}, // Host controller
			{Bus: 0, Device: intelme.PCIDevice, Function: intelme.PCIFunction},
			{Bus: 0, Device: 0x1f, Function: 5}, // SPI controller
		},
	}
	DefaultConfigurationModTime = func() time.Time {
		// 02 Jan 06 15:04 MST
		ts, err := time.Parse(time.RFC822, "24 Mar 22 16:46 GMT+2")
		if err != nil {
			panic(err)
		} else {
			return ts
		}
	}()
)
