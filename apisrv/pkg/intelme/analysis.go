package intelme

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	ResetWorkingState           = 0x00
	InitializingWorkingState    = 0x01
	RecoveryWorkingState        = 0x02
	NormalWorkingState          = 0x05
	DisableWaitWorkingState     = 0x06
	StateTransitionWorkingState = 0x07
	InvalidStateWorkingState    = 0x08
	HaltWorkingState            = 0x0e

	PrebootOperationState           = 0b000
	M0withUMAOperationState         = 0b001
	M0PowerGatedOperationState      = 0b010
	M3withoutUMAOperationState      = 0b100
	M0withoutUMAOperationState      = 0b101
	BringupOperationState           = 0b110
	M0withoutUMAErrorOperationState = 0b111

	NoError              = 0x00
	UncategorizedFailure = 0x01
	Disabled             = 0x02
	ImageFailure         = 0x03
	FatalError           = 0x04 // extended error status valid

	NormalOperationMode        = 0
	DebugOperationMode         = 2
	SoftDisableOperationMode   = 3
	JumperDisableOperationMode = 4
	ReflashOperationMode       = 5
	EnhancedDebugOperationMode = 7

	ACPower = 1
	DCPower = 2

	PCHCConfigError  = 0
	PCHCConfigOk     = 1
	OEMDataProgError = 2
	OEMDataProgOk    = 3

	BUP_STATUS_BEGIN                             = 0x00 // Initialization starts
	BUP_STATUS_DISABLE_HOST_WAKE_EVENT           = 0x01 // Disable the host wake event
	BUP_STATUS_CLOCK_GATING_ENABLED              = 0x02 // Enabling clock gating for Intel® ME
	BUP_STATUS_HOST_PM_HANDSHAKE_ENABLE          = 0x03 // Enabling PM ME handshaking
	BUP_STATUS_FLOW_DETERMINATION                = 0x04 // Flow determination start process
	BUP_STATUS_PMC_PATCHING                      = 0x05 // PMC patching process
	BUP_STATUS_GET_FLASH_VSCC                    = 0x06 // Get the flash VSCC parameters
	BUP_STATUS_SET_FLASH_VSCC                    = 0x07 // Set (program) the flash VSCC registers
	BUP_STATUS_VSCC_FAILURE                      = 0x08 // Error reading/matching the VSCC table in the descriptor
	BUP_STATUS_EFFS_INITIALIZATION               = 0x09 // Initialize EFFS (overlay and SDM memory management)
	BUP_STATUS_CHECK_CSE_STRAP_DISABLED          = 0x0a // Check to view if straps say Intel® ME DISABLED
	BUP_STATUS_T34_MISSING                       = 0x0b // Timeout waiting for T34
	BUP_STATUS_CHECK_STRAP_DISABLED              = 0x0c // Check to view if EFFS say Intel® ME DISABLED
	BUP_STATUS_CHECK_FLASH_OVERRIDE              = 0x0d // Possibly handle BUP manufacturing override strap
	BUP_STATUS_CHECK_CSE_POLICY_DISABLED         = 0x0e // Check to view if EFFS say Intel® ME DISABLED.
	BUP_STATUS_PKTPM_INITIALIZATION              = 0x0f // Initialize PKTPM
	BUP_STATUS_BLOB_INITIALIZATION               = 0x10 // Initialize MiniBLOB
	BUP_STATUS_CM3                               = 0x11 // Bring-up in CM3
	BUP_STATUS_CM0                               = 0x12 // Bring-up in CM0
	BUP_STATUS_FLOW_ERROR                        = 0x13 // Flow detection error
	BUP_STATUS_CM3_CLOCK_SWITCH                  = 0x14 // CM3 clock switching
	BUP_STATUS_CM3_CLOCK_SWITCH_ERROR            = 0x15 // CM3 clock switching error
	BUP_STATUS_CM3_FLASH_PAGING                  = 0x16 // CM3 flash paging flow
	BUP_STATUS_CM3_ICV_RECOVERY                  = 0x17 // CM3 ICV recovery flow. Host error - CPU reset timeout, DID timeout, memory missing
	BUP_STATUS_CM3_LOAD_KERNEL                   = 0x18 // CM3 kernel load
	BUP_STATUS_CM0_HOST_PREP                     = 0x19 // CM0 Host prep sequence
	BUP_STATUS_CM0_SKIP_HOST_PREP                = 0x1a // CM0 Host skip prep sequence
	BUP_STATUS_CM0_ICC_PROGRAMMING               = 0x1b // ICC programming
	BUP_STATUS_CM0_T34_ERROR                     = 0x1c // T34 missing – cannot program ICC
	BUP_STATUS_CM0_FLEX_SKU                      = 0x1d // FLEX SKU programming
	BUP_STATUS_CM0_TPM_START                     = 0x1e // TPM start (interface un-isolation)
	BUP_STATUS_CM0_DID_WAIT                      = 0x1f // Waiting for DID BIOS message
	BUP_STATUS_CM0_DID_ERROR                     = 0x20 // Waiting for DID BIOS message failure
	BUP_STATUS_CM0_DID_NOMEM                     = 0x21 // DID reported no error
	BUP_STATUS_CM0_UMA_ENABLE                    = 0x22 // Enabling UMA
	BUP_STATUS_CM0_UMA_ENABLE_ERROR              = 0x23 // Enabling UMA error
	BUP_STATUS_CM0_DID_ACK                       = 0x24 // Sending DID Ack to BIOS
	BUP_STATUS_CM0_DID_ACK_ERROR                 = 0x25 // Sending DID Ack to BIOS error
	BUP_STATUS_CM0_CLOCK_SWITCH                  = 0x26 // Switching clocks in M0
	BUP_STATUS_CM0_CLOCK_SWITCH_ERROR            = 0x27 // Switching clocks in M0 error
	BUP_STATUS_CM0_TEMP_DISABLE                  = 0x28 // Intel® ME in temp disable
	BUP_STATUS_CM0_TEMP_DISABLE_ERROR            = 0x29 // (Error) Intel® ME in temp disable
	BUP_STATUS_CM0_TEMP_DISABLE_UMA_ENABLE       = 0x2a // Intel® ME in temp disable - exiting and UMA enabling
	BUP_STATUS_CM0_IPK_CHECK                     = 0x2b // Intel® ME IPK check
	BUP_STATUS_CM0_IPK_RECREATION                = 0x2c // Intel® ME IPK recreation
	BUP_STATUS_CM0_IPK_RECREATION_ERROR          = 0x2d // Intel® ME IPK recreation error
	BUP_STATUS_CM0_UMA_VALIDATION                = 0x2e // Intel® ME UMA validation for resume
	BUP_STATUS_CM0_UMA_VALIDATION_ERROR          = 0x2f // Intel® ME UMA validation for resume error
	BUP_STATUS_CM0_UMA_VALIDATION_IPK            = 0x30 // Intel® ME UMA validation for resume error, so IPK recreation
	BUP_STATUS_CM0_PKVENOM_START                 = 0x31 // CM0 PK VENOM start
	BUP_STATUS_CM0_LOAD_IBLS                     = 0x32 // CM0 load IBLs
	BUP_STATUS_CM0_PK_FTPM_INIT                  = 0x33 // CM0 PK FTPM initial phase
	BUP_STATUS_CM0_PK_FTPM_ABORT                 = 0x34 // CM0 PK FTPM Abort
	BUP_STATUS_HALT_UPON_FIPS_BUP_ERR            = 0x35 // FIPS halt - BUP self-test error
	BUP_STATUS_HALT_UPON_FIPS_CRYPTO_DRV_ERR     = 0x36 // FIPS halt - CRYPTO_DRV self-test error
	BUP_STATUS_HALT_UPON_FIPS_TLS_ERR            = 0x37 // FIPS halt - TLS self-test error
	BUP_STATUS_HALT_UPON_FIPS_DT_ERR             = 0x38 // FIPS halt - DT self-test error
	BUP_STATUS_HALT_UPON_FIPS_UNKNOWN_ERR        = 0x39 // FIPS halt - DT self-test error
	BUP_STATUS_CM0_MEMORY_ACCESS_RANGE_ERR       = 0x3a // Error enabling memory access range
	BUP_STATUS_CSE_RESET_LIMIT_ERR               = 0x3b // Intel® ME reset limit reached
	BUP_STATUS_CSE_RESET_ENTER_RECOVERY          = 0x3c // Two Intel® ME resets within short time detected, enter recovery
	BUP_STATUS_HALT_UPON_UNKNOWN_ERROR           = 0x3d // Unknown halt reason detected, halt Intel® ME.
	BUP_STATUS_VALIDATE_NFT                      = 0x3e // Validating GLUT and NFT manifest
	BUP_STATUS_READ_FIXED_DATA                   = 0x3f // Reading kernel fixed data from NVAR
	BUP_STATUS_READ_ICC_DATA                     = 0x40 // Reading ICC data
	BUP_STATUS_ZERO_UMA                          = 0x41 // Zeroing out UMA
	BUP_STATUS_HALT_UPON_SKU_ERROR               = 0x42 // Full FW SKU running on Ignition (Slim) HW SKU
	BUP_STATUS_DERIVE_CHIPSET_KEY_ERROR          = 0x43 // Error when deriving chipset keys
	BUP_STATUS_HOST_ERROR                        = 0x44 // Bad BIOS, CPU DOA, CPU Missing
	BUP_STATUS_FTP_LOAD_ERROR                    = 0x46 // Failure in loading FTP
	BUP_STATUS_MFG_CMRST                         = 0x47 // Intel® ME halted in BUP from Mfg., ME reset
	BUP_STATUS_MPR_VIOLATION_CMRST               = 0x48 // Intel® ME was reset because of MPR protection violation
	BUP_STATUS_ICC_START_POLL_BEGIN              = 0x49 // START_ICC_CONFIG polling begins
	BUP_STATUS_ICC_START_POLL_END                = 0x4a // START_ICC_CONFIG polling end
	BUP_STATUS_HOBIT_SET                         = 0x4b // set HOBIT
	BUP_STATUS_POLL_CPURST_DEASSERT_BEGIN        = 0x4c // CPU_RST_DONE polling begins
	BUP_STATUS_CPURST_DEASSERT_DONE              = 0x4d // CPU_RST_DONE polling end
	BUP_STATUS_DID_POLL_BEGIN                    = 0x4e // Dram Initial Done polling begins
	BUP_STATUS_DID_RECVD                         = 0x4f // Dram Initial Done polling end
	BUP_STATUS_ZERO_UMA_BEGIN                    = 0x50
	BUP_STATUS_ZERO_UMA_DONE                     = 0x51
	BUP_STATUS_DID_ACK_SENT                      = 0x52 // Dram Initial Done Ack sent to BIOS
	BUP_STATUS_ICC_REQUEST_GLOBAL_RESET          = 0x53 // ICC requested a global reset after DID timeout
	BUP_STATUS_MTP_CLOCK_FREQ_CHECK              = 0x54 // Perform ROSC clock frequency check
	BUP_STATUS_MTP_CLOCK_FREQ_CHECK_FAIL         = 0x55 // Clock check failed
	BUP_STATUS_CPURST_DEASSERT_FAIL              = 0x56 // CPU_RESET_DONE_ACK not received Reserved
	BUP_STATUS_ULV_CHECK_START                   = 0x58 // ULV PCH check-in progress
	BUP_STATUS_ULV_CHECK_FAIL                    = 0x59 // ULV PCH not paired with LV/ULV CPU
	BUP_STATUS_VDM_HW_FAIL                       = 0x5a // VDM handshake failure
	BUP_STATUS_FFS_ENTRY                         = 0x5d // FFS entry
	BUP_STATUS_FFS_EXIT                          = 0x5e // FFS exit
	BUP_STATUS_DRNG_ERROR                        = 0x60 // Error when getting a random number from DRNG
	BUP_STATUS_GET_SP_CANARY                     = 0x61 // Find stack protection canary from NVAR or TRNG
	BUP_STATUS_VDM_GET_SID                       = 0x62 // VDM Get SID Message in progress
	BUP_STATUS_VLB_WAIT                          = 0x63 // Wait for CPU to issue VLB authentication response
	BUP_STATUS_GPDMA_MEMORY_ACCESS_CMRST         = 0x64 // ME Reset due to invalid access of host memory by NP
	BUP_STATUS_PCH_MISMATCH                      = 0x65 // PCH HW type Mismatch versus what was emulated using FITC
	BUP_STATUS_MBP_WRITE                         = 0x66 // MBP written to HECI buffer
	BUP_STATUS_FFS_EXIT_ERROR                    = 0x67 // FW expects FFS exit but BIOS did not set bit in DID
	BUP_STATUS_DEEP_S3_EXIT                      = 0x6c // Exiting DeepS3
	BUP_STATUS_DEEP_S3_EXIT_ERROR                = 0x6d // Error exiting DeepS3
	BUP_STATUS_DRNG_BIST_INCOMPLETE_CMRST        = 0x6e
	BUP_STATUS_DRNG_BIST_ES_KAT_FAILURE_CMRST    = 0x6f
	BUP_STATUS_DRNG_BIST_INTEGRITY_FAILURE_CMRST = 0x70
	BUP_STATUS_DRNG_FUSE_ERROR_CMRST             = 0x71
	BUP_STATUS_DRNG_TIMEOUT_CMRST                = 0x72
	BUP_STATUS_DID_ACK_REQ_GRST                  = 0x73
	BUP_STATUS_DID_ACK_REQ_PCR                   = 0x74
	BUP_STATUS_DID_ACK_REQ_NPCR                  = 0x75
	BUP_STATUS_SAFE_MODE_ENTRY                   = 0x76
	BUP_STATUS_RECOVERY_ENTRY_FTP_BAD            = 0x77
	BUP_STATUS_RECOVERY_ENTRY_NFTP_BAD           = 0x78
	BUP_STATUS_START_FIRST_BOOT_ME_HASH          = 0x79
	BUP_STATUS_END_OF_FIRST_BOOT_ME_HASH         = 0x7a
	BUP_STATUS_PCH_ID_MISMATCH                   = 0x7b // PCH ID (UMCHID) Mismatch for production platform only. Saved in flash UMCHID does not match UMCHID read from ROM BIST data
	BUP_STATUS_GRST_AT_REQUEST                   = 0x7c
	BUP_STATUS_GRST_AT_TIMER_EXPIRE              = 0x7d
	BUP_STATUS_CLINK_FATAL_ERROR_CMRST           = 0x7e
	BUP_STATUS_SECURE_BOOT_FAILURE               = 0x7f
	BUP_STATUS_EXCEPTION_RST_PBO                 = 0x80
	BUP_STATUS_SECURE_BOOT_SM_FAILURE            = 0x81
	BUP_STATUS_SECURE_BOOT_EN_PCH_PLUGGED_IN     = 0x82
	BUP_STATUS_MEBX_INVOCATION_REQUESTED         = 0x83
	BUP_STATUS_CM0_MKHI_HANDLER_START            = 0x84
	BUP_STATUS_CM0_MKHI_HANDLER_STOP             = 0x85
	BUP_STATUS_CM0_MBP_WRITE_SUCCESS             = 0x86
	BUP_STATUS_CM0_MBP_WRITE_ERROR               = 0x87
	BUP_STATUS_HOST_BOOT_PREP_FAIL               = 0x88
	BUP_STATUS_HECI_LINK_RESET_START             = 0x89
	BUP_STATUS_HECI_LINK_RESET_DONE              = 0x8a
	BUP_STATUS_UNSUPPORTED_PROD_PCH              = 0x8b
	BUP_STATUS_UNSUPPORTED_SERVER_PCH            = 0x8c // Unsupported PCH running on server chipset/SoCs
	BUP_STATUS_XEON_CPU_AND_NON_SERVER_WS_PCH    = 0x8d
	BUP_STATUS_LP_AND_H_PCH_MISMATCH             = 0x8e
	BUP_STATUS_NON_PV_FW_ON_REVENUE_HW_FAIL      = 0x8f
	BUP_STATUS_DT_H_CPU_AND_MB_PCH               = 0x90 // Desktop H Processor Pairing with Mobile PCH
	BUP_STATUS_MBL_H_CPU_AND_DT_PCH              = 0x91 // Mobile H Processor Pairing with DT and Server PCH
	BUP_STATUS_PLATRST_DEASSERTION_FAILURE       = 0x92
	BUP_STATUS_INVALID_PCH_CPU_COMBINATION       = 0x94 // Illegal CPU/PCH combination is detected

	BrinupPhase            = 0x3
	HostCommunicationPhase = 0x6
	FirmwareUpdatePhase    = 0x7

	IgnitionSKU   = 0
	TXESKU        = 1
	MEConsumerSKU = 2
	MEBusinessSKU = 3
	LightSKU      = 5
	SPSSKU        = 6

	OkRPMCStatus                   = 0x0
	CSEKeysUnavailableRPMCStatus   = 0x1
	CryptoFailureRPMCStatus        = 0x2
	FlashHardwareFailureRPMCStatus = 0x3
	PCHHardwareFailureRPMCStatus   = 0x4

	ResumeS4S5G3      = 0
	ResumeS3S3DeepRIT = 1

	BootGuardACMSource = 0
	MicrocodeSource    = 1

	InitializationFailedBootGuardError = 0x01
	KMVerificationFailedBootGuardError = 0x02
	BPMFailedBootGuardError            = 0x03
	IBBFailedBootGuardError            = 0x04
	FITFailedBootGuardError            = 0x05
	DMAFailedBootGuardError            = 0x06
	NEMFailedBootGuardError            = 0x07
	TPMFailedBootGuardError            = 0x08
	IBBMeasurementFailedBootGuardError = 0x09
	MEConnectionFailedBootGuardError   = 0x0a

	NoStartEnforcement  = 0
	PCHStartEnforcement = 1

	NoShutdownENF    = 0
	ShutdownENF      = 1
	ShutdownNowENF   = 2
	Shutdown30MinENF = 3
)

