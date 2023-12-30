package eventlog

import (
	"crypto"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/x509"
)

type specIDEventHeader struct {
	Signature     [16]byte
	PlatformClass uint32
	VersionMinor  uint8
	VersionMajor  uint8
	Errata        uint8
	UintnSize     uint8
	NumAlgs       uint32
}

// EFIGUID represents the EFI_GUID type.
// See section "2.3.1 Data Types" in the specification for more information.
// type EFIGUID [16]byte
type EFIGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// efiConfigurationTable represents the EFI_CONFIGURATION_TABLE type.
// See section "4.6 EFI Configuration Table & Properties Table" in the specification for more information.
type efiConfigurationTable struct {
	VendorGUID  EFIGUID
	VendorTable uint64 // "A pointer to the table associated with VendorGuid"
}

type UEFIVariableDataHeader struct {
	VariableName       EFIGUID
	UnicodeNameLength  uint64 // uintN
	VariableDataLength uint64 // uintN
}

// UEFIVariableData represents the UEFI_VARIABLE_DATA structure.
type UEFIVariableData struct {
	Header       UEFIVariableDataHeader
	UnicodeName  []uint16
	VariableData []byte // []int8
}

// efiTableHeader represents the EFI_TABLE_HEADER type.
// See section "4.2 EFI Table Header" in the specification for more information.
type efiTableHeader struct {
	Signature  uint64
	Revision   uint32
	HeaderSize uint32
	CRC32      uint32
	Reserved   uint32
}

// efiPartitionTableHeader represents the structure described by "Table 20. GPT Header."
// See section "5.3.2 GPT Header" in the specification for more information.
type EFIPartitionTableHeader struct {
	Header                   efiTableHeader
	MyLBA                    efiLBA
	AlternateLBA             efiLBA
	FirstUsableLBA           efiLBA
	LastUsableLBA            efiLBA
	DiskGUID                 EFIGUID
	PartitionEntryLBA        efiLBA
	NumberOfPartitionEntries uint32
	SizeOfPartitionEntry     uint32
	PartitionEntryArrayCRC32 uint32
}

// efiPartition represents the structure described by "Table 21. GPT Partition Entry."
// See section "5.3.3 GPT Partition Entry" in the specification for more information.
type EFIPartition struct {
	TypeGUID       EFIGUID
	PartitionGUID  EFIGUID
	FirstLBA       efiLBA
	LastLBA        efiLBA
	AttributeFlags uint64
	PartitionName  [36]uint16
}

// EFISignatureData represents the EFI_SIGNATURE_DATA type.
// See section "31.4.1 Signature Database" in the specification
// for more information.
type EFISignatureData struct {
	SignatureOwner EFIGUID
	SignatureData  []byte // []int8
}

// efiSignatureData represents the EFI_SIGNATURE_DATA type.
// See section "31.4.1 Signature Database" in the specification for more information.
type efiSignatureData struct {
	SignatureOwner EFIGUID
	SignatureData  []byte // []int8
}

// efiSignatureList represents the EFI_SIGNATURE_LIST type.
// See section "31.4.1 Signature Database" in the specification for more information.
type efiSignatureListHeader struct {
	SignatureType       EFIGUID
	SignatureListSize   uint32
	SignatureHeaderSize uint32
	SignatureSize       uint32
}

type efiSignatureList struct {
	Header        efiSignatureListHeader
	SignatureData []byte
	Signatures    []byte
}

// EFIVariableAuthority describes the contents of a UEFI variable authority
// event.
type efiVariableAuthority struct {
	Certs []x509.Certificate
}

// TaggedEventData represents the TCG_PCClientTaggedEventStruct structure,
// as defined by 11.3.2.1 in the "TCG PC Client Specific Implementation
// Specification for Conventional BIOS", version 1.21.
type TaggedEventData struct {
	ID   uint32
	Data []byte
}

type TPMEvent interface {
	RawEvent() Event
}

type baseEvent struct {
	Event
	Err error
}

type stringEvent struct {
	baseEvent
	Message string
}

type PrebootCertEvent struct {
	baseEvent
}

type PostEvent struct {
	stringEvent
}

type NoActionEvent struct {
	baseEvent
}

type SeparatorEvent struct {
	baseEvent
}

type ActionEvent struct {
	stringEvent
}

type EventTagEvent struct {
	baseEvent
	EventID   eventID
	EventData []byte
}

type CRTMContentEvent struct {
	stringEvent
}

type CRTMEvent struct {
	stringEvent
}

type MicrocodeEvent struct {
	stringEvent
}

type PlatformConfigFlagsEvent struct {
	baseEvent
}

type TableOfDevicesEvent struct {
	baseEvent
}

type CompactHashEvent struct {
	baseEvent
}

type IPLEvent struct {
	stringEvent
}

