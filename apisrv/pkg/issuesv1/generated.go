package issuesv1

import "encoding/json"

type Root map[string]interface{}

type Binarly struct {
	Common
}

const Firmware string = "firmware"
const Configuration string = "configuration"
const Bootloader string = "bootloader"
const OperatingSystem string = "operating-system"
const EndpointProtection string = "endpoint-protection"
const SupplyChain string = "supply-chain"
const BinarlyAspect string = "firmware"
const Brly2021017 string = "brly/2021-017"
const Brly2021001 string = "brly/2021-001"
const Brly2022011 string = "brly/2022-011"
const BrlyMoonbounceCoreDxe string = "brly/moonbounce-core-dxe"
const Brly2021012 string = "brly/2021-012"
const BrlyThinkpwn string = "brly/thinkpwn"
const Brly2022014 string = "brly/2022-014"
const BrlyRkloader string = "brly/rkloader"
const Brly2021025 string = "brly/2021-025"
const BrlyIntelBssaDft string = "brly/intel-bssa-dft"
const BrlyUsbrtIntelSa00057 string = "brly/usbrt-intel-sa-00057"
const Brly2021008 string = "brly/2021-008"
const Brly2021015 string = "brly/2021-015"
const Brly2022012 string = "brly/2022-012"
const Brly2021036 string = "brly/2021-036"
const Brly2021043 string = "brly/2021-043"
const Brly2022013 string = "brly/2022-013"
const Brly2021046 string = "brly/2021-046"
const BrlyUsbrtCve20175721 string = "brly/usbrt-cve-2017-5721"
const Brly2021009 string = "brly/2021-009"
const Brly2022028Rsbstuffing string = "brly/2022-028-rsbstuffing"
const Brly2021035 string = "brly/2021-035"
const Brly2021053 string = "brly/2021-053"
const Brly2021018 string = "brly/2021-018"
const Brly2021034 string = "brly/2021-034"
const Brly2021037 string = "brly/2021-037"
const Brly2021050 string = "brly/2021-050"
const Brly2021040 string = "brly/2021-040"
const Brly2022016 string = "brly/2022-016"
const BrlyLojaxSecdxe string = "brly/lojax-secdxe"
const Brly2021020 string = "brly/2021-020"
const Brly2021041 string = "brly/2021-041"
const Brly20210111 string = "brly/2021-011-1"
const Brly2021010 string = "brly/2021-010"
const Brly2021030 string = "brly/2021-030"
const Brly20210091 string = "brly/2021-009-1"
const Brly2021029 string = "brly/2021-029"
const Brly2021021 string = "brly/2021-021"
const Brly2021028 string = "brly/2021-028"
const Brly2021004 string = "brly/2021-004"
const Brly2021026 string = "brly/2021-026"
const Brly2021007 string = "brly/2021-007"
const Brly2021042 string = "brly/2021-042"
const Brly20210291 string = "brly/2021-029-1"
const Brly2022015 string = "brly/2022-015"
const Brly2022010 string = "brly/2022-010"
const Brly2021051 string = "brly/2021-051"
const Brly2021024 string = "brly/2021-024"
const Brly2021031 string = "brly/2021-031"
const Brly2021027 string = "brly/2021-027"
const Brly2021045 string = "brly/2021-045"
const Brly20210081 string = "brly/2021-008-1"
const Brly2021032 string = "brly/2021-032"
const Brly2021023 string = "brly/2021-023"
const Brly2021013 string = "brly/2021-013"
const Brly2021005 string = "brly/2021-005"
const Brly2021006 string = "brly/2021-006"
const Brly2021033 string = "brly/2021-033"
const Brly2021003 string = "brly/2021-003"
const Brly20210101 string = "brly/2021-010-1"
const Brly2022009 string = "brly/2022-009"
const Brly2021022 string = "brly/2021-022"
const Brly2021016 string = "brly/2021-016"
const Brly2021039 string = "brly/2021-039"
const Brly2021011 string = "brly/2021-011"
const BrlyUsbrtSwsmiCve202012301 string = "brly/usbrt-swsmi-cve-2020-12301"
const Brly2021001SwsmiLen65529 string = "brly/2021-001-swsmi-len-65529"
const Brly2021019 string = "brly/2021-019"
const BrlyMosaicregressor string = "brly/mosaicregressor"
const Brly2021014 string = "brly/2021-014"
const Brly2021038 string = "brly/2021-038"
const Brly2022004 string = "brly/2022-004"
const BrlyEspecter string = "brly/especter"
const BrlyUsbrtUsbsmiCve202012301 string = "brly/usbrt-usbsmi-cve-2020-12301"
const Brly2022027 string = "brly/2022-027"
const Brly2021047 string = "brly/2021-047"
const BinarlyIncident bool = false