type StatusRegisters struct {
	FWSTS1 uint32

	WorkingState             int
	ManufacturingMode        bool
	BadChecksum              bool
	OperationState           int
	InitComplete             bool
	BringupLoadFailed        bool
	FirmwareUpdateInProgress bool
	ErrorCode                int
	OperationMode            int
	ResetCount               int
	BootOptionsPresent       bool
	InvokeEnhancedDebugMode  bool
	BISTTestState            bool
	BISTResetRequest         bool
	PowerSource              int
	D0i3Support              bool

	FWSTS2 uint32

	FlashPartitionFailure bool
	ICCProgrammingStatus  int
	InvokeIMBEX           bool
	CPUReplaced           bool
	FileSystemCorruption  bool
	WarmResetRequested    bool
	CPUReplacedValid      bool
	LowPower              bool
	PowerGating           bool
	IUPNeeded             bool
	ForcedSafeBoot        bool
	ListenerChanged       bool
	BringupState          int // BUP_STATUS_*
	PMEvent               int
	Phase                 int

	FWSTS3 uint32

	FirmwareSKU int
	RPMCStatus  int

	FWSTS4 uint32

	BootGuardEnforcement     bool
	ResumeType               int
	TPMDisconnected          bool
	BootGuardBitsValid       bool
	BootGuardSelfTestFailure bool

	FWSTS5 uint32

	BootGuardACMActive        bool
	BootGuardACMBitsValid     bool
	BootGuardResultCodeSource int
	BootGuardErrorCode        int
	BootGuardACMDone          bool
	StartupModuleTimeoutCount int
	EnableSCRTMIndicator      bool
	IncrementBootGuardACMSVN  int
	IncrementKMACMSVN         int
	IncrementBPMACMSVN        int
	StartEnforcement          int

	FWSTS6 uint32

	ForceBootGuard            bool
	CPUDebugDisabled          bool
	BSPInitializationDisabled bool
	ProtectBIOSEnvironment    bool
	ErrorEnforcementPolicy    int
	MeasurmentPolicy          bool
	VerifiedBootPolicy        bool
	BootGuardACMSVN           int
	KMACMSVN                  int
	BPMACMSVN                 int
	KMID                      int
	BSTExecutedBPM            bool
	EnforcementError          int
	BootGuardDisabled         bool
	FPFDisable                bool
	FPFLocked                 bool
	TXTSupport                bool
}