type IPLPartitionEvent struct {
	baseEvent
}

type NonHostCodeEvent struct {
	baseEvent
}

type NonHostConfigEvent struct {
	baseEvent
}

type NonHostInfoEvent struct {
	baseEvent
}

type OmitBootDeviceEventsEvent struct {
	stringEvent
}

type UEFIVariableEvent struct {
	baseEvent
	VariableGUID EFIGUID
	VariableName string
	VariableData []byte
}

type UEFIVariableDriverConfigEvent struct {
	UEFIVariableEvent
}

type UEFIBootVariableEvent struct {
	UEFIVariableEvent
	Description   string
	DevicePath    string
	DevicePathRaw []byte
	OptionalData  []byte
}

type UEFIVariableAuthorityEvent struct {
	UEFIVariableEvent
}

type UEFIImageLoadEvent struct {
	baseEvent
	ImageLocationInMemory uint64
	ImageLengthInMemory   uint64
	ImageLinkTimeAddress  uint64
	DevicePath            string
	DevicePathRaw         []byte
}

type UEFIBootServicesApplicationEvent struct {
	UEFIImageLoadEvent
}

type UEFIBootServicesDriverEvent struct {
	UEFIImageLoadEvent
}

type UEFIRuntimeServicesDriverEvent struct {
	UEFIImageLoadEvent
}

type UEFIActionEvent struct {
	stringEvent
}

type UEFIGPTEvent struct {
	baseEvent
	UEFIPartitionHeader EFIPartitionTableHeader
	Partitions          []EFIPartition
}

type UEFIPlatformFirmwareBlobEvent struct {
	baseEvent
	BlobBase   uint64
	BlobLength uint64
}

type UEFIHandoffTableEvent struct {
	baseEvent
	Tables []efiConfigurationTable
}

type OptionROMConfigEvent struct {
	baseEvent
	PFA             uint16
	OptionROMStruct []byte
}

type MicrosoftBootEvent struct {
	baseEvent
	Events []MicrosoftEvent
}

type microsoftEventHeader struct {
	Type windowsEvent
	Size uint32
}

type microsoftBootRevocationList struct {
	CreationTime  uint64
	DigestLength  uint32
	HashAlgorithm uint16
	Digest        []byte
}

type windowsEvent uint32

// MicrosoftEvent ...
type MicrosoftEvent interface {
}

type microsoftBaseEvent struct {
	Type windowsEvent
}

// MicrosoftStringEvent ...
type MicrosoftStringEvent struct {
	microsoftBaseEvent
	Message string
}

// MicrosoftRevocationEvent ...
type MicrosoftRevocationEvent struct {
	microsoftBaseEvent
	CreationTime  uint64
	DigestLength  uint32
	HashAlgorithm uint16
	Digest        []byte
}

// MicrosoftDataEvent ...
type MicrosoftDataEvent struct {
	microsoftBaseEvent
	Data []byte
}

// ReplayError describes the parsed events that failed to verify against
// a particular PCR.
type ReplayError struct {
	Events []Event
	// InvalidPCRs reports the set of PCRs where the event log replay failed.
	InvalidPCRs []int
}

// Event is a single event from a TCG event log. This reports descrete items such
// as BIOs measurements or EFI states.
//
// There are many pitfalls for using event log events correctly to determine the
// state of a machine[1]. In general it's must safer to only rely on the raw PCR
// values and use the event log for debugging.
//
// [1] https://github.com/google/go-attestation/blob/master/docs/event-log-disclosure.md
type Event struct {
	// order of the event in the event log.
	Sequence int
	// Index of the PCR that this event was replayed against.
	Index int
	// Untrusted type of the event. This value is not verified by event log replays
	// and can be tampered with. It should NOT be used without additional context,
	// and unrecognized event types should result in errors.
	Type EventType

	// Data of the event. For certain kinds of events, this must match the event
	// digest to be valid.
	Data []byte
	// Digest is the verified digest of the event data. While an event can have
	// multiple for different hash values, this is the one that was matched to the
	// PCR value.
	Digest []byte
	Alg    HashAlg
}

// EventLog is a parsed measurement log. This contains unverified data representing
// boot events that must be replayed against PCR values to determine authenticity.
type EventLog struct {
	// Algs holds the set of algorithms that the event log uses.
	Algs      []HashAlg
	rawEvents []rawEvent
}

type rawAttestationData struct {
	Version [4]byte  // This MUST be 1.1.0.0
	Fixed   [4]byte  // This SHALL always be the string ‘QUOT’
	Digest  [20]byte // PCR Composite Hash
	Nonce   [20]byte // Nonce Hash
}

type rawPCRComposite struct {
	Size    uint16 // always 3
	PCRMask [3]byte
	Values  tpmutil.U32Bytes
}