func (i *Binarly) Id() string {
	return i.Common.Id
}

func (i *Binarly) Incident() bool {
	return i.Common.Incident
}

func (i *Binarly) Aspect() string {
	return i.Common.Aspect
}

type Common struct {
	Aspect   string `json:"aspect"`
	Id       string `json:"id"`
	Incident bool   `json:"incident"`
}

type CsmeDowngrade struct {
	Common
	Args struct {
		Combined   *CsmeDowngradeComponent   `json:"combined,omitempty"`
		Components []*CsmeDowngradeComponent `json:"components,omitempty"`
	} `json:"args"`
}

const CsmeDowngradeId string = "csme/downgrade"
const CsmeDowngradeIncident bool = true
const CsmeDowngradeAspect string = "firmware"

func (i *CsmeDowngrade) Id() string {
	return i.Common.Id
}

func (i *CsmeDowngrade) Incident() bool {
	return i.Common.Incident
}

func (i *CsmeDowngrade) Aspect() string {
	return i.Common.Aspect
}

type CsmeDowngradeComponent struct {
	After  string `json:"after"`
	Before string `json:"before"`
	Name   string `json:"name,omitempty"`
}

type CsmeNoUpdate struct {
	Common
	Args struct {
		Components []CsmeNoUpdateComponent `json:"components"`
	} `json:"args"`
}

const CsmeNoUpdateAspect string = "firmware"
const CsmeNoUpdateId string = "csme/no-update"
const CsmeNoUpdateIncident bool = true

func (i *CsmeNoUpdate) Id() string {
	return i.Common.Id
}

func (i *CsmeNoUpdate) Incident() bool {
	return i.Common.Incident
}

func (i *CsmeNoUpdate) Aspect() string {
	return i.Common.Aspect
}