const (
	fwsts1Offset = 0x40
	fwsts2Offset = 0x48
	fwsts3Offset = 0x60
	fwsts4Offset = 0x64
	fwsts5Offset = 0x68
	fwsts6Offset = 0x6C
)

func ParseStatusRegisters(configSpace []byte) (*StatusRegisters, error) {
	var regs StatusRegisters
	rd := bytes.NewReader(configSpace)

	readReg := func(offset int64, reg *uint32) error {
		if _, err := rd.Seek(offset, io.SeekStart); err != nil {
			return err
		}
		return binary.Read(rd, binary.LittleEndian, reg)
	}

	if err := readReg(fwsts1Offset, &regs.FWSTS1); err != nil {
		return nil, err
	}
	// 3:0
	regs.WorkingState = int(regs.FWSTS1) & 0b1111
	// 4
	regs.ManufacturingMode = regs.FWSTS1&(1<<4) != 0
	// 5
	regs.BadChecksum = regs.FWSTS1&(1<<5) != 0
	// 8:6
	regs.OperationState = (int(regs.FWSTS1) >> 6) & 0b111
	// 9
	regs.InitComplete = regs.FWSTS1&(1<<9) != 0
	// 10
	regs.BringupLoadFailed = regs.FWSTS1&(1<<10) != 0
	// 11
	regs.FirmwareUpdateInProgress = regs.FWSTS1&(1<<11) != 0
	// 15:12
	regs.ErrorCode = (int(regs.FWSTS1) >> 12) & 0b1111
	// 19:16
	regs.OperationMode = (int(regs.FWSTS1) >> 16) & 0b1111
	// 23:20
	regs.ResetCount = (int(regs.FWSTS1) >> 20) & 0b1111
	// 24
	regs.BootOptionsPresent = regs.FWSTS1&(1<<20) != 0
	// 25
	regs.InvokeEnhancedDebugMode = regs.FWSTS1&(1<<25) != 0
	// 26
	regs.BISTTestState = regs.FWSTS1&(1<<26) != 0
	// 27
	regs.BISTResetRequest = regs.FWSTS1&(1<<27) != 0
	// 29:28
	regs.PowerSource = (int(regs.FWSTS1) >> 28) & 0b11
	// 31
	regs.D0i3Support = regs.FWSTS1&(1<<31) != 0

	if err := readReg(fwsts2Offset, &regs.FWSTS2); err != nil {
		return nil, err
	}
	// 0
	regs.FlashPartitionFailure = regs.FWSTS2&1 != 0
	// 2:1
	regs.ICCProgrammingStatus = (int(regs.FWSTS2) >> 1) & 0b11
	// 3
	regs.InvokeIMBEX = regs.FWSTS2&(1<<3) != 0
	// 4
	regs.CPUReplaced = regs.FWSTS2&(1<<4) != 0
	// 6
	regs.FileSystemCorruption = regs.FWSTS2&(1<<6) != 0
	// 7
	regs.WarmResetRequested = regs.FWSTS2&(1<<7) != 0
	// 8
	regs.CPUReplacedValid = regs.FWSTS2&(1<<8) != 0
	// 9
	regs.LowPower = regs.FWSTS2&(1<<9) != 0
	// 10
	regs.PowerGating = regs.FWSTS2&(1<<10) != 0
	// 11
	regs.IUPNeeded = regs.FWSTS2&(1<<11) != 0
	// 12
	regs.ForcedSafeBoot = regs.FWSTS2&(1<<12) != 0
	// 15
	regs.ListenerChanged = regs.FWSTS2&(1<<15) != 0
	// 23:16
	regs.BringupState = (int(regs.FWSTS2) >> 16) & 0b11111111
	// 27:24
	regs.PMEvent = (int(regs.FWSTS2) >> 24) & 0b1111
	// 31:28
	regs.Phase = (int(regs.FWSTS2) >> 28) & 0b1111

	if err := readReg(fwsts3Offset, &regs.FWSTS3); err != nil {
		return nil, err
	}
	// 6:4
	regs.FirmwareSKU = (int(regs.FWSTS3) >> 4) & 0b111
	// 13:11
	regs.RPMCStatus = (int(regs.FWSTS3) >> 11) & 0b111

	if err := readReg(fwsts4Offset, &regs.FWSTS4); err != nil {
		return nil, err
	}
	// 9
	regs.BootGuardEnforcement = regs.FWSTS4&(1<<0) != 0
	// 10
	regs.ResumeType = (int(regs.FWSTS4) >> 10) & 0b1
	// 12
	regs.TPMDisconnected = regs.FWSTS4&(1<<0) != 0
	// 14
	regs.BootGuardBitsValid = regs.FWSTS4&(1<<0) != 0
	// 15
	regs.BootGuardSelfTestFailure = regs.FWSTS4&(1<<0) != 0

	if err := readReg(fwsts5Offset, &regs.FWSTS5); err != nil {
		return nil, err
	}
	// 0
	regs.BootGuardACMActive = regs.FWSTS5&(1<<0) != 0
	// 1
	regs.BootGuardACMBitsValid = regs.FWSTS5&(1<<1) != 0
	// 2
	regs.BootGuardResultCodeSource = (int(regs.FWSTS5) >> 2) & 0b1
	// 7:3
	regs.BootGuardErrorCode = (int(regs.FWSTS5) >> 3) & 0b11111
	// 8
	regs.BootGuardACMDone = regs.FWSTS5&(1<<8) != 0
	// 15:9
	regs.StartupModuleTimeoutCount = (int(regs.FWSTS5) >> 9) & 0b1111111
	// 16
	regs.EnableSCRTMIndicator = regs.FWSTS5&(1<<16) != 0
	// 20:17
	regs.IncrementBootGuardACMSVN = (int(regs.FWSTS5) >> 17) & 0b1111
	// 24:21
	regs.IncrementKMACMSVN = (int(regs.FWSTS5) >> 21) & 0b1111
	// 28:25
	regs.IncrementBPMACMSVN = (int(regs.FWSTS5) >> 25) & 0b1111
	// 31
	regs.StartEnforcement = (int(regs.FWSTS5) >> 31) & 0b1

	if err := readReg(fwsts6Offset, &regs.FWSTS6); err != nil {
		return nil, err
	}
	// 0
	regs.ForceBootGuard = regs.FWSTS6&(1<<0) != 0
	// 1
	regs.CPUDebugDisabled = regs.FWSTS6&(1<<0) != 0
	// 2
	regs.BSPInitializationDisabled = regs.FWSTS6&(1<<0) != 0
	// 3
	regs.ProtectBIOSEnvironment = regs.FWSTS6&(1<<0) != 0
	// 7:6
	regs.ErrorEnforcementPolicy = (int(regs.FWSTS6) >> 6) & 0b11
	// 8
	regs.MeasurmentPolicy = regs.FWSTS6&(1<<0) != 0
	// 9
	regs.VerifiedBootPolicy = regs.FWSTS6&(1<<0) != 0
	// 13:10
	regs.BootGuardACMSVN = (int(regs.FWSTS6) >> 10) & 0b1111
	// 17:14
	regs.KMACMSVN = (int(regs.FWSTS6) >> 14) & 0b1111
	// 21:18
	regs.BPMACMSVN = (int(regs.FWSTS6) >> 18) & 0b1111
	// 25:22
	regs.KMID = (int(regs.FWSTS6) >> 22) & 0b1111
	// 26
	regs.BSTExecutedBPM = regs.FWSTS6&(1<<0) != 0
	// 27
	regs.EnforcementError = (int(regs.FWSTS6) >> 27) & 0b1
	// 28
	regs.BootGuardDisabled = regs.FWSTS6&(1<<0) != 0
	// 29
	regs.FPFDisable = regs.FWSTS6&(1<<0) != 0
	// 30
	regs.FPFLocked = regs.FWSTS6&(1<<0) != 0
	// 31
	regs.TXTSupport = regs.FWSTS6&(1<<0) != 0

	return &regs, nil
}