type pcrReplayResult struct {
	events     []Event
	successful bool
}

type specIDEvent struct {
	algs []specAlgSize
}

type specAlgSize struct {
	ID   uint16
	Size uint16
}

type digest struct {
	hash crypto.Hash
	data []byte
}

type rawEvent struct {
	sequence int
	index    int
	typ      EventType
	data     []byte
	digests  []digest
}

// TPM 1.2 event log format. See "5.1 SHA1 Event Log Entry Format"
// https://trustedcomputinggroup.org/wp-content/uploads/EFI-Protocol-Specification-rev13-160330final.pdf#page=15
type rawEventHeader struct {
	PCRIndex  uint32
	Type      uint32
	Digest    [20]byte
	EventSize uint32
}

type eventSizeErr struct {
	eventSize uint32
	logSize   int
}

// TPM 2.0 event log format. See "5.2 Crypto Agile Log Entry Format"
// https://trustedcomputinggroup.org/wp-content/uploads/EFI-Protocol-Specification-rev13-160330final.pdf#page=15
type rawEvent2Header struct {
	PCRIndex uint32
	Type     uint32
}

type elWorkaround struct {
	id          string
	affectedPCR int
	apply       func(e *EventLog) error
}

// efiDevicePathHeader represents the EFI_DEVICE_PATH_PROTOCOL type.
// See section "10.2 EFI Device Path Protocol" in the specification for more information.
type efiDevicePathHeader struct {
	Type    efiDevicePathType
	SubType uint8
	Length  uint16
}

// The canonical representation of EFI Device Paths to text is Table
// 102 in section 10.6.1.6 of the spec. The reference implementation is
// https://github.com/tianocore/edk2/blob/master/MdePkg/Library/UefiDevicePathLib/DevicePathFromText.c
type efiPCIDevicePath struct {
	Function uint8
	Device   uint8
}

type efiMMIODevicePath struct {
	MemoryType   uint32
	StartAddress uint64
	EndAddress   uint64
}

// efiLBA represents the EFI_LBA type.
// See section "2.3.1 Data Types" in the specification for more information.
type efiLBA uint64

type efiHardDriveDevicePath struct {
	Partition          uint32
	PartitionStart     efiLBA
	PartitionSize      efiLBA
	PartitionSignature [16]byte
	PartitionFormat    uint8
	SignatureType      efiSignatureType
}

type efiMacMessagingDevicePath struct {
	MAC    [32]byte
	IfType byte
}

type efiIpv4MessagingDevicePath struct {
	LocalAddress   [4]byte
	RemoteAddress  [4]byte
	LocalPort      uint16
	RemotePort     uint16
	Protocol       uint16
	StaticIP       byte
	GatewayAddress [4]byte
	SubnetMask     [4]byte
}

type efiIpv6MessagingDevicePath struct {
	LocalAddress  [16]byte
	RemoteAddress [16]byte
	LocalPort     uint16
	RemotePort    uint16
	Protocol      uint16
	AddressOrigin byte
	PrefixLength  byte
	GatewayIP     [16]byte
}

type efiUsbMessagingDevicePath struct {
	ParentPort byte
	Interface  byte
}

type efiVendorMessagingDevicePath struct {
	GUID EFIGUID
	Data []byte
}

type efiSataMessagingDevicePath struct {
	HBA            uint16
	PortMultiplier uint16
	LUN            uint16
}

type efiNvmMessagingDevicePath struct {
	Namespace uint32
	EUI       [8]byte
}

type efiACPIDevicePath struct {
	HID uint32
	UID uint32
}

type efiExpandedACPIDevicePathFixed struct {
	HID uint32
	UID uint32
	CID uint32
}

type efiExpandedACPIDevicePath struct {
	Fixed  efiExpandedACPIDevicePathFixed
	HIDStr string
	UIDStr string
	CIDStr string
}

type efiAdrACPIDevicePath struct {
	ADRs []uint32
}

type efiPiwgFileDevicePath struct {
	GUID EFIGUID
}

type efiPiwgVolumeDevicePath struct {
	GUID EFIGUID
}

type efiOffsetDevicePath struct {
	Reserved    uint32
	StartOffset uint64
	EndOffset   uint64
}

type efiBBSDevicePathFixed struct {
	DeviceType uint16
	Status     uint16
}

type efiBBSDevicePath struct {
	Fixed       efiBBSDevicePathFixed
	Description []byte
}

// Dump describes the layout of serialized information from the dump command.
type Dump struct {
	Log struct {
		PCRs   []PCR
		PCRAlg tpm2.Algorithm
		Raw    []byte // The measured boot log in binary form.
	}
}

type WinCSPAlg uint32

