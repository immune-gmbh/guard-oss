// JSON:API structures sent and received over the wire.
//
// Keep in sync with agent/pkg/api/types.go
package api

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

type FirmwareError string

const (
	NoError        FirmwareError = ""
	UnknownError   FirmwareError = "unkn"
	NoPermission   FirmwareError = "no-perm"
	NoResponse     FirmwareError = "no-resp"
	NotImplemented FirmwareError = "not-impl"
)

// /v2/info (apisrv)
type Info struct {
	APIVersion string `jsonapi:"attr,api_version" json:"api_version"`
}

// /v2/configuration (apisrv)
type KeyTemplate struct {
	Public PublicKey `json:"public"`
	Label  string    `json:"label"`
}

// /v2/configuration (apisrv)
type Configuration struct {
	Root            KeyTemplate            `jsonapi:"attr,root" json:"root"`
	Keys            map[string]KeyTemplate `jsonapi:"attr,keys" json:"keys"`
	PCRBank         uint16                 `jsonapi:"attr,pcr_bank" json:"pcr_bank"`
	PCRs            []int                  `jsonapi:"attr,pcrs" json:"pcrs"`
	UEFIVariables   []UEFIVariable         `jsonapi:"attr,uefi" json:"uefi"`
	MSRs            []MSR                  `jsonapi:"attr,msrs" json:"msrs"`
	CPUIDLeafs      []CPUIDLeaf            `jsonapi:"attr,cpuid" json:"cpuid"`
	TPM2NVRAM       []uint32               `jsonapi:"attr,tpm2_nvram" json:"tpm2_nvram"`
	SEV             []SEVCommand           `jsonapi:"attr,sev" json:"sev"`
	ME              []MEClientCommands     `jsonapi:"attr,me" json:"me"`
	TPM2Properties  []TPM2Property         `jsonapi:"attr,tpm2_properties" json:"tpm2_properties"`
	PCIConfigSpaces []PCIConfigSpace       `jsonapi:"attr,pci" json:"pci"`
}

// /v2/attest (apisrv)
type FirmwareProperties struct {
	UEFIVariables   []UEFIVariable     `json:"uefi,omitempty"`
	MSRs            []MSR              `json:"msrs,omitempty"`
	CPUIDLeafs      []CPUIDLeaf        `json:"cpuid,omitempty"`
	SEV             []SEVCommand       `json:"sev,omitempty"`
	ME              []MEClientCommands `json:"me,omitempty"`
	TPM2Properties  []TPM2Property     `json:"tpm2_properties,omitempty"`
	TPM2NVRAM       []TPM2NVIndex      `json:"tpm2_nvram,omitempty"`
	PCIConfigSpaces []PCIConfigSpace   `json:"pci,omitempty"`
	ACPI            ACPITablesV1       `json:"acpi"`
	SMBIOS          HashBlob           `json:"smbios" blobstore:"smbios"`
	TXTPublicSpace  HashBlob           `json:"txt" blobstore:"txt"`
	VTdRegisterSet  HashBlob           `json:"vtd"`
	Flash           HashBlob           `json:"flash" blobstore:"bios"`
	TPM2EventLog    ErrorBuffer        `json:"event_log"`             // deprecated
	TPM2EventLogZ   *ErrorBuffer       `json:"event_log_z,omitempty"` // deprecated
	TPM2EventLogs   []HashBlob         `json:"event_logs,omitempty" blobstore:"eventlog"`
	PCPQuoteKeys    map[string]Buffer  `json:"pcp_quote_keys,omitempty"` // windows only
	MACAddresses    MACAddresses       `json:"mac"`
	OS              OS                 `json:"os"`
	NICs            *NICList           `json:"nic,omitempty"`
	Memory          Memory             `json:"memory"`
	Agent           *Agent             `json:"agent,omitempty"`
	Devices         *Devices           `json:"devices,omitempty"`
	IMALog          *ErrorBuffer       `json:"ima_log,omitempty"`
	EPPInfo         *EPPInfo           `json:"epp_info,omitempty"`
	BootApps        *BootApps          `json:"boot_apps,omitempty"`
}

type BootApps struct {
	Images    map[string]HashBlob `json:"images,omitempty" blobstore:"uefi-app"` // path -> pe file
	ImagesErr FirmwareError       `json:"images_err,omitempty"`
}

type EPPInfo struct {
	AntimalwareProcesses    map[string]HashBlob `json:"antimalware_processes,omitempty" blobstore:"windows-exeutable"` // path -> exe file
	AntimalwareProcessesErr FirmwareError       `json:"antimalware_processes_err,omitempty"`
	EarlyLaunchDrivers      map[string]HashBlob `json:"early_launch_drivers,omitempty" blobstore:"windows-exeutable"` // path -> sys file
	EarlyLaunchDriversErr   FirmwareError       `json:"early_launch_drivers_err,omitempty"`
	ESET                    *ESETConfig         `json:"eset,omitempty"` // Linux only
}

type ESETConfig struct {
	Enabled           ErrorBuffer `json:"enabled"`
	ExcludedFiles     ErrorBuffer `json:"excluded_files"`
	ExcludedProcesses ErrorBuffer `json:"excluded_processes"`
}

type HashBlob struct {
	Sha256 Buffer        `json:"sha256,omitempty"` // hash of uncompressed data
	ZData  Buffer        `json:"z_data,omitempty"` // zstd compressed data, maybe omitted if data is assumed to be known
	Data   Buffer        `json:"data,omitempty"`   // deprecated: uncompressed data for backwards compatibility to ErrorBuffer
	Error  FirmwareError `json:"error,omitempty"`  // FirmwareErr*
}

type Devices struct {
	FWUPdVersion string                        `json:"fwupd_version"`
	Topology     []FWUPdDevice                 `json:"topology"`
	Releases     map[string][]FWUPdReleaseInfo `json:"releases,omitempty"`
}

type FWUPdDevice = map[string]interface{}
type FWUPdReleaseInfo = map[string]interface{}

type Agent struct {
	Release   string      `json:"release"`
	ImageSHA2 ErrorBuffer `json:"sha,omitempty"`
}