// Result of FWCAPS FeatureState
// https://github.com/intel/dynamic-application-loader-host-interface/blob/cbf9e015dd9de03eb5df1124d568de50ec7784d5/common/FWUpdate/FwCapsMsgs.h
func Features(bits uint32) []string {
	var features []string
	if (bits>>0)&1 != 0 {
		features = append(features, "Full network manageability")
	}
	if (bits>>1)&1 != 0 {
		features = append(features, "Standard network manageability")
	}
	if (bits>>2)&1 != 0 {
		features = append(features, "Manageability")
	}
	if (bits>>4)&1 != 0 {
		features = append(features, "Intel Integrated Touch")
	}
	if (bits>>5)&1 != 0 {
		features = append(features, "Anti-Theft Technology")
	}
	if (bits>>6)&1 != 0 {
		features = append(features, "Capability Licensing Service")
	}
	if (bits>>7)&1 != 0 {
		features = append(features, "Virtualization Engine")
	}
	if (bits>>10)&1 != 0 {
		features = append(features, "Intel Sensor Hub")
	}
	if (bits>>11)&1 != 0 {
		features = append(features, "ICC")
	}
	if (bits>>12)&1 != 0 {
		features = append(features, "Protected Audio Video Path")
	}
	if (bits>>16)&1 != 0 {
		features = append(features, "High Assurance Platform")
	}
	if (bits>>17)&1 != 0 {
		features = append(features, "IPv6")
	}
	if (bits>>18)&1 != 0 {
		features = append(features, "KVM Remote Control")
	}
	if (bits>>20)&1 != 0 {
		features = append(features, "Dynamic Application Loader")
	}
	if (bits>>21)&1 != 0 {
		features = append(features, "Cipher Transport Layer")
	}
	if (bits>>23)&1 != 0 {
		features = append(features, "Wireless Lan")
	}
	if (bits>>24)&1 != 0 {
		features = append(features, "Wireless Display")
	}
	if (bits>>25)&1 != 0 {
		features = append(features, "USB 3.0")
	}
	if (bits>>29)&1 != 0 {
		features = append(features, "Platform Trust Technology")
	}
	if (bits>>31)&1 != 0 {
		features = append(features, "NFC")
	}
	return features
}