// Valid CSP Algorithm IDs.
const (
	WinAlgMD4    WinCSPAlg = 0x02
	WinAlgMD5    WinCSPAlg = 0x03
	WinAlgSHA1   WinCSPAlg = 0x04
	WinAlgSHA256 WinCSPAlg = 0x0c
	WinAlgSHA384 WinCSPAlg = 0x0d
	WinAlgSHA512 WinCSPAlg = 0x0e
)

// BitlockerStatus describes the status of BitLocker on a Windows system.
type BitlockerStatus uint8

// Valid BitlockerStatus values.
const (
	BitlockerStatusCached   = 0x01
	BitlockerStatusMedia    = 0x02
	BitlockerStatusTPM      = 0x04
	BitlockerStatusPin      = 0x10
	BitlockerStatusExternal = 0x20
	BitlockerStatusRecovery = 0x40
)

// Ternary describes a boolean value that can additionally be unknown.
type Ternary uint8

// Valid Ternary values.
const (
	TernaryUnknown Ternary = iota
	TernaryTrue
	TernaryFalse
)

// WinEvents describes information from the event log recorded during
// bootup of Microsoft Windows.
type WinEvents struct {
	// ColdBoot is set to true if the system was not resuming from hibernation.
	ColdBoot bool
	// BootCount contains the value of the monotonic boot counter. This
	// value is not set for TPM 1.2 devices and some TPMs with buggy
	// implementations of monotonic counters.
	BootCount uint64
	// EventCount contains the value of the monotonic event counter. It is
	// incremented every time bootmgr is run; this is also the case on resume
	// from hibernate. BootCount in contrast is only incremented on cold boots.
	EventCount uint64
	// ID of nvram counter
	EventCounterId uint64
	// LoadedModules contains authenticode hashes for binaries which
	// were loaded during boot.
	LoadedModules map[string]WinModuleLoad
	// ELAM describes the configuration of each Early Launch AntiMalware driver,
	// for each AV Vendor key.
	ELAM map[string]WinELAM
	// TrustPointQuote contains quotes over WBCLs for each AIK id
	TrustPointQuote map[string]WinWBCLQuote
	// BootDebuggingEnabled is true if boot debugging was ever reported
	// as enabled.
	BootDebuggingEnabled bool
	// KernelDebugEnabled is true if kernel debugging was recorded as
	// enabled at any point during boot.
	KernelDebugEnabled bool
	// DEPEnabled is true if NX (Data Execution Prevention) was consistently
	// reported as enabled.
	DEPEnabled Ternary
	// CodeIntegrityEnabled is true if code integrity was consistently
	// reported as enabled.
	CodeIntegrityEnabled Ternary
	// TestSigningEnabled is true if test-mode signature verification was
	// ever reported as enabled.
	TestSigningEnabled bool
	// BitlockerUnlocks reports the bitlocker status for every instance of
	// a disk unlock, where bitlocker was used to secure the disk.
	BitlockerUnlocks []BitlockerStatus
}

// WinModuleLoad describes a module which was loaded while
// Windows booted.
type WinModuleLoad struct {
	// FilePath represents the path from which the module was loaded. This
	// information is not always present.
	FilePath string
	// AuthenticodeHash contains the authenticode hash of the binary
	// blob which was loaded.
	AuthenticodeHash []byte
	// ImageBase describes all the addresses to which the the blob was loaded.
	ImageBase []uint64
	// ImageSize describes the size of the image in bytes. This information
	// is not always present.
	ImageSize uint64
	// HashAlgorithm describes the hash algorithm used.
	HashAlgorithm WinCSPAlg
	// ImageValidated is set if the post-boot loader validated the image.
	ImageValidated bool

	// AuthorityIssuer identifies the issuer of the certificate which certifies
	// the signature on this module.
	AuthorityIssuer string
	// AuthorityPublisher identifies the publisher of the certificate which
	// certifies the signature on this module.
	AuthorityPublisher string
	// AuthoritySerial contains the serial of the certificate certifying this
	// module.
	AuthoritySerial []byte
	// AuthoritySHA1 is the SHA1 hash of the certificate thumbprint.
	AuthoritySHA1 []byte
}

// WinELAM describes the configuration of an Early Launch AntiMalware driver.
// These values represent the 3 measured registry values stored in the ELAM
// hive for the driver.
type WinELAM struct {
	Measured []byte
	Config   []byte
	Policy   []byte
}

// WinWBCLQuote contains the data found within a TrusPoint aggregation quote at the end of WBCLs
type WinWBCLQuote struct {
	AIKPubDigest   []byte
	Quote          []byte
	QuoteSignature []byte
}

type HumanEventLog struct {
	Algorithms   []string `json:"Algorithms"`
	Verified     bool     `json:"Verifiable"`
	SHA1Events   []string `json:"SHA-1 Events"`
	SHA256Events []string `json:"SHA-256 Events"`
}