type NICList struct {
	List  []NIC         `json:"list,omitempty"`
	Error FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type NIC struct {
	Name  string        `json:"name,omitempty"`
	IPv4  []string      `json:"ipv4,omitempty"`
	IPv6  []string      `json:"ipv6,omitempty"`
	MAC   string        `json:"mac"`
	Error FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type OS struct {
	Hostname string        `json:"hostname"`
	Release  string        `json:"name"`
	Error    FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type SEVCommand struct {
	Command    uint32        `json:"command"` // firmware.SEV*
	ReadLength uint32        `json:"read_length"`
	Response   *Buffer       `json:"response,omitempty"`
	Error      FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type MEClientCommands struct {
	GUID     *uuid.UUID    `json:"guid,omitempty"`
	Address  string        `json:"address,omitempty"`
	Commands []MECommand   `json:"commands"`
	Error    FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type MECommand struct {
	Command  Buffer        `json:"command"`
	Response Buffer        `json:"response,omitempty"`
	Error    FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type UEFIVariable struct {
	Vendor string        `json:"vendor"`
	Name   string        `json:"name"`
	Value  *Buffer       `json:"value,omitempty"`
	Error  FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type MSR struct {
	MSR    uint32        `json:"msr,string"`
	Values []uint64      `json:"value,omitempty"`
	Error  FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type CPUIDLeaf struct {
	LeafEAX uint32        `json:"leaf_eax,string"`
	LeafECX uint32        `json:"leaf_ecx,string"`
	EAX     *uint32       `json:"eax,string,omitempty"`
	EBX     *uint32       `json:"ebx,string,omitempty"`
	ECX     *uint32       `json:"ecx,string,omitempty"`
	EDX     *uint32       `json:"edx,string,omitempty"`
	Error   FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type TPM2Property struct {
	Property uint32        `json:"property,string"`
	Value    *uint32       `json:"value,omitempty,string"`
	Error    FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type TPM2NVIndex struct {
	Index  uint32        `json:"index,string"`
	Public *NVPublic     `json:"public,omitempty"`
	Value  *Buffer       `json:"value,omitempty"`
	Error  FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type Memory struct {
	Values []MemoryRange `json:"values,omitempty"`
	Error  FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type MemoryRange struct {
	Start    uint64 `json:"start,string"`
	Bytes    uint64 `json:"bytes,string"`
	Reserved bool   `json:"reserved"`
}

type ACPITables struct {
	Blobs map[string]HashBlob `json:"blobs,omitempty"`
	Error FirmwareError       `json:"error,omitempty"` // FirmwareErr*
}

// deprecated
type ACPITablesV1 struct {
	Tables map[string]Buffer   `json:"tables,omitempty"` // deprecated: values are migrated to blobs
	Blobs  map[string]HashBlob `json:"blobs,omitempty" blobstore:"acpi"`
	Error  FirmwareError       `json:"error,omitempty"` // FirmwareErr*
}

type ErrorBuffer struct {
	Data  Buffer        `json:"data,omitempty"`
	Error FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type MACAddresses struct {
	Addresses []string      `json:"addrs"`
	Error     FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

type PCIConfigSpace struct {
	Bus      uint16        `json:"bus,string"`
	Device   uint16        `json:"device,string"`
	Function uint8         `json:"function,string"`
	Value    Buffer        `json:"value,omitempty"`
	Error    FirmwareError `json:"error,omitempty"` // FirmwareErr*
}

// /v2/enroll (apisrv)
type Enrollment struct {
	NameHint               string         `jsonapi:"attr,name_hint" json:"name_hint"`
	EndorsementKey         PublicKey      `jsonapi:"attr,endoresment_key" json:"endoresment_key"`
	EndorsementCertificate *Certificate   `jsonapi:"attr,endoresment_certificate" json:"endoresment_certificate"`
	Root                   PublicKey      `jsonapi:"attr,root" json:"root"`
	Keys                   map[string]Key `jsonapi:"attr,keys" json:"keys"`
	Cookie                 string         `jsonapi:"attr,cookie" json:"cookie"`
}

// /v2/enroll (apisrv)
type Key struct {
	Public                 PublicKey `json:"public"`
	CreationProof          Attest    `json:"certify_info"`
	CreationProofSignature Signature `json:"certify_signature"`
}

const EvidenceType = "evidence/1"

// /v2/attest (apisrv)
type Evidence struct {
	Type      string                       `jsonapi:"attr,type" json:"type"`
	Quote     Attest                       `jsonapi:"attr,quote" json:"quote"`
	Signature Signature                    `jsonapi:"attr,signature" json:"signature"`
	Algorithm string                       `jsonapi:"attr,algorithm" json:"algorithm"`
	PCRs      map[string]Buffer            `jsonapi:"attr,pcrs" json:"pcrs"`
	AllPCRs   map[string]map[string]Buffer `jsonapi:"attr,allpcrs" json:"allpcrs"`
	Firmware  FirmwareProperties           `jsonapi:"attr,firmware" json:"firmware"`
	Cookie    string                       `jsonapi:"attr,cookie" json:"cookie"`
}

// /v2/enroll (apisrv)
type EncryptedCredential struct {
	Name       string `jsonapi:"attr,name" json:"name"`
	KeyID      Buffer `jsonapi:"attr,key_id" json:"key_id"`
	Credential Buffer `jsonapi:"attr,credential" json:"credential"` // encrypted JWT
	Secret     Buffer `jsonapi:"attr,secret" json:"secret"`
	Nonce      Buffer `jsonapi:"attr,nonce" json:"nonce"`
}

// /v2/devices (apisrv)
type Appraisal struct {
	Id        string           `jsonapi:"primary,appraisals"     json:"id"`
	Appraised time.Time        `jsonapi:"attr,appraised,rfc3339" json:"appraised"`
	Expires   time.Time        `jsonapi:"attr,expires,rfc3339"   json:"expires"`
	Verdict   Verdict          `jsonapi:"attr,verdict"           json:"verdict"`
	Issues    *issuesv1.Issues `jsonapi:"attr,issues,omitempty"  json:"issues,omitempty"`
	Device    *Device          `jsonapi:"relation,device"        json:"device"`

	// soon to be deprecated
	Report   Report    `jsonapi:"attr,report,omitempty" json:"report,omitempty"`
	Received time.Time `jsonapi:"attr,received,rfc3339" json:"received"` // equals appraised now

	// internal
	linkSelfWeb string
}

const VerdictType = "verdict/3"
const VerdictTypeV2 = "verdict/2"
const VerdictTypeV1 = "verdict/1"

const (
	Unsupported = "unsupported"
	Trusted     = "trusted"
	Vulnerable  = "vulnerable"
)

// /v2/devices (apisrv)
type Verdict struct {
	Type string `json:"type"`

	Result             string `json:"result"`
	SupplyChain        string `json:"supply_chain"`
	Configuration      string `json:"configuration"`
	Firmware           string `json:"firmware"`
	Bootloader         string `json:"bootloader"`
	OperatingSystem    string `json:"operating_system"`
	EndpointProtection string `json:"endpoint_protection"`
}
type VerdictV2 struct {
	Type               string `json:"type"`
	Result             bool   `json:"result"`
	SupplyChain        bool   `json:"supply_chain"`
	Configuration      bool   `json:"configuration"`
	Firmware           bool   `json:"firmware"`
	Bootloader         bool   `json:"bootloader"`
	OperatingSystem    bool   `json:"operating_system"`
	EndpointProtection bool   `json:"endpoint_protection"`

	Bootchain bool `json:"bootchain"` // deprecated
}
type VerdictV1 struct {
	Type          string `json:"type"`
	Result        bool   `json:"result"`
	Bootchain     bool   `json:"bootchain"`
	Firmware      bool   `json:"firmware"`
	Configuration bool   `json:"configuration"`
}

const ReportType = "report/2"

// /v2/devices (apisrv)
type Report struct {
	Type        string       `json:"type"`
	Values      ReportValues `json:"values"`
	Annotations []Annotation `json:"annotations"`
}

type ReportValues struct {
	Host         Host    `json:"host"`
	SMBIOS       *SMBIOS `json:"smbios,omitempty"`
	UEFI         *UEFI   `json:"uefi,omitempty"` // deprecated
	TPM          *TPM    `json:"tpm,omitempty"`  // deprecated
	ME           *ME     `json:"me,omitempty"`   // deprecated
	AMT          *AMT    `json:"amt,omitempty"`  // deprecated
	SGX          *SGX    `json:"sgx,omitempty"`  // deprecated
	TXT          *TXT    `json:"txt,omitempty"`  // deprecated
	SEV          *SEV    `json:"sev,omitempty"`  // deprecated
	NICs         []NIC   `json:"nics,omitempty"`
	AgentRelease *string `json:"agent_release,omitempty"`
}

const (
	OSWindows = "windows"
	OSLinux   = "linux"
	OSUnknown = "unknown"
)

type CPUVendor string

const (
	IntelCPU CPUVendor = "GenuineIntel"
	AMDCPU   CPUVendor = "AuthenticAMD"
)

type Host struct {
	// Windows: <ProductName> <CurrentMajorVersionNumber>.<CurrentMinorVersionNumber> Build <CurrentBuild>
	// Linux: /etc/os-release PRETTY_NAME or lsb_release -d
	OSName    string    `json:"name"`
	Hostname  string    `json:"hostname"`
	OSType    string    `json:"type"` // OS*
	CPUVendor CPUVendor `json:"cpu_vendor"`
}

type SMBIOS struct {
	Manufacturer    string `json:"manufacturer"`
	Product         string `json:"product"`
	Serial          string `json:"serial,omitempty"`
	UUID            string `json:"uuid,omitempty"`
	BIOSReleaseDate string `json:"bios_release_date"`
	BIOSVendor      string `json:"bios_vendor"`
	BIOSVersion     string `json:"bios_version"`
}

const (
	EFICertificate = "certificate"
	EFIFingerprint = "fingerprint"
)

type EFISignature struct {
	Type        string     `json:"type"`              // EFIFingerprint or EFICertificate
	Subject     *string    `json:"subject,omitempty"` // certificate only
	Issuer      *string    `json:"issuer,omitempty"`  // certificate only
	Fingerprint string     `json:"fingerprint"`
	NotBefore   *time.Time `json:"not_before,omitempty"` // certificate only
	NotAfter    *time.Time `json:"not_after,omitempty"`  // certificate only
	Algorithm   *string    `json:"algorithm,omitempty"`  // certificate only
}

const (
	ModeSetup    = "setup"
	ModeAudit    = "audit"
	ModeUser     = "user"
	ModeDeployed = "deployed"
)

type LoadOption struct {
}

type UEFI struct {
	Mode          string          `json:"mode"` // Mode*
	SecureBoot    bool            `json:"secureboot"`
	PlatformKeys  *[]EFISignature `json:"platform_keys"`
	ExchangeKeys  *[]EFISignature `json:"exchange_keys"`
	PermittedKeys *[]EFISignature `json:"permitted_keys"`
	ForbiddenKeys *[]EFISignature `json:"forbidden_keys"`
	Drivers       *[]LoadOption   `json:"drivers"`
}

type TPM struct {
	Manufacturer string            `json:"manufacturer"`
	VendorID     string            `json:"vendor_id"`
	SpecVersion  string            `json:"spec_version"`
	EventLog     []TPMEvent        `json:"eventlog"`
	PCR          map[string]string `json:"pcr"`
}

const (
	ICU        = "ICU"
	TXE        = "TXE"
	ConsumerME = "Consumer CSME"
	BusinessME = "Business CSME"
	LightME    = "Light ME"
	SPS        = "SPS"
	UnknownME  = "Unrecognized"

	NormalState   = "normal"
	BootingState  = "booting"
	UpdatingState = "updating"
	FailureState  = "failure"
	UnsecureState = "insecure"
	RecoveryState = "recovery"
)

type ME struct {
	// MEI or config space
	Variant string `json:"variant"` // constants above
	// Config space
	State string `json:"state,omitempty"`
	// MEI
	Version           []int         `json:"version,omitempty"`
	FITCVersion       []int         `json:"fitc_version,omitempty"`
	RecoveryVersion   []int         `json:"recovery_version,omitempty"`
	AvailableFeatures []string      `json:"available_features,omitempty"`
	EnabledFeatures   []string      `json:"enabled_features,omitempty"`
	FlashLocked       *bool         `json:"flash_locked,omitempty"`
	Components        []MEComponent `json:"components,omitempty"`

	// ME 15+
	Tag          string `json:"tag,omitempty"`
	FIPS         *bool  `json:"fips,omitempty"`
	MeasuredBoot *bool  `json:"measured_boot,omitempty"`
}

type MEComponent struct {
	Name       string `json:"name"`
	Version    []int  `json:"version"`
	Vendor     string `json:"vendor"`
	SVN        int    `json:"svn"`
	MinimalSVN *int   `json:"min_svn,omitempty"`
}

type AMT struct {
	Enabled         bool              `json:"enabled"`
	Provisioned     bool              `json:"provisioned"`
	Type            string            `json:"type"`
	KVM             *AMTRemoteFeature `json:"kvm,omitempty"`
	SerialOverLAN   *AMTRemoteFeature `json:"sol,omitempty"`
	IDERedirect     *AMTRemoteFeature `json:"ider,omitempty"`
	RemoteAccess    *AMTRemoteFeature `json:"ra,omitempty"`
	WebUI           *AMTRemoteFeature `json:"webui,omitempty"`
	Fuse            bool              `json:"fuse"`
	FlashProtection bool              `json:"flash_protections"`
	Update          bool              `json:"update"`
	HardwareCrypto  bool              `json:"hw_crypto"`
	ZeroTouch       bool              `json:"zero_touch"`
	Versions        *AMTVersions      `json:"versions,omitempty"`
	AuditLog        *AMTAuditLog      `json:"audit_log,omitempty"`
}

type AMTVersions struct {
	AMT      []int  `json:"amt,omitempty"`
	Recovery []int  `json:"recovery,omitempty"`
	Flash    []int  `json:"flash,omitempty"`
	Netstack []int  `json:"netstack,omitempty"`
	Apps     []int  `json:"apps,omitempty"`
	HasQST   bool   `json:"has_qst"`
	HasAMT   bool   `json:"has_amt"`
	HasASF   bool   `json:"has_asf"`
	Vendor   string `json:"vendor"`
	BIOS     string `json:"bios"`
}

type AMTRemoteFeature struct {
	Enabled bool `json:"enabled"`
	Active  bool `json:"active"`
}

type AMTAuditRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
}

type AMTAuditLog struct {
	Signed bool             `json:"signed"`
	Log    []AMTAuditRecord `json:"log"`
}

type UniqueID struct {
	Enabled            bool   `json:"enabled"`
	OSControl          bool   `json:"os_control"`
	OEMID              string `json:"oem"`
	CSMEID             string `json:"csme"`
	RefubishingCounter int    `json:"counter"`
}

type SGX struct {
	Version          uint               `json:"version"`
	Enabled          bool               `json:"enabled"`
	FLC              bool               `json:"flc"`
	KSS              bool               `json:"kss"`
	MaxEnclaveSize32 uint               `json:"enclave_size_32"`
	MaxEnclaveSize64 uint               `json:"enclave_size_64"`
	EPC              []EnclavePageCache `json:"epc"`
}

type TXT struct {
	Ready bool `json:"ready"`
}

type SEV struct {
	Enabled bool   `json:"enabled"`
	Version []uint `json:"version"`
	SME     bool   `json:"sme"`
	ES      bool   `json:"es"`
	VTE     bool   `json:"vte"`
	SNP     bool   `json:"snp"`
	VMPL    bool   `json:"vmpl"`
	Guests  uint   `json:"guests"`
	MinASID uint   `json:"min_asid"`
}

type BootGuard struct {
	BootGuardSVN          int
	ACMSVN                int
	KeyManifestSVN        int
	RoTKeyManifestSVN     int
	OEMKeyManifestSVN     int
	BootPolicyManifestSVN int
}

// /v2/devices (apisrv)
type TPMEvent struct {
	PCR       uint   `json:"pcr"`
	Value     string `json:"value"`
	Algorithm uint   `json:"algorithm"`
	Note      string `json:"note"`
}

// /v2/devices (apisrv)
type EnclavePageCache struct {
	Base          uint64 `json:"base"`
	Size          uint64 `json:"size"`
	CIRProtection bool   `json:"cir_protection"`
}

// /v2/devices (apisrv)
type Annotation struct {
	Id       AnnotationID `json:"id"`
	Expected string       `json:"expected,omitempty"`
	Path     string       `json:"path"`
	Fatal    bool         `json:"fatal"`
}

const (
	BootchainCategory          = "bootchain" // deprecated
	FirmwareCategory           = "firmware"
	ConfigurationCategory      = "configuration"
	SupplyChainCategory        = "supply-chain"
	BootloaderCategory         = "bootloader"
	OperatingSystemCategory    = "operating-system"
	EndpointProtectionCategory = "endpoint-protection"
)

func (a Annotation) Category() string {
	if cat := annotationCategory(a.Id); cat != "" {
		return cat
	} else {
		return ConfigurationCategory
	}
}

func annotationCategory(ann AnnotationID) string {
	switch ann {

	// Supply Chain Incidents
	case AnnTCNotInLVFS:
		fallthrough
	case AnnTCDummyTPM:
		fallthrough
	case AnnTSCEKMismatch:
		fallthrough
	case AnnTSCPCRMismatch:
		fallthrough
	case AnnInternalNoTSC:
		fallthrough
	case AnnInvalidTPMSpecVersion:
		fallthrough
	case AnnTCEndorsementCertUnverified:
		fallthrough
	case AnnInternalNoBinarly:
		fallthrough
	case AnnTCTSCRequired:
		return SupplyChainCategory

	// Device Configuration Incidents
	case AnnTCUEFIConfigChanged:
		fallthrough
	case AnnTCUEFISecureBootOff:
		fallthrough
	case AnnTCUEFIdbxRemoved:
		fallthrough
	case AnnTCUEFIdbxIncomplete:
		fallthrough
	case AnnTCUEFIKeysChanged:
		return ConfigurationCategory

	// Firmware Incidents
	case AnnTCCSMENoUpgrade:
		fallthrough
	case AnnTCCSMEDowngrade:
		fallthrough
	case AnnTCPlatformNoUpgrade:
		fallthrough
	case AnnTCFirmwareSetChanged:
		fallthrough
	case AnnTCBootFailed:
		fallthrough
	case AnnTCUEFINoExit:
		fallthrough
	case AnnTCInvalidEventlog:
		fallthrough
	case AnnTCNoEventlog:
		fallthrough
	case AnnTCCSMEVersionVuln:
		fallthrough
	case AnnMEVersionInvalid:
		return FirmwareCategory

	// Bootloader Incidents
	case AnnTCShimMokListsChanged:
		fallthrough
	case AnnTCUEFIBootChanged:
		fallthrough
	case AnnTCGPTChanged:
		return BootloaderCategory

	// Operating System Incidents, string(ann)
	case AnnTCUnsecureWindowsBoot:
		fallthrough
	case AnnTCWindowsBootReplay:
		fallthrough
	case AnnTCGrub:
		return OperatingSystemCategory

	// Endpoint protection
	case AnnTCESETDisabled:
		fallthrough
	case AnnTCESETExcluded:
		fallthrough
	case AnnTCESETNotRunning:
		fallthrough
	case AnnTCESETManipulated:
		fallthrough
	case AnnTCBootAggregate:
		fallthrough
	case AnnTCRuntimeMeasurements:
		fallthrough
	case AnnTCInvalidIMAlog:
		fallthrough
	case AnnTCEPPRequired:
		return EndpointProtectionCategory

	// TODO: remove
	case AnnPCRInvalid:
		fallthrough
	case AnnPCRMissing:
		return BootchainCategory

	default:
		if strings.HasPrefix(string(ann), "brly-") {
			return SupplyChainCategory
		} else {
			return ""
		}
	}
}

type AnnotationID string

// keep in sync with allAnnotations in public_test.go
const (
	// host
	AnnHostname  = "host-hostname"
	AnnOSType    = "host-type"
	AnnCPUVendor = "host-cpu-ven"

	// smbios
	AnnNoSMBIOS           = "smbios-miss"
	AnnInvalidSMBIOS      = "smbios-inv"
	AnnSMBIOSType0Missing = "smbios-type0-miss"
	AnnSMBIOSType0Dup     = "smbios-type0-dup"
	AnnSMBIOSType1Missing = "smbios-type1-miss"
	AnnSMBIOSType1Dup     = "smbios-type1-dup"

	// uefi
	AnnNoEFI                = "uefi-vars-miss"
	AnnNoSecureBoot         = "uefi-secure-boot"
	AnnNoDeployedSecureBoot = "uefi-deployed-secure-boot"
	AnnMissingEventLog      = "uefi-eventlog-miss"
	AnnModeInvalid          = "uefi-mode-inv"
	AnnPKMissing            = "uefi-pk-miss"
	AnnPKInvalid            = "uefi-pk-inv"
	AnnKEKMissing           = "uefi-kek-miss"
	AnnKEKInvalid           = "uefi-kek-inv"
	AnnDBMissing            = "uefi-db-miss"
	AnnDBInvalid            = "uefi-db-inv"
	AnnDBxMissing           = "uefi-dbx-miss"
	AnnDBxInvalid           = "uefi-dbx-inv"

	// txt
	AnnNoTXTPubspace = "txt-public-miss"
	// css-*

	// sgx
	AnnNoSGX            = "sgx-missing"
	AnnSGXDisabled      = "sgx-disabled"
	AnnSGXCaps0Missing  = "sgx-cpuid0-miss"
	AnnSGXCaps1Missing  = "sgx-cpuid1-miss"
	AnnSGXCaps29Missing = "sgx-cpuid2-9-miss"

	// tpm
	AnnNoTPM                  = "tpm-miss"
	AnnNoTPMManufacturer      = "tpm-manuf-miss"
	AnnInvalidTPMManufacturer = "tpm-manuf-inv"
	AnnNoTPMVendorID          = "tpm-vid-miss"
	AnnInvalidTPMVendorID     = "tpm-vid-inv"
	AnnNoTPMSpecVersion       = "tpm-spec-miss"
	AnnInvalidTPMSpecVersion  = "tpm-spec-inv"
	AnnEventLogMissing        = "tpm-eventlog-miss"
	AnnEventLogInvalid        = "tpm-eventlog-inv"
	AnnEventLogBad            = "tpm-eventlog-bad"
	AnnPCRInvalid             = "tpm-pcr-inv"
	AnnPCRMissing             = "tpm-pcr-miss"

	// sev
	AnnNoSEV                 = "sev-miss"
	AnnSEVDisabled           = "sev-disabled"
	AnnPlatformStatusMissing = "sev-ps-miss"
	AnnPlatformStatusInvalid = "sev-ps-inv"

	// me
	AnnNoMEDevice            = "me-miss"
	AnnNoMEAccess            = "me-access"
	AnnMEConfigSpaceInvalid  = "me-inv"
	AnnMEVariantInvalid      = "me-variant-inv"
	AnnMEVersionMissing      = "me-version-miss"
	AnnMEVersionInvalid      = "me-version-inv"
	AnnMEFeaturesMissing     = "me-feat-miss"
	AnnMEFeaturesInvalid     = "me-feat-inv"
	AnnMEFWUPMissing         = "me-fwup-miss"
	AnnMEFWUPInvalid         = "me-fwup-inv"
	AnnMEBadChecksum         = "me-csum"
	AnnMEBUPFailure          = "me-bup"
	AnnMEFlashFailure        = "me-flash"
	AnnMEFSCorruption        = "me-fs"
	AnnMECPUInvalid          = "me-cpu-inv"
	AnnMEUncatError          = "me-unk"
	AnnMEImageError          = "me-image"
	AnnMEFatalError          = "me-fatal"
	AnnMEInvalidWorkingState = "me-ws-inv"
	AnnMEM0Error             = "me-m0"
	AnnMESPIDataInvalid      = "me-spi-inv"
	AnnMEUpdating            = "me-updating"
	AnnMEHalted              = "me-halt"
	AnnMEManufacturingMode   = "me-manuf"
	AnnMEUnlocked            = "me-unlocked"
	AnnMEDebugMode           = "me-debug"
	AnnMEResetRequest        = "me-reset-req"
	AnnMEMBEXRequest         = "me-mbex"
	AnnMECPUReplaced         = "me-cpurepl"
	AnnMESafeBoot            = "me-safe"
	AnnMEWasReset            = "me-was-reset"
	AnnMEDisabled            = "me-disabled"
	AnnMEBooting             = "me-booting"
	AnnMEInRecovery          = "me-recovery"
	AnnMESVNDisabled         = "me-svn-disabled"
	AnnMESVNUpdated          = "me-svn-updated"
	AnnMESVNOutdated         = "me-svn-outdated"

	// amt
	AnnAMTSKUInvalid               = "amt-sku-inv"
	AnnAMTVersionInvalid           = "amt-ver-inv"
	AnnAMTVersionMissing           = "amt-ver-miss"
	AnnAMTAuditLogInvalid          = "amt-log-inv"
	AnnAMTAuditLogMissing          = "amt-log-miss"
	AnnAMTAuditLogSignatureInvalid = "amt-log-sig-inv"
	AnnAMTTypeInvalid              = "amt-ty-inv"
	AnnAMTTypeMissing              = "amt-ty-miss"

	// trust chain
	AnnTCCSMENoUpgrade             = "tc-csme-no-upgrade"
	AnnTCCSMEDowngrade             = "tc-csme-downgrade"
	AnnTCPlatformNoUpgrade         = "tc-platform-no-upgrade"
	AnnTCFirmwareSetChanged        = "tc-oprom-set-changed"
	AnnTCUEFIConfigChanged         = "tc-uefi-config-changed"
	AnnTCUEFIBootChanged           = "tc-uefi-boot-changed"
	AnnTCUEFINoExit                = "tc-uefi-no-exit"
	AnnTCGPTChanged                = "tc-gpt-changed"
	AnnTCUEFIKeysChanged           = "tc-uefi-keys-changed"
	AnnTCUEFISecureBootOff         = "tc-uefi-secure-boot-off"
	AnnTCUEFIdbxRemoved            = "tc-uefi-dbx-removed"
	AnnTCShimMokListsChanged       = "tc-shim-moklist-changed"
	AnnTCGrub                      = "tc-grub-os-changed"
	AnnTCBootFailed                = "tc-boot-failed"
	AnnTCNoEventlog                = "tc-no-eventlog"
	AnnTCInvalidEventlog           = "tc-invalid-eventlog"
	AnnTCDummyTPM                  = "tc-dummy-tpm"
	AnnTCEndorsementCertUnverified = "tc-endorsement-cert-unverified"
	AnnTCBootAggregate             = "tc-ima-boot-aggregate"
	AnnTCRuntimeMeasurements       = "tc-ima-runtime-measurements"
	AnnTCInvalidIMAlog             = "tc-invalid-imalog"

	// non-fatal for now
	AnnTCUEFIdbxIncomplete = "tc-dbx-incomplete"
	AnnTCNotInLVFS         = "tc-not-in-lvfs"
	AnnTCCSMEVersionVuln   = "tc-csme-vulnerable"

	// Binarly vulnerabilities
	// created with:
	// cat fwhunt_report_linear.json | grep Name | cut -d : -f 2 | tr -d " \"," | xargs -I {} bash -c "echo Ann`echo {} | tr -d -` = \\\"`echo {} | tr [:upper:] [:lower:]`\\\""
	AnnBRLY2021033                  = "brly-2021-033"
	AnnBRLYESPecter                 = "brly-especter"
	AnnBRLY2021040                  = "brly-2021-040"
	AnnBRLY2021013                  = "brly-2021-013"
	AnnBRLY2022004                  = "brly-2022-004"
	AnnBRLYMoonBounceCOREDXE        = "brly-moonbounce-core-dxe"
	AnnBRLY2021004                  = "brly-2021-004"
	AnnBRLY2021024                  = "brly-2021-024"
	AnnBRLY2022027                  = "brly-2022-027"
	AnnBRLY2022014                  = "brly-2022-014"
	AnnBRLY2021009                  = "brly-2021-009"
	AnnBRLY2021036                  = "brly-2021-036"
	AnnBRLY2021011                  = "brly-2021-011"
	AnnBRLYThinkPwn                 = "brly-thinkpwn"
	AnnBRLY2021042                  = "brly-2021-042"
	AnnBRLY2021003                  = "brly-2021-003"
	AnnBRLY2021035                  = "brly-2021-035"
	AnnBRLY2021053                  = "brly-2021-053"
	AnnBRLY2021037                  = "brly-2021-037"
	AnnBRLYIntelBSSADFTINTELSA00525 = "brly-intel-bssa-dft"
	AnnBRLY2021045                  = "brly-2021-045"
	AnnBRLY2022011                  = "brly-2022-011"
	AnnBRLY2021005                  = "brly-2021-005"
	AnnBRLY2021012                  = "brly-2021-012"
	AnnBRLY2021022                  = "brly-2021-022"
	AnnBRLYLojaxSecDxe              = "brly-lojax-secdxe"
	AnnBRLYUsbRtCVE20175721         = "brly-usbrt-cve-2017-5721"
	AnnBRLY2021027                  = "brly-2021-027"
	AnnBRLY2021029                  = "brly-2021-029"
	AnnBRLY2021043                  = "brly-2021-043"
	AnnBRLY2021051                  = "brly-2021-051"
	AnnBRLY2021008                  = "brly-2021-008"
	AnnBRLY2021007                  = "brly-2021-007"
	AnnBRLY2021018                  = "brly-2021-018"
	AnnBRLY2021038                  = "brly-2021-038"
	AnnBRLYUsbRtSwSmiCVE202012301   = "brly-usbrt-swsmi-cve-2020-12301"
	AnnBRLY2021010                  = "brly-2021-010"
	AnnBRLY2021020                  = "brly-2021-020"
	AnnBRLY2021034                  = "brly-2021-034"
	AnnBRLY2021039                  = "brly-2021-039"
	AnnBRLYMosaicRegressor          = "brly-mosaicregressor"
	AnnBRLY2021028                  = "brly-2021-028"
	AnnBRLY2022013                  = "brly-2022-013"
	AnnBRLY2021006                  = "brly-2021-006"
	AnnBRLY2021026                  = "brly-2021-026"
	AnnBRLY2021021                  = "brly-2021-021"
	AnnBRLY2021023                  = "brly-2021-023"
	AnnBRLY2021019                  = "brly-2021-019"
	AnnBRLY2022016                  = "brly-2022-016"
	AnnBRLYRkloader                 = "brly-rkloader"
	AnnBRLYUsbRtUsbSmiCVE202012301  = "brly-usbrt-usbsmi-cve-2020-12301"
	AnnBRLY2021017                  = "brly-2021-017"
	AnnBRLY2021014                  = "brly-2021-014"
	AnnBRLY2021047                  = "brly-2021-047"
	AnnBRLY2022028RsbStuffing       = "brly-2022-028-rsbstuffing"
	AnnBRLY2021031                  = "brly-2021-031"
	AnnBRLY2021050                  = "brly-2021-050"
	AnnBRLY2021015                  = "brly-2021-015"
	AnnBRLYUsbRtINTELSA00057        = "brly-usbrt-intel-sa-00057"
	AnnBRLY2021016                  = "brly-2021-016"
	AnnBRLY2022015                  = "brly-2022-015"
	AnnBRLY2021041                  = "brly-2021-041"
	AnnBRLY2022009                  = "brly-2022-009"
	AnnBRLY2022010                  = "brly-2022-010"
	AnnBRLY2022012                  = "brly-2022-012"
	AnnBRLY2021046                  = "brly-2021-046"
	AnnBRLY2021032                  = "brly-2021-032"
	AnnBRLY2021030                  = "brly-2021-030"
	AnnBRLY2021025                  = "brly-2021-025"
	AnnBRLY2021001                  = "brly-2021-001"

	AnnTCESETNotRunning      = "tc-eset-not-running" // ELAM driver missing or PPL not running
	AnnTCESETDisabled        = "tc-eset-disabled"    // Linux module disabled
	AnnTCESETExcluded        = "tc-eset-excluded"    // Files/processes added to sysfs exclution list
	AnnTCESETManipulated     = "tc-eset-manipulated" // Files used by ESET were changed
	AnnTCUnsecureWindowsBoot = "tc-unsecure-windows-boot"
	AnnTCWindowsBootReplay   = "tc-windows-boot-replay"
	AnnTSCEKMismatch         = "tsc-ek-missmatch"
	AnnTSCPCRMismatch        = "tsc-pcr-mismatch"

	AnnDeviceFirmwareOutdated = "fw-outdated"

	AnnTCTSCRequired = "tc-tsc-required"
	AnnTCEPPRequired = "tc-epp-required"

	// HACK: we use these to signal that no supply chain check was done. They are
	// filtered out before returning the appraisal to the client
	AnnInternalNoTSC       = "internal-no-tsc"
	AnnInternalNoBinarly   = "internal-no-binarly"
	AnnInternalNoIMA       = "internal-no-ima"
	AnnInternalNoELAM      = "internal-no-elam-ppl"
	AnnInternalNoESETLinux = "internal-no-eset-linux"
)

var AnnFatal = map[AnnotationID]bool{
	// host
	AnnHostname:  false,
	AnnOSType:    false,
	AnnCPUVendor: false,

	// smbios
	AnnNoSMBIOS:           false,
	AnnInvalidSMBIOS:      false,
	AnnSMBIOSType0Missing: false,
	AnnSMBIOSType0Dup:     false,
	AnnSMBIOSType1Missing: false,
	AnnSMBIOSType1Dup:     false,

	// uefi
	AnnNoEFI:                false,
	AnnNoSecureBoot:         false,
	AnnNoDeployedSecureBoot: false,
	AnnMissingEventLog:      false,
	AnnModeInvalid:          false,
	AnnPKMissing:            false,
	AnnPKInvalid:            false,
	AnnKEKMissing:           false,
	AnnKEKInvalid:           false,
	AnnDBMissing:            false,
	AnnDBInvalid:            false,
	AnnDBxMissing:           false,
	AnnDBxInvalid:           false,

	// txt
	AnnNoTXTPubspace: false,
	// css-*

	// sgx
	AnnNoSGX:            false,
	AnnSGXDisabled:      false,
	AnnSGXCaps0Missing:  false,
	AnnSGXCaps1Missing:  false,
	AnnSGXCaps29Missing: false,

	// tpm
	AnnNoTPM:                  false,
	AnnNoTPMManufacturer:      false,
	AnnInvalidTPMManufacturer: false,
	AnnNoTPMVendorID:          false,
	AnnInvalidTPMVendorID:     false,
	AnnNoTPMSpecVersion:       false,
	AnnInvalidTPMSpecVersion:  false,
	AnnEventLogMissing:        false,
	AnnEventLogInvalid:        false,
	AnnEventLogBad:            false,
	AnnPCRInvalid:             false,
	AnnPCRMissing:             false,

	// sev
	AnnNoSEV:                 false,
	AnnSEVDisabled:           false,
	AnnPlatformStatusMissing: false,
	AnnPlatformStatusInvalid: false,

	// me
	AnnNoMEDevice:            false,
	AnnNoMEAccess:            false,
	AnnMEConfigSpaceInvalid:  false,
	AnnMEVariantInvalid:      false,
	AnnMEVersionMissing:      false,
	AnnMEVersionInvalid:      false,
	AnnMEFeaturesMissing:     false,
	AnnMEFeaturesInvalid:     false,
	AnnMEFWUPMissing:         false,
	AnnMEFWUPInvalid:         false,
	AnnMEBadChecksum:         false,
	AnnMEBUPFailure:          false,
	AnnMEFlashFailure:        false,
	AnnMEFSCorruption:        false,
	AnnMECPUInvalid:          false,
	AnnMEUncatError:          false,
	AnnMEImageError:          false,
	AnnMEFatalError:          false,
	AnnMEInvalidWorkingState: false,
	AnnMEM0Error:             false,
	AnnMESPIDataInvalid:      false,
	AnnMEUpdating:            false,
	AnnMEHalted:              false,
	AnnMEManufacturingMode:   false,
	AnnMEUnlocked:            false,
	AnnMEDebugMode:           false,
	AnnMEResetRequest:        false,
	AnnMEMBEXRequest:         false,
	AnnMECPUReplaced:         false,
	AnnMESafeBoot:            false,
	AnnMEWasReset:            false,
	AnnMEDisabled:            false,
	AnnMEBooting:             false,
	AnnMEInRecovery:          false,
	AnnMESVNDisabled:         false,
	AnnMESVNUpdated:          false,
	AnnMESVNOutdated:         false,

	// amt
	AnnAMTSKUInvalid:               false,
	AnnAMTVersionInvalid:           false,
	AnnAMTVersionMissing:           false,
	AnnAMTAuditLogInvalid:          false,
	AnnAMTAuditLogMissing:          false,
	AnnAMTAuditLogSignatureInvalid: false,
	AnnAMTTypeInvalid:              false,
	AnnAMTTypeMissing:              false,

	// trust chain
	AnnTCCSMENoUpgrade:             true,
	AnnTCCSMEDowngrade:             true,
	AnnTCPlatformNoUpgrade:         true,
	AnnTCFirmwareSetChanged:        true,
	AnnTCUEFIConfigChanged:         true,
	AnnTCUEFIBootChanged:           true,
	AnnTCUEFINoExit:                false, // #1041
	AnnTCGPTChanged:                true,
	AnnTCUEFIKeysChanged:           true,
	AnnTCUEFISecureBootOff:         true,
	AnnTCUEFIdbxRemoved:            true,
	AnnTCShimMokListsChanged:       true,
	AnnTCGrub:                      true,
	AnnTCBootFailed:                true,
	AnnTCDummyTPM:                  true,
	AnnTSCEKMismatch:               true,
	AnnTCEndorsementCertUnverified: true,
	AnnTSCPCRMismatch:              true,
	AnnTCBootAggregate:             true,
	AnnTCRuntimeMeasurements:       true,
	AnnTCInvalidEventlog:           true,
	AnnTCInvalidIMAlog:             true,
	AnnTCESETNotRunning:            true,
	AnnTCESETDisabled:              true,
	AnnTCESETExcluded:              true,
	AnnTCESETManipulated:           true,
	AnnTCWindowsBootReplay:         true,

	AnnTCTSCRequired: true,
	AnnTCEPPRequired: true,

	// non-fatal
	AnnTCNoEventlog: false,

	// disabled for now
	AnnTCUEFIdbxIncomplete:   false,
	AnnTCNotInLVFS:           false,
	AnnTCCSMEVersionVuln:     false,
	AnnTCUnsecureWindowsBoot: false,

	AnnInternalNoTSC:       false,
	AnnInternalNoBinarly:   false,
	AnnInternalNoIMA:       false,
	AnnInternalNoELAM:      false,
	AnnInternalNoESETLinux: false,
}

func NewAnnotation(id AnnotationID) *Annotation {
	fatal, ok := AnnFatal[id]
	return &Annotation{Id: id, Fatal: ok && fatal}
}

var (
	ChangeEnroll     = "enroll"      // device
	ChangeRename     = "rename"      // device
	ChangeTag        = "tag"         // device
	ChangeAssociate  = "associate"   // device,policy
	ChangeTemplate   = "template"    // policy
	ChangeNew        = "new"         // policy
	ChangeInstaciate = "instanciate" // policy
	ChangeRevoke     = "revoke"      // policy
	ChangeRetire     = "retire"      // device
)

// /v2/changes
type Change struct {
	Id        string    `jsonapi:"primary,changes" json:"id"`
	Actor     *string   `jsonapi:"attr,actor,omitempty" json:"actor,omitempty"`
	Timestamp time.Time `jsonapi:"attr,timestamp,rfc3339" json:"timestamp"`
	Comment   *string   `jsonapi:"attr,comment,omitempty" json:"comment,omitempty"`
	Type      string    `jsonapi:"attr,type" json:"type"` // Change*
	Device    *Device   `jsonapi:"relation,devices,omitempty" json:"device,omitempty"`
}

const (
	StateNew           = "new"
	StateUnseen        = "unseen"
	StateVuln          = "vulnerable"
	StateTrusted       = "trusted"
	StateOutdated      = "outdated"
	StateRetired       = "retired"
	StateResurrectable = "resurrectable"
)

// /v2/devices
type Device struct {
	Id                    string                 `jsonapi:"primary,devices" json:"id"`
	Cookie                string                 `jsonapi:"attr,cookie,omitempty" json:"cookie,omitempty"`
	Name                  string                 `jsonapi:"attr,name" json:"name"`
	State                 string                 `jsonapi:"attr,state" json:"state"`
	Tags                  []*Tag                 `jsonapi:"relation,tags,omitempty" json:"tags,omitempty"`
	Policy                map[string]interface{} `jsonapi:"attr,policy" json:"policy"`
	Hwid                  string                 `jsonapi:"attr,hwid" json:"hwid"`
	AttestationInProgress *time.Time             `jsonapi:"attr,attestation_in_progress,rfc3339,omitempty" json:"attestation_in_progress,omitempty"`
	Replaces              []*Device              `jsonapi:"relation,replaces,omitempty" json:"replaces,omitempty"`       // deprecated
	ReplacedBy            []*Device              `jsonapi:"relation,replaced_by,omitempty" json:"replaced_by,omitempty"` // deprecated
	Appraisals            []*Appraisal           `jsonapi:"relation,appraisals,omitempty" json:"appraisals,omitempty"`

	// internal
	linkSelfWeb string
}

type DevicePolicy struct {
	EndpointProtection string `json:"endpoint_protection"`
	IntelTSC           string `json:"intel_tsc"`
}

// /v2/devices
type DevicePatch struct {
	Id      string        `jsonapi:"primary,devices" json:"id"`
	Name    *string       `jsonapi:"attr,name,omitempty" json:"name,omitempty"`
	Tags    []Tag         `jsonapi:"attr,tags,omitempty" json:"tags,omitempty"`
	Policy  *DevicePolicy `jsonapi:"attr,policy,omitempty" json:"policy,omitempty"`
	State   *string       `jsonapi:"attr,state,omitempty" json:"state,omitempty"`
	Comment *string       `jsonapi:"attr,comment,omitempty" json:"comment,omitempty"`
}

// /v2/policies
type PolicyCreation struct {
	Id         string     `jsonapi:"primary,policies" json:"id"`
	Name       string     `jsonapi:"attr,name,omitempty" json:"name"`
	Devices    []*Device  `jsonapi:"relation,devices,omitempty" json:"devices"`
	Cookie     *string    `jsonapi:"attr,cookie,omitempty" json:"cookie,omitempty"`
	ValidSince *time.Time `jsonapi:"attr,valid_from,omitempty,rfc3339" json:"valid_since,omitempty"`
	ValidUntil *time.Time `jsonapi:"attr,valid_until,omitempty,rfc3339" json:"valid_until,omitempty"`
	Comment    *string    `jsonapi:"attr,comment,omitempty" json:"comment,omitempty"`

	// policy.Template
	PCRTemplate  []string   `jsonapi:"attr,pcr_template,omitempty" json:"pcr_template,omitempty"`
	FWTemplate   []string   `jsonapi:"attr,fw_template,omitempty" json:"fw_template,omitempty"`
	RevokeActive *time.Time `jsonapi:"attr,revoke_active,omitempty,rfc3339" json:"revoke_active,omitempty"`

	// policy.New
	PCRs        map[string]interface{} `jsonapi:"attr,pcrs,omitempty" json:"pcrs,omitempty"`
	FWOverrides []string               `jsonapi:"attr,fw_overrides,omitempty" json:"fw_overrides,omitempty"`
}

// /v2/tags
type Tag struct {
	Id    string  `jsonapi:"primary,tags" json:"id"`
	Key   string  `jsonapi:"attr,key,omitempty" json:"key,omitempty"`
	Score float32 `jsonapi:"attr,score,omitempty" json:"score,omitempty"`
}

type IncidentStatsEntry struct {
	IssueTypeId     string    `json:"issue_type_id"`
	LatestOccurence time.Time `json:"latest_occurence"`
	DevicesAffected int       `json:"devices_affected"`
}

type RiskStatsEntry struct {
	IssueTypeId   string `json:"issue_type_id"`
	NumOccurences int    `json:"num_occurences"`
}
type DeviceStats struct {
	NumTrusted      int `json:"num_trusted"`
	NumWithIncident int `json:"num_with_incident"`
	NumAtRisk       int `json:"num_at_risk"`
	NumUnresponsive int `json:"num_unresponsive"`
}

// /v2/dashboard
type Dashboard struct {
	Id               string            `jsonapi:"primary,dashboard" json:"id"`
	IncidentCount    int               `jsonapi:"attr,incident_count" json:"incident_count"`
	IncidentDevCount int               `jsonapi:"attr,incident_dev_count" json:"incident_dev_count"`
	DeviceStats      *DeviceStats      `jsonapi:"attr,device_stats" json:"device_stats"`
	Risks            []*RiskStatsEntry `jsonapi:"attr,risks,omitempty" json:"risks,omitempty"`
}

// /v2/incidents
type Incidents struct {
	Id        string                `jsonapi:"primary,incidents" json:"id"`
	Incidents []*IncidentStatsEntry `jsonapi:"attr,incidents,omitempty" json:"incidents,omitempty"`
}

// /v2/risks
type Risks struct {
	Id    string            `jsonapi:"primary,risks" json:"id"`
	Risks []*RiskStatsEntry `jsonapi:"attr,risks,omitempty" json:"risks,omitempty"`
}