func spsFeatures(fwState *SPSGetMEBIOSResponse) []string {
	var features []string
	if (fwState.Features1>>0)&1 != 0 {
		features = append(features, "Intel Node Manager")
	}
	if (fwState.Features1>>1)&1 != 0 {
		features = append(features, "PECI Proxy")
	}
	if (fwState.Features1>>2)&1 != 0 {
		features = append(features, "ICC")
	}
	if (fwState.Features1>>3)&1 != 0 {
		features = append(features, "Intel ME Storage Services")
	}
	if (fwState.Features1>>4)&1 != 0 {
		features = append(features, "Intel Boot Guard")
	}
	if (fwState.Features1>>5)&1 != 0 {
		features = append(features, "Intel Platform Trust Technology (PTT)")
	}
	if (fwState.Features1>>6)&1 != 0 {
		features = append(features, "OEM Defined CPU Debug Policy")
	}
	if (fwState.Features1>>7)&1 != 0 {
		features = append(features, "Reset Suppression")
	}
	if (fwState.Features1>>8)&1 != 0 {
		features = append(features, "PMBus Proxy over HECI")
	}
	if (fwState.Features1>>9)&1 != 0 {
		features = append(features, "CPU Hot-Plug/Remove")
	}
	if (fwState.Features1>>10)&1 != 0 {
		features = append(features, "MIC/IPMB Proxy")
	}
	if (fwState.Features1>>11)&1 != 0 {
		features = append(features, "MCTP Proxy")
	}
	if (fwState.Features1>>12)&1 != 0 {
		features = append(features, "Thermal Reporting and Volumetric Airflow")
	}
	if (fwState.Features1>>13)&1 != 0 {
		features = append(features, "SoC Thermal Reporting")
	}
	if (fwState.Features1>>14)&1 != 0 {
		features = append(features, "Dual BIOS Support")
	}
	if (fwState.Features1>>15)&1 != 0 {
		features = append(features, "MPHY Survivability Programming")
	}
	if (fwState.Features1>>16)&1 != 0 {
		features = append(features, "Inband PECI")
	}
	if (fwState.Features1>>17)&1 != 0 {
		features = append(features, "PCH Debug (JTAG)")
	}
	if (fwState.Features1>>18)&1 != 0 {
		features = append(features, "Power Thermal Utility Support")
	}
	if (fwState.Features1>>19)&1 != 0 {
		features = append(features, "FIA Mux Configuration")
	}
	if (fwState.Features1>>20)&1 != 0 {
		features = append(features, "PCH Thermal Sensor Init")
	}
	if (fwState.Features1>>21)&1 != 0 {
		features = append(features, "DeepSx Support")
	}
	if (fwState.Features1>>22)&1 != 0 {
		features = append(features, "Dual Intel ME FW Image")
	}
	if (fwState.Features1>>23)&1 != 0 {
		features = append(features, "Direct FW Update")
	}
	if (fwState.Features1>>24)&1 != 0 {
		features = append(features, "MCTP Infrastructure")
	}
	if (fwState.Features1>>25)&1 != 0 {
		features = append(features, "CUPS")
	}
	if (fwState.Features1>>26)&1 != 0 {
		features = append(features, "Flash Descriptor Region Verification")
	}
	if (fwState.Features1>>27)&1 != 0 {
		features = append(features, "Intel Software Guard Extensions (SGX)")
	}
	if (fwState.Features1>>28)&1 != 0 {
		features = append(features, "Turbo State Limiting")
	}
	if (fwState.Features1>>29)&1 != 0 {
		features = append(features, "Telemetry Hub")
	}
	if (fwState.Features1>>30)&1 != 0 {
		features = append(features, "Intel ME Shutdown on EOP")
	}
	if (fwState.Features1>>31)&1 != 0 {
		features = append(features, "ASA")
	}
	if (fwState.Features2>>0)&1 != 0 {
		features = append(features, "Warm Reset Notification")
	}
	return features
}
