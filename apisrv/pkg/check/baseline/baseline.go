package baseline

import (
	"errors"
	"sort"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

const (
	BaselineType    = "baseline/3"
	BaselineTypeV23 = "baseline/2.3"
	BaselineTypeV22 = "baseline/2.2"
	BaselineTypeV2  = "baseline/2"
	BaselineTypeV1  = "baseline/1"
)

var (
	ErrUnknownRowVersion = errors.New("unknown version")
)

type valuesV1 struct {
	Type string `json:"type"`

	// csmeRuntimeMeasurments
	CSMERuntime  api.Buffer `json:"csme_runtime"`
	CSMEVersion  []int      `json:"csme_version"`
	CSMEFITC     []int      `json:"csem_fitc"`
	CSMERecovery []int      `json:"csme_recovery"`

	// bootGuard
	BootGuardIBB    api.Buffer `json:"bootguard_ibb"`
	BIOSVersion     string     `json:"bios_version"`
	BIOSReleaseDate string     `json:"bios_release_date"`

	// csmeVulnerableVersion
	AllowVulnerableCSME bool `json:"allow_vulnerable_csme"`

	// pciOptionROMs
	OptionROMs map[string]api.Buffer `json:"option_roms"`

	// uefiBootConfig
	BootVariables map[string]api.Buffer `json:"boot_variables"`

	// dbxRevokation
	RevokedKeyWhitelist []string `json:"revoked_key_whitelist"` // sorted!

	// uefiSetup
	Setup api.Buffer `json:"setup_variable"`

	// partitionTable
	GPT api.Buffer `json:"gpt"`

	// uefiSecureBoot
	PK                api.Buffer      `json:"pk"`
	KEK               api.Buffer      `json:"kek"`
	DBXContents       map[string]bool `json:"dbx"`
	SecureBootEnabled bool            `json:"secureboot_enabled"`

	// grub
	LinuxPath        string     `json:"linux_path"`
	LinuxDigest      api.Buffer `json:"linux_digest"`
	LinuxCommandLine []string   `json:"linux_command_line"`
	InitrdPath       string     `json:"initrd_path"`
	InitrdDigest     api.Buffer `json:"initrd_digest"`

	// shimKeys
	MokList  api.Buffer `json:"moklist"`
	MokListX api.Buffer `json:"moklistx"`

	// context less overrides
	AllowNoEventlog              bool  `json:"allow_no_eventlog"`
	AllowMissingExitBootServices bool  `json:"allow_exit_boot_services"`
	AllowBootFailure             []int `json:"allow_boot_failure"` // sorted!
	AllowMissingLVFS             bool  `json:"allow_missing_lvfs"`
}

type ValuesV2 struct {
	Type string `json:"type"`

	// csmeRuntimeMeasurments
	CSMEComponentHash    map[int]api.Buffer `json:"csme_component_hash"`
	CSMEComponentVersion map[int][]int      `json:"csme_component_version"`
	CSMEVersion          []int              `json:"csme_version"`
	CSMEFITC             []int              `json:"csem_fitc"`
	CSMERecovery         []int              `json:"csme_recovery"`

	// csmeSVNRollback
	CSMEComponentSVN map[int]int `json:"csme_component_svn"`
	CSMEComponentARB map[int]int `json:"csme_component_arb"`
	CSMEComponentVCN map[int]int `json:"csme_component_vcn"`

	// bootGuard
	BootGuardIBB    Hash   `json:"bootguard_ibb"`
	BIOSVersion     string `json:"bios_version"`
	BIOSReleaseDate string `json:"bios_release_date"`

	// csmeVulnerableVersion
	AllowVulnerableCSME bool `json:"allow_vulnerable_csme"`

	// embeddedFirmware
	EmbeddedFirmware map[string]Hash `json:"option_roms"`

	// uefiBootConfig
	BootVariables map[string]Hash `json:"boot_variables"`

	// uefiBootApp
	BootApplications map[string]Hash `json:"boot_applications"`

	// dbxRevokation
	RevokedKeyWhitelist []string `json:"revoked_key_whitelist"` // sorted!

	// uefiSetup
	Setup Hash `json:"setup_variable"`

	// partitionTable
	GPT Hash `json:"gpt"`

	// uefiSecureBoot
	PK                Hash            `json:"pk"`
	KEK               Hash            `json:"kek"`
	DBXContents       map[string]bool `json:"dbx"`
	SecureBootEnabled bool            `json:"secureboot_enabled"`

	// grub
	LinuxPath        string   `json:"linux_path"`
	LinuxDigest      Hash     `json:"linux_digest"`
	LinuxCommandLine []string `json:"linux_command_line"`
	InitrdPath       string   `json:"initrd_path"`
	InitrdDigest     Hash     `json:"initrd_digest"`

	// shimKeys
	MokList  Hash `json:"moklist"`
	MokListX Hash `json:"moklistx"`

	// TPM
	EndorsementCertificate *api.Certificate `json:"endorsement_certificate"`

	// context less overrides
	AllowNoEventlog                bool  `json:"allow_no_eventlog"`
	AllowInvalidEventlog           bool  `json:"allow_invalid_eventlog"`
	AllowInvalidImaLog             bool  `json:"allow_invalid_ima_log"` // added in 2.3
	AllowMissingExitBootServices   bool  `json:"allow_exit_boot_services"`
	AllowBootFailure               []int `json:"allow_boot_failure"` // sorted!
	AllowMissingLVFS               bool  `json:"allow_missing_lvfs"`
	AllowTSCPlaformRegsMismatch    bool  `json:"allow_tsc_pcr_mismatch"`
	AllowTSCEndorsementKeyMismatch bool  `json:"allow_tsc_ek_mismatch"`
	AllowEKCertificateUnverified   bool  `json:"allow_ek_cert_unverified"`
	AllowUnsecureWindowsBoot       bool  `json:"allow_unsecure_windows_boot"`
	AllowOutdatedFirmware          bool  `json:"allow_outdated_firmware"`

	// insecure TPM override
	AllowDummyTPM bool `json:"allow_dummy_tpm"`

	// binarly fwhunt vulnerability overrides
	AllowBinarlyVulnerabilityIDs []string `json:"allow_binarly_vulnerability_ids"`

	// ima
	BootAggregate    Hash            `json:"boot_aggregate"`
	FileMeasurements map[string]Hash `json:"file_measurements"`

	// linuxESET
	AllowDisabledESET     bool
	ESETExcludedFiles     []string // sorted!
	ESETExcludedProcesses []string // sorted!
	ESETFiles             map[string]Hash
	ESETKernelModule      Hash

	// windows boot counter
	BootCount string `json:"boot_count"`
}

type Values struct {
	Type string `json:"type"`

	// csmeRuntimeMeasurments
	CSMEComponentHash    map[int]api.Buffer `json:"csme_component_hash"`
	CSMEComponentVersion map[int][]int      `json:"csme_component_version"`
	CSMEVersion          []int              `json:"csme_version"`
	CSMEFITC             []int              `json:"csem_fitc"`
	CSMERecovery         []int              `json:"csme_recovery"`

	// csmeSVNRollback
	CSMEComponentSVN map[int]int `json:"csme_component_svn"`
	CSMEComponentARB map[int]int `json:"csme_component_arb"`
	CSMEComponentVCN map[int]int `json:"csme_component_vcn"`

	// bootGuard
	BootGuardIBB    Hash   `json:"bootguard_ibb"`
	BIOSVersion     string `json:"bios_version"`
	BIOSReleaseDate string `json:"bios_release_date"`

	// csmeVulnerableVersion
	AllowVulnerableCSME bool `json:"allow_vulnerable_csme"`

	// embeddedFirmware
	EmbeddedFirmware map[string]Hash `json:"option_roms"`

	// uefiBootConfig
	BootVariables map[string]Hash `json:"boot_variables"`

	// uefiBootApp
	BootApplications map[string]BootAppMeasurement `json:"boot_applications"`

	// dbxRevokation
	RevokedKeyWhitelist []string `json:"revoked_key_whitelist"` // sorted!

	// uefiSetup
	Setup Hash `json:"setup_variable"`

	// partitionTable
	GPT Hash `json:"gpt"`

	// uefiSecureBoot
	PK                Hash            `json:"pk"`
	KEK               Hash            `json:"kek"`
	DBXContents       map[string]bool `json:"dbx"`
	SecureBootEnabled bool            `json:"secureboot_enabled"`

	// grub
	LinuxPath        string   `json:"linux_path"`
	LinuxDigest      Hash     `json:"linux_digest"`
	LinuxCommandLine []string `json:"linux_command_line"`
	InitrdPath       string   `json:"initrd_path"`
	InitrdDigest     Hash     `json:"initrd_digest"`

	// shimKeys
	MokList  Hash `json:"moklist"`
	MokListX Hash `json:"moklistx"`

	// TPM
	EndorsementCertificate *api.Certificate `json:"endorsement_certificate"`

	// context less overrides
	AllowNoEventlog                bool  `json:"allow_no_eventlog"`
	AllowInvalidEventlog           bool  `json:"allow_invalid_eventlog"`
	AllowInvalidImaLog             bool  `json:"allow_invalid_ima_log"` // added in 2.3
	AllowMissingExitBootServices   bool  `json:"allow_exit_boot_services"`
	AllowBootFailure               []int `json:"allow_boot_failure"` // sorted!
	AllowMissingLVFS               bool  `json:"allow_missing_lvfs"`
	AllowTSCPlaformRegsMismatch    bool  `json:"allow_tsc_pcr_mismatch"`
	AllowTSCEndorsementKeyMismatch bool  `json:"allow_tsc_ek_mismatch"`
	AllowEKCertificateUnverified   bool  `json:"allow_ek_cert_unverified"`
	AllowUnsecureWindowsBoot       bool  `json:"allow_unsecure_windows_boot"`
	AllowOutdatedFirmware          bool  `json:"allow_outdated_firmware"`

	// insecure TPM override
	AllowDummyTPM bool `json:"allow_dummy_tpm"`

	// binarly fwhunt vulnerability overrides
	AllowBinarlyVulnerabilityIDs []string `json:"allow_binarly_vulnerability_ids"`

	// ima
	BootAggregate    Hash            `json:"boot_aggregate"`
	FileMeasurements map[string]Hash `json:"file_measurements"`

	// linuxESET
	AllowDisabledESET     bool
	ESETExcludedFiles     []string // sorted!
	ESETExcludedProcesses []string // sorted!
	ESETFiles             map[string]Hash
	ESETKernelModule      Hash

	// windows boot counter
	BootCount string `json:"boot_count"`
}

type BootAppMeasurement struct {
	Hash                         Hash      `json:"hash"`
	PinnedCertificateFingerprint *[32]byte `json:"cert_fp,omitempty"`
}

func New() *Values {
	return &Values{
		Type:             BaselineType,
		BootGuardIBB:     Hash{},
		EmbeddedFirmware: nil,
		BootVariables:    nil,
		DBXContents:      nil,
	}
}

func migrateHash(buf []byte) Hash {
	hash, err := NewHash(buf)
	if err == nil {
		return hash
	} else {
		return Hash{}
	}
}

func migrateMap(in map[string]api.Buffer) map[string]Hash {
	if in == nil {
		return nil
	}
	ret := make(map[string]Hash, len(in))
	for key, buf := range in {
		ret[key] = migrateHash(buf)
	}
	return ret
}

func FromRow(doc database.Document) (*Values, error) {
	if doc.IsNull() {
		return nil, nil
	}
	return parse(doc)
}

func migrateV1ToV23(bv1 *valuesV1) *ValuesV2 {
	vals := ValuesV2{
		Type: BaselineTypeV23,

		// csmeRuntimeMeasurments
		CSMEVersion:  bv1.CSMEVersion,
		CSMEFITC:     bv1.CSMEFITC,
		CSMERecovery: bv1.CSMERecovery,

		// bootGuard
		BootGuardIBB:    migrateHash(bv1.BootGuardIBB),
		BIOSVersion:     bv1.BIOSVersion,
		BIOSReleaseDate: bv1.BIOSReleaseDate,

		// csmeVulnerableVersion
		AllowVulnerableCSME: bv1.AllowVulnerableCSME,

		// embeddedFirmware
		EmbeddedFirmware: migrateMap(bv1.OptionROMs),

		// uefiBootConfig
		BootVariables: migrateMap(bv1.BootVariables),

		// dbxRevokation
		RevokedKeyWhitelist: bv1.RevokedKeyWhitelist,

		// uefiSetup
		Setup: migrateHash(bv1.Setup),

		// partitionTable
		GPT: migrateHash(bv1.GPT),

		// uefiSecureBoot
		PK:                migrateHash(bv1.PK),
		KEK:               migrateHash(bv1.KEK),
		DBXContents:       bv1.DBXContents,
		SecureBootEnabled: bv1.SecureBootEnabled,

		// grub
		LinuxPath:        bv1.LinuxPath,
		LinuxDigest:      migrateHash(bv1.LinuxDigest),
		LinuxCommandLine: bv1.LinuxCommandLine,
		InitrdPath:       bv1.InitrdPath,
		InitrdDigest:     migrateHash(bv1.InitrdDigest),

		// shimKeys
		MokList:  migrateHash(bv1.MokList),
		MokListX: migrateHash(bv1.MokListX),

		// context less overrides
		AllowNoEventlog:              bv1.AllowNoEventlog,
		AllowMissingExitBootServices: bv1.AllowMissingExitBootServices,
		AllowBootFailure:             bv1.AllowBootFailure,
		AllowMissingLVFS:             bv1.AllowMissingLVFS,
	}
	sort.Ints(vals.AllowBootFailure)
	sort.Strings(vals.RevokedKeyWhitelist)
	sort.Strings(vals.ESETExcludedFiles)
	sort.Strings(vals.ESETExcludedProcesses)
	return &vals
}

func MigrateV2ToV3(bv2 *ValuesV2) *Values {
	vals := Values{
		Type: BaselineType,

		// csmeRuntimeMeasurments
		CSMEComponentHash:    bv2.CSMEComponentHash,
		CSMEComponentVersion: bv2.CSMEComponentVersion,
		CSMEVersion:          bv2.CSMEVersion,
		CSMEFITC:             bv2.CSMEFITC,
		CSMERecovery:         bv2.CSMERecovery,

		// csmeSVNRollback
		CSMEComponentSVN: bv2.CSMEComponentSVN,
		CSMEComponentARB: bv2.CSMEComponentARB,
		CSMEComponentVCN: bv2.CSMEComponentVCN,

		// bootGuard
		BootGuardIBB:    bv2.BootGuardIBB,
		BIOSVersion:     bv2.BIOSVersion,
		BIOSReleaseDate: bv2.BIOSReleaseDate,

		// csmeVulnerableVersion
		AllowVulnerableCSME: bv2.AllowVulnerableCSME,

		// embeddedFirmware
		EmbeddedFirmware: bv2.EmbeddedFirmware,

		// uefiBootConfig
		BootVariables: bv2.BootVariables,

		// uefiBootApp
		BootApplications: migrateV2BootApplications(bv2.BootApplications),

		// dbxRevokation
		RevokedKeyWhitelist: bv2.RevokedKeyWhitelist,

		// uefiSetup
		Setup: bv2.Setup,

		// partitionTable
		GPT: bv2.GPT,

		// uefiSecureBoot
		PK:                bv2.PK,
		KEK:               bv2.KEK,
		DBXContents:       bv2.DBXContents,
		SecureBootEnabled: bv2.SecureBootEnabled,

		// grub
		LinuxPath:        bv2.LinuxPath,
		LinuxDigest:      bv2.LinuxDigest,
		LinuxCommandLine: bv2.LinuxCommandLine,
		InitrdPath:       bv2.InitrdPath,
		InitrdDigest:     bv2.InitrdDigest,

		// shimKeys
		MokList:  bv2.MokList,
		MokListX: bv2.MokListX,

		// TPM
		EndorsementCertificate: bv2.EndorsementCertificate,

		// context less overrides
		AllowNoEventlog:                bv2.AllowNoEventlog,
		AllowInvalidEventlog:           bv2.AllowInvalidEventlog,
		AllowInvalidImaLog:             bv2.AllowInvalidImaLog,
		AllowMissingExitBootServices:   bv2.AllowMissingExitBootServices,
		AllowBootFailure:               bv2.AllowBootFailure,
		AllowMissingLVFS:               bv2.AllowMissingLVFS,
		AllowTSCPlaformRegsMismatch:    bv2.AllowTSCPlaformRegsMismatch,
		AllowTSCEndorsementKeyMismatch: bv2.AllowTSCEndorsementKeyMismatch,
		AllowEKCertificateUnverified:   bv2.AllowEKCertificateUnverified,
		AllowUnsecureWindowsBoot:       bv2.AllowUnsecureWindowsBoot,
		AllowOutdatedFirmware:          bv2.AllowOutdatedFirmware,

		// insecure TPM override
		AllowDummyTPM: bv2.AllowDummyTPM,

		// binarly fwhunt vulnerability overrides
		AllowBinarlyVulnerabilityIDs: bv2.AllowBinarlyVulnerabilityIDs,

		// ima
		BootAggregate:    bv2.BootAggregate,
		FileMeasurements: bv2.FileMeasurements,

		// linuxESET
		AllowDisabledESET:     bv2.AllowDisabledESET,
		ESETExcludedFiles:     bv2.ESETExcludedFiles,
		ESETExcludedProcesses: bv2.ESETExcludedProcesses,
		ESETFiles:             bv2.ESETFiles,
		ESETKernelModule:      bv2.ESETKernelModule,

		// windows boot counter
		BootCount: bv2.BootCount,
	}

	sort.Ints(vals.AllowBootFailure)
	sort.Strings(vals.RevokedKeyWhitelist)
	sort.Strings(vals.ESETExcludedFiles)
	sort.Strings(vals.ESETExcludedProcesses)

	return &vals
}

func migrateV2BootApplications(bootApps map[string]Hash) map[string]BootAppMeasurement {
	bootAppsOut := make(map[string]BootAppMeasurement)
	for k, v := range bootApps {
		bootAppsOut[k] = BootAppMeasurement{
			Hash: v,
		}
	}
	return bootAppsOut
}

func parse(doc database.Document) (*Values, error) {
	switch doc.Type() {
	case "baseline/1":
		var bv1 valuesV1
		err := doc.Decode(&bv1)
		if err != nil {
			return nil, err
		}
		return MigrateV2ToV3(migrateV1ToV23(&bv1)), nil
	case "baseline/2":
		fallthrough
	case "baseline/2.1":
		fallthrough
	case "baseline/2.2":
		fallthrough
	case "baseline/2.3":
		var vals ValuesV2
		err := doc.Decode(&vals)
		if err != nil {
			return nil, err
		}

		return MigrateV2ToV3(&vals), nil

	case "baseline/3":
		var vals Values
		err := doc.Decode(&vals)
		if err != nil {
			return nil, err
		}
		return &vals, nil

	default:
		return nil, ErrUnknownRowVersion
	}
}

func ToRow(vals *Values) (database.Document, error) {
	return database.NewDocument(*vals)
}