type CsmeNoUpdateComponent struct {
	After   string `json:"after"`
	Before  string `json:"before"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type EsetDisabled struct {
	Common
}

const EsetDisabledIncident bool = true
const EsetDisabledAspect string = "endpoint-protection"
const EsetDisabledId string = "eset/disabled"

func (i *EsetDisabled) Id() string {
	return i.Common.Id
}

func (i *EsetDisabled) Incident() bool {
	return i.Common.Incident
}

func (i *EsetDisabled) Aspect() string {
	return i.Common.Aspect
}

type EsetExcludedSet struct {
	Common
	Args struct {
		Files     []string `json:"files"`
		Processes []string `json:"processes"`
	} `json:"args"`
}

const EsetExcludedSetId string = "eset/excluded-set"
const EsetExcludedSetIncident bool = true
const EsetExcludedSetAspect string = "endpoint-protection"

func (i *EsetExcludedSet) Id() string {
	return i.Common.Id
}

func (i *EsetExcludedSet) Incident() bool {
	return i.Common.Incident
}

func (i *EsetExcludedSet) Aspect() string {
	return i.Common.Aspect
}

type EsetManipulated struct {
	Common
	Args struct {
		Components []EsetManipulatedFile `json:"components"`
	} `json:"args"`
}

const EsetManipulatedAspect string = "endpoint-protection"
const EsetManipulatedId string = "eset/manipulated"
const EsetManipulatedIncident bool = true

func (i *EsetManipulated) Id() string {
	return i.Common.Id
}

func (i *EsetManipulated) Incident() bool {
	return i.Common.Incident
}

func (i *EsetManipulated) Aspect() string {
	return i.Common.Aspect
}

type EsetManipulatedFile struct {
	After  string `json:"after"`
	Before string `json:"before"`
	Path   string `json:"path"`
}

type EsetNotStarted struct {
	Common
	Args struct {
		Components []EsetNotStartedComponent `json:"components"`
	} `json:"args"`
}

const EsetNotStartedAspect string = "endpoint-protection"
const EsetNotStartedId string = "eset/not-started"
const EsetNotStartedIncident bool = true

func (i *EsetNotStarted) Id() string {
	return i.Common.Id
}

func (i *EsetNotStarted) Incident() bool {
	return i.Common.Incident
}

func (i *EsetNotStarted) Aspect() string {
	return i.Common.Aspect
}

type EsetNotStartedComponent struct {
	Path    string `json:"path"`
	Started bool   `json:"started"`
}

type FirmwareUpdate struct {
	Common
	Args struct {
		Updates []FirmwareUpdateUpdates `json:"updates"`
	} `json:"args"`
}

const FirmwareUpdateIncident bool = false
const FirmwareUpdateAspect string = "firmware"
const FirmwareUpdateId string = "fw/update"

func (i *FirmwareUpdate) Id() string {
	return i.Common.Id
}

func (i *FirmwareUpdate) Incident() bool {
	return i.Common.Incident
}

func (i *FirmwareUpdate) Aspect() string {
	return i.Common.Aspect
}

type FirmwareUpdateUpdates struct {
	Current string `json:"current"`
	Name    string `json:"name"`
	Next    string `json:"next"`
}

type GrubBootChanged struct {
	Common
	Args struct {
		After  GrubBootChangedConfig `json:"after"`
		Before GrubBootChangedConfig `json:"before"`
	} `json:"args"`
}

const GrubBootChangedIncident bool = true
const GrubBootChangedAspect string = "bootloader"
const GrubBootChangedId string = "grub/boot-changed"

func (i *GrubBootChanged) Id() string {
	return i.Common.Id
}

func (i *GrubBootChanged) Incident() bool {
	return i.Common.Incident
}

func (i *GrubBootChanged) Aspect() string {
	return i.Common.Aspect
}

type GrubBootChangedConfig struct {
	CommandLine []string `json:"command_line"`
	Initrd      string   `json:"initrd"`
	InitrdPath  string   `json:"initrd_path"`
	Kernel      string   `json:"kernel"`
	KernelPath  string   `json:"kernel_path"`
}

type ImaBootAggregate struct {
	Common
	Args struct {
		Computed string `json:"computed"`
		Logged   string `json:"logged"`
	} `json:"args"`
}

const ImaBootAggregateIncident bool = true
const ImaBootAggregateAspect string = "endpoint-protection"
const ImaBootAggregateId string = "ima/boot-aggregate"

func (i *ImaBootAggregate) Id() string {
	return i.Common.Id
}

func (i *ImaBootAggregate) Incident() bool {
	return i.Common.Incident
}

func (i *ImaBootAggregate) Aspect() string {
	return i.Common.Aspect
}

type ImaInvalidLog struct {
	Common
	Args struct {
		Pcr []ImaInvalidLogPcr `json:"pcr"`
	} `json:"args"`
}

const ImaInvalidLogAspect string = "endpoint-protection"
const ImaInvalidLogId string = "ima/invalid-log"
const ImaInvalidLogIncident bool = true

func (i *ImaInvalidLog) Id() string {
	return i.Common.Id
}

func (i *ImaInvalidLog) Incident() bool {
	return i.Common.Incident
}

func (i *ImaInvalidLog) Aspect() string {
	return i.Common.Aspect
}

type ImaInvalidLogPcr struct {
	Computed string `json:"computed"`
	Number   string `json:"number"`
	Quoted   string `json:"quoted"`
}

type ImaRuntimeMeasurements struct {
	Common
	Args struct {
		Files []ImaRuntimeMeasurementsFile `json:"files"`
	} `json:"args"`
}

const ImaRuntimeMeasurementsIncident bool = true
const ImaRuntimeMeasurementsAspect string = "endpoint-protection"
const ImaRuntimeMeasurementsId string = "ima/runtime-measurements"

func (i *ImaRuntimeMeasurements) Id() string {
	return i.Common.Id
}

func (i *ImaRuntimeMeasurements) Incident() bool {
	return i.Common.Incident
}

func (i *ImaRuntimeMeasurements) Aspect() string {
	return i.Common.Aspect
}

type ImaRuntimeMeasurementsFile struct {
	After  string `json:"after"`
	Before string `json:"before"`
	Path   string `json:"path"`
}

type Issues struct {
	Issues []json.RawMessage `json:"issues"`
	Type   string            `json:"type"`
}

type PolicyEndpointProtection struct {
	Common
}

const PolicyEndpointProtectionAspect string = "endpoint-protection"
const PolicyEndpointProtectionId string = "policy/endpoint-protection"
const PolicyEndpointProtectionIncident bool = true

func (i *PolicyEndpointProtection) Id() string {
	return i.Common.Id
}

func (i *PolicyEndpointProtection) Incident() bool {
	return i.Common.Incident
}

func (i *PolicyEndpointProtection) Aspect() string {
	return i.Common.Aspect
}

type PolicyIntelTsc struct {
	Common
}

const PolicyIntelTscId string = "policy/intel-tsc"
const PolicyIntelTscIncident bool = true
const PolicyIntelTscAspect string = "supply-chain"

func (i *PolicyIntelTsc) Id() string {
	return i.Common.Id
}

func (i *PolicyIntelTsc) Incident() bool {
	return i.Common.Incident
}

func (i *PolicyIntelTsc) Aspect() string {
	return i.Common.Aspect
}

type TpmDummy struct {
	Common
	Args struct{} `json:"args"`
}

const TpmDummyAspect string = "supply-chain"
const TpmDummyId string = "tpm/dummy"
const TpmDummyIncident bool = true

func (i *TpmDummy) Id() string {
	return i.Common.Id
}

func (i *TpmDummy) Incident() bool {
	return i.Common.Incident
}

func (i *TpmDummy) Aspect() string {
	return i.Common.Aspect
}

type TpmEndorsementCertUnverified struct {
	Common
	Args struct {
		EkIssuer  string `json:"ek_issuer,omitempty"`
		EkVendor  string `json:"ek_vendor,omitempty"`
		EkVersion string `json:"ek_version,omitempty"`
		Error     string `json:"error"`
		Vendor    string `json:"vendor"`
	} `json:"args"`
}

const TpmEndorsementCertUnverifiedId string = "tpm/endorsement-cert-unverified"
const TpmEndorsementCertUnverifiedIncident bool = true
const SanInvalid string = "san-invalid"
const SanMismatch string = "san-mismatch"
const NoEku string = "no-eku"
const InvalidCertificate string = "invalid-certificate"
const TpmEndorsementCertUnverifiedAspect string = "supply-chain"

func (i *TpmEndorsementCertUnverified) Id() string {
	return i.Common.Id
}

func (i *TpmEndorsementCertUnverified) Incident() bool {
	return i.Common.Incident
}

func (i *TpmEndorsementCertUnverified) Aspect() string {
	return i.Common.Aspect
}

type TpmInvalidEventlog struct {
	Common
	Args struct {
		Error string                   `json:"error"`
		Pcr   []*TpmInvalidEventlogPcr `json:"pcr,omitempty"`
	} `json:"args"`
}

const TpmInvalidEventlogAspect string = "firmware"
const TpmInvalidEventlogId string = "tpm/invalid-eventlog"
const TpmInvalidEventlogIncident bool = true
const FormatInvalid string = "format-invalid"
const PcrMismatch string = "pcr-mismatch"

func (i *TpmInvalidEventlog) Id() string {
	return i.Common.Id
}

func (i *TpmInvalidEventlog) Incident() bool {
	return i.Common.Incident
}

func (i *TpmInvalidEventlog) Aspect() string {
	return i.Common.Aspect
}

type TpmInvalidEventlogPcr struct {
	Computed string `json:"computed,omitempty"`
	Number   string `json:"number"`
	Quoted   string `json:"quoted,omitempty"`
}

type TpmNoEventlog struct {
	Common
	Args struct{} `json:"args"`
}

const TpmNoEventlogIncident bool = true
const TpmNoEventlogAspect string = "firmware"
const TpmNoEventlogId string = "tpm/no-eventlog"

func (i *TpmNoEventlog) Id() string {
	return i.Common.Id
}

func (i *TpmNoEventlog) Incident() bool {
	return i.Common.Incident
}

func (i *TpmNoEventlog) Aspect() string {
	return i.Common.Aspect
}

type TscEndorsementCertificate struct {
	Common
	Args struct {
		EkIssuer     string `json:"ek_issuer,omitempty"`
		EkSerial     string `json:"ek_serial,omitempty"`
		Error        string `json:"error"`
		HolderIssuer string `json:"holder_issuer,omitempty"`
		HolderSerial string `json:"holder_serial,omitempty"`
		XmlSerial    string `json:"xml_serial,omitempty"`
	} `json:"args"`
}

const HolderIssuer string = "holder-issuer"
const XmlSerial string = "xml-serial"
const HolderSerial string = "holder-serial"
const TscEndorsementCertificateAspect string = "supply-chain"
const TscEndorsementCertificateId string = "tsc/endorsement-certificate"
const TscEndorsementCertificateIncident bool = true

func (i *TscEndorsementCertificate) Id() string {
	return i.Common.Id
}

func (i *TscEndorsementCertificate) Incident() bool {
	return i.Common.Incident
}

func (i *TscEndorsementCertificate) Aspect() string {
	return i.Common.Aspect
}

type TscPcrValues struct {
	Common
	Args struct {
		Values []TscPcrValuesPcr `json:"values"`
	} `json:"args"`
}

const TscPcrValuesAspect string = "supply-chain"
const TscPcrValuesId string = "tsc/pcr-values"
const TscPcrValuesIncident bool = true

func (i *TscPcrValues) Id() string {
	return i.Common.Id
}

func (i *TscPcrValues) Incident() bool {
	return i.Common.Incident
}

func (i *TscPcrValues) Aspect() string {
	return i.Common.Aspect
}

type TscPcrValuesPcr struct {
	Number string `json:"number"`
	Quoted string `json:"quoted"`
	Tsc    string `json:"tsc"`
}

type UefiBootAppSet struct {
	Common
	Args struct {
		Apps []UefiBootAppSetApp `json:"apps"`
	} `json:"args"`
}

const UefiBootAppSetId string = "uefi/boot-app-set"
const UefiBootAppSetIncident bool = true
const UefiBootAppSetAspect string = "bootloader"

func (i *UefiBootAppSet) Id() string {
	return i.Common.Id
}

func (i *UefiBootAppSet) Incident() bool {
	return i.Common.Incident
}

func (i *UefiBootAppSet) Aspect() string {
	return i.Common.Aspect
}

type UefiBootAppSetApp struct {
	After  string `json:"after"`
	Before string `json:"before"`
	Path   string `json:"path"`
}

type UefiBootFailure struct {
	Common
	Args struct {
		Pcr0 string `json:"pcr0"`
		Pcr1 string `json:"pcr1"`
		Pcr2 string `json:"pcr2"`
		Pcr3 string `json:"pcr3"`
		Pcr4 string `json:"pcr4"`
		Pcr5 string `json:"pcr5"`
		Pcr6 string `json:"pcr6"`
		Pcr7 string `json:"pcr7"`
	} `json:"args"`
}

const UefiBootFailureAspect string = "firmware"
const UefiBootFailureId string = "uefi/boot-failure"
const UefiBootFailureIncident bool = true

func (i *UefiBootFailure) Id() string {
	return i.Common.Id
}

func (i *UefiBootFailure) Incident() bool {
	return i.Common.Incident
}

func (i *UefiBootFailure) Aspect() string {
	return i.Common.Aspect
}

type UefiBootOrder struct {
	Common
	Args struct {
		Variables []UefiBootOrderVariable `json:"variables"`
	} `json:"args"`
}

const UefiBootOrderIncident bool = true
const UefiBootOrderAspect string = "configuration"
const UefiBootOrderId string = "uefi/boot-order"

func (i *UefiBootOrder) Id() string {
	return i.Common.Id
}

func (i *UefiBootOrder) Incident() bool {
	return i.Common.Incident
}

func (i *UefiBootOrder) Aspect() string {
	return i.Common.Aspect
}

type UefiBootOrderVariable struct {
	After  string `json:"after"`
	Before string `json:"before"`
	Name   string `json:"name"`
}

type UefiGptChanged struct {
	Common
	Args struct {
		After      string                    `json:"after"`
		Before     string                    `json:"before"`
		Guid       string                    `json:"guid"`
		Partitions []UefiGptChangedPartition `json:"partitions"`
	} `json:"args"`
}

const UefiGptChangedIncident bool = true
const UefiGptChangedAspect string = "bootloader"
const UefiGptChangedId string = "uefi/gpt-changed"

func (i *UefiGptChanged) Id() string {
	return i.Common.Id
}

func (i *UefiGptChanged) Incident() bool {
	return i.Common.Incident
}

func (i *UefiGptChanged) Aspect() string {
	return i.Common.Aspect
}

type UefiGptChangedPartition struct {
	End   string `json:"end"`
	Guid  string `json:"guid"`
	Name  string `json:"name,omitempty"`
	Start string `json:"start"`
	Type  string `json:"type"`
}

type UefiIbbNoUpdate struct {
	Common
	Args struct {
		After       string `json:"after"`
		Before      string `json:"before"`
		ReleaseDate string `json:"release_date"`
		Vendor      string `json:"vendor"`
		Version     string `json:"version"`
	} `json:"args"`
}

const UefiIbbNoUpdateAspect string = "firmware"
const UefiIbbNoUpdateId string = "uefi/ibb-no-update"
const UefiIbbNoUpdateIncident bool = true

func (i *UefiIbbNoUpdate) Id() string {
	return i.Common.Id
}

func (i *UefiIbbNoUpdate) Incident() bool {
	return i.Common.Incident
}

func (i *UefiIbbNoUpdate) Aspect() string {
	return i.Common.Aspect
}

type UefiNoExitBootSrv struct {
	Common
	Args struct {
		Entered bool `json:"entered"`
	} `json:"args"`
}

const UefiNoExitBootSrvIncident bool = false
const UefiNoExitBootSrvAspect string = "firmware"
const UefiNoExitBootSrvId string = "uefi/no-exit-boot-srv"

func (i *UefiNoExitBootSrv) Id() string {
	return i.Common.Id
}

func (i *UefiNoExitBootSrv) Incident() bool {
	return i.Common.Incident
}

func (i *UefiNoExitBootSrv) Aspect() string {
	return i.Common.Aspect
}

type UefiOfficialDbx struct {
	Common
	Args struct {
		Fprs []string `json:"fprs"`
	} `json:"args"`
}

const UefiOfficialDbxAspect string = "configuration"
const UefiOfficialDbxId string = "uefi/official-dbx"
const UefiOfficialDbxIncident bool = false

func (i *UefiOfficialDbx) Id() string {
	return i.Common.Id
}

func (i *UefiOfficialDbx) Incident() bool {
	return i.Common.Incident
}

func (i *UefiOfficialDbx) Aspect() string {
	return i.Common.Aspect
}

type UefiOptionRomSet struct {
	Common
	Args struct {
		Devices []UefiOptionRomSetDevice `json:"devices"`
	} `json:"args"`
}

const UefiOptionRomSetAspect string = "firmware"
const UefiOptionRomSetId string = "uefi/option-rom-set"
const UefiOptionRomSetIncident bool = true

func (i *UefiOptionRomSet) Id() string {
	return i.Common.Id
}

func (i *UefiOptionRomSet) Incident() bool {
	return i.Common.Incident
}

func (i *UefiOptionRomSet) Aspect() string {
	return i.Common.Aspect
}

type UefiOptionRomSetDevice struct {
	Address string `json:"address,omitempty"`
	After   string `json:"after"`
	Before  string `json:"before"`
	Name    string `json:"name"`
	Vendor  string `json:"vendor"`
}

type UefiSecureBootDbx struct {
	Common
	Args struct {
		Fprs []string `json:"fprs"`
	} `json:"args"`
}

const UefiSecureBootDbxId string = "uefi/secure-boot-dbx"
const UefiSecureBootDbxIncident bool = true
const UefiSecureBootDbxAspect string = "configuration"

func (i *UefiSecureBootDbx) Id() string {
	return i.Common.Id
}

func (i *UefiSecureBootDbx) Incident() bool {
	return i.Common.Incident
}

func (i *UefiSecureBootDbx) Aspect() string {
	return i.Common.Aspect
}

type UefiSecureBootKeys struct {
	Common
	Args struct {
		Kek []*UefiSecureBootKeysCertificate `json:"kek,omitempty"`
		Pk  *UefiSecureBootKeysCertificate   `json:"pk,omitempty"`
	} `json:"args"`
}

const UefiSecureBootKeysIncident bool = true
const UefiSecureBootKeysAspect string = "firmware"
const UefiSecureBootKeysId string = "uefi/secure-boot-keys"

func (i *UefiSecureBootKeys) Id() string {
	return i.Common.Id
}

func (i *UefiSecureBootKeys) Incident() bool {
	return i.Common.Incident
}

func (i *UefiSecureBootKeys) Aspect() string {
	return i.Common.Aspect
}

type UefiSecureBootKeysCertificate struct {
	Fpr       string `json:"fpr"`
	Issuer    string `json:"issuer"`
	NotAfter  string `json:"not_after"`
	NotBefore string `json:"not_before"`
	Subject   string `json:"subject"`
}

type UefiSecureBootVariables struct {
	Common
	Args struct {
		AuditMode    string `json:"audit_mode"`
		DeployedMode string `json:"deployed_mode"`
		SecureBoot   string `json:"secure_boot"`
		SetupMode    string `json:"setup_mode"`
	} `json:"args"`
}

const UefiSecureBootVariablesAspect string = "configuration"
const UefiSecureBootVariablesId string = "uefi/secure-boot-variables"
const UefiSecureBootVariablesIncident bool = true

func (i *UefiSecureBootVariables) Id() string {
	return i.Common.Id
}

func (i *UefiSecureBootVariables) Incident() bool {
	return i.Common.Incident
}

func (i *UefiSecureBootVariables) Aspect() string {
	return i.Common.Aspect
}

type WindowsBootConfig struct {
	Common
	Args struct {
		BootDebugging         bool `json:"boot_debugging"`
		CodeIntegrityDisabled bool `json:"code_integrity_disabled,omitempty"`
		DepDisabled           bool `json:"dep_disabled,omitempty"`
		KernelDebugging       bool `json:"kernel_debugging"`
		TestSigning           bool `json:"test_signing"`
	} `json:"args"`
}

const WindowsBootConfigAspect string = "operating-system"
const WindowsBootConfigId string = "windows/boot-config"
const WindowsBootConfigIncident bool = false

func (i *WindowsBootConfig) Id() string {
	return i.Common.Id
}

func (i *WindowsBootConfig) Incident() bool {
	return i.Common.Incident
}

func (i *WindowsBootConfig) Aspect() string {
	return i.Common.Aspect
}

type WindowsBootCounterReplay struct {
	Common
	Args struct {
		Latest   string `json:"latest"`
		Received string `json:"received"`
	} `json:"args"`
}

const WindowsBootCounterReplayAspect string = "operating-system"
const WindowsBootCounterReplayId string = "windows/boot-counter-replay"
const WindowsBootCounterReplayIncident bool = true

func (i *WindowsBootCounterReplay) Id() string {
	return i.Common.Id
}

func (i *WindowsBootCounterReplay) Incident() bool {
	return i.Common.Incident
}

func (i *WindowsBootCounterReplay) Aspect() string {
	return i.Common.Aspect
}

type WindowsBootLogQuotes struct {
	Common
	Args struct {
		Error string `json:"error"`
		Log   int64  `json:"log,omitempty"`
	} `json:"args"`
}

const WindowsBootLogQuotesId string = "windows/boot-log"
const WindowsBootLogQuotesIncident bool = true
const MissingTrustPoint string = "missing-trust-point"
const WrongFormat string = "wrong-format"
const WrongSignature string = "wrong-signature"
const WrongQuote string = "wrong-quote"
const WindowsBootLogQuotesAspect string = "operating-system"

func (i *WindowsBootLogQuotes) Id() string {
	return i.Common.Id
}

func (i *WindowsBootLogQuotes) Incident() bool {
	return i.Common.Incident
}

func (i *WindowsBootLogQuotes) Aspect() string {
	return i.Common.Aspect
}
