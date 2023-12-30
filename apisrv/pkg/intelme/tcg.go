package intelme

import (
	"bytes"
	"crypto"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
)

const (
	TPMEventSigMeasurment = "IntelCSxEEvent01\000\000\000\000"
	TPMEventSigInfo       = "IntelCSxEInfoEvent\000\000"
	TPMEventSigConfig     = "IntelCSMEAmtConfig\000\000"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
)

func MeasuredEntityToString(eventDataType uint8, measuredEntity uint8) string {
	switch eventDataType {
	case 0:
		fallthrough
	case 1:
		fallthrough
	case 2:
		switch measuredEntity {
		case PMCManifest:
			return "PMC Manifest"
		case IntelRBEManifest:
			return "Intel RBE Manifest"
		case ROTKeyManifest:
			return "ROT Key Manifest"
		case TCSSIOMManifest:
			return "TCSS IOM Manifest"
		case TCSSPhyManifest:
			return "TCSS Phy Manifest"
		case TCSSTBTManifest:
			return "TCSS TBT Manifest"
		case SynopsisPhyManifest:
			return "Synopsis Phys Manifest"
		case PCHCManifest:
			return "PCHC Manifest"
		case IDLMManifest:
			return "IDLM Manifest"
		case ISIIntelManifest:
			return "ISI Intel Manifest"
		case SurvivabilityEngineFirmwareManifest:
			return "SAM firmware Manifest"
		case SurvivabilityEnginePhyManifest:
			return "SAM Phy Manifest"
		case IUNITBootLoaderManifest:
			return "IUNIT Boot Loader Manifest"
		case AudioDSPExtROM:
			return "Audio DSP ROM Extension"
		case OEMISHManifest:
			return "OEM ISH Manifest"
		case OEMKeyManifest:
			return "OEM Key Manifest"
		case ISIOEMManifest:
			return "ISI OEM Manifest"
		case ESE:
			return "ESE"
		case DMU:
			return "DMU"
		case PUNIT:
			return "PUNIT"
		case ESExx:
			return "ESE++"
		case SOCPMC:
			return "SOC PMC"
		case SOCFirmware:
			return "SOC Firmware"
		case SOCSynopsisPhy:
			return "SOC Synopsis Phy"
		}

	case 3:
		switch measuredEntity {
		case 0:
			return "CSME Security Parameters"
		case 2:
			return "CSME OEM enabled capabilities"
		case 3:
			return "CSME Operation Mode"
		}
	}
	return fmt.Sprintf("Unknown entity %d for event type %d", measuredEntity, eventDataType)
}

type OperationMode uint8

func (mode OperationMode) String() string {
	switch mode {
	case Normal:
		return "normal"
	case DebugMode:
		return "debug"
	case SoftTemporaryDisableByBIOS:
		return "soft disable"
	case DisabledByHDA_SDO:
		return "HDA_SDO disable"
	case TemporaryDisabledForRefurbishing:
		return "disabled for refurbishing"
	case EnhancedDebugModeSetInImage:
		return "enhanced debug"
	default:
		return "unknown"
	}
}

const (
	InitializeManifest uint8 = 0
	ExtendManifest     uint8 = 1
	ManifestVersion    uint8 = 2
	ConfigurationData  uint8 = 3

	PMCManifest                         uint8 = 2   // Comet Lake
	IntelRBEManifest                    uint8 = 3   // Comet Lake
	ROTKeyManifest                      uint8 = 5   // Comet Lake
	TCSSIOMManifest                     uint8 = 6   // Tiger Lake
	TCSSPhyManifest                     uint8 = 7   // Tiger Lake
	TCSSTBTManifest                     uint8 = 8   // Tiger Lake
	SynopsisPhyManifest                 uint8 = 13  // Tiger Lake
	PCHCManifest                        uint8 = 14  // Alder Lake
	IDLMManifest                        uint8 = 15  // Comet Lake
	ISIIntelManifest                    uint8 = 16  // Tiger Lake
	SurvivabilityEngineFirmwareManifest uint8 = 17  // Rocket Lake only
	SurvivabilityEnginePhyManifest      uint8 = 18  // Rocket Lake only
	IUNITBootLoaderManifest             uint8 = 33  // Tiger Lake
	AudioDSPExtROM                      uint8 = 35  // Meteor Lake
	OEMISHManifest                      uint8 = 41  // Tiger Lake
	OEMKeyManifest                      uint8 = 45  // Comet Lake
	ISIOEMManifest                      uint8 = 58  // Tiger Lake
	ESE                                 uint8 = 192 // Meteor Lake
	DMU                                 uint8 = 193 // Meteor Lake
	PUNIT                               uint8 = 194 // Meteor Lake
	ESExx                               uint8 = 195 // Meteor Lake desktop only
	SOCPMC                              uint8 = 196 // Meteor Lake desktop only
	SOCFirmware                         uint8 = 197 // Meteor Lake desktop only
	SOCSynopsisPhy                      uint8 = 198 // Meteor Lake desktop only

	SecurityParameters     uint8 = 0
	OEMEnabledCapabilities uint8 = 2 // Legacy
	OperationModeID        uint8 = 3
	SKUInformation         uint8 = 4

	Normal                           OperationMode = 0
	DebugMode                        OperationMode = 2
	SoftTemporaryDisableByBIOS       OperationMode = 3
	DisabledByHDA_SDO                OperationMode = 4
	TemporaryDisabledForRefurbishing OperationMode = 5
	EnhancedDebugModeSetInImage      OperationMode = 7

	SlimSKUInfo      uint8 = 0
	ConsumerSKUInfo  uint8 = 1
	CorporateSKUInfo uint8 = 2
)

type taggedEvent struct {
	EventDataType    uint8
	MeasuredEntityID uint8
	// Reserved uint16
	// DataSize uint32
	Data []byte
}

type Event struct {
	eventlog.Event
	EventDataType    uint8
	MeasuredEntityID uint8
}

func (ev Event) RawEvent() eventlog.Event {
	return ev.Event
}

type InitializeManifestEvent struct {
	Event
}

type ExtendManifestEvent struct {
	Event
}

type ManifestVersionEvent struct {
	Event
	ManifestVersion ManifestVersionPayload
}

type SecurityParametersEvent struct {
	Event
	SecurityParameters SecurityParametersPayload
}

type OperationModeEvent struct {
	Event
	OperationMode
}

type OEMEnabledCapabilitiesEvent struct {
	Event
	OEMEnabledCapabilities []string
}

type SKUInformationEvent struct {
	Event
	SKUInformation uint8
}

const (
	// VerificationStatus
	Failed   uint8 = 1
	Passed   uint8 = 2
	NoSigned uint8 = 3
	Skipped  uint8 = 4
)

const (
	// BootSource
	BootFromSPI     int = 0
	BootFromUFSeMMC int = 1
)

type SecurityParametersPayload struct {
	SOCConfigLockFuse             bool
	EndOfManufacturing            bool
	ManageabilityHardwareDisabled bool
	BootSource                    int
	SPIRegionWriteLocked          bool
	SPIDescriptorLocked           bool
	RPMC_ENABLED                  bool
}

type ManifestVersionPayload struct {
	Version            [4]uint16
	TCBSVN             uint32
	ARBSVN             uint32
	VCN                uint32
	VerificationStatus uint8
	ManifestIdentifier uint8
}

const (
	SHA1Algorithm    = "sha1"
	SHA256Algorithm  = "sha256"
	SHA384Algorithm  = "sha384"
	UnknownAlgorithm = "unknown"
)

type FirmwareMeasurmentEvent struct {
	Signature       [20]byte
	ERHashAlgorithm string
	Events          []eventlog.TPMEvent
}

type FirmwareInfoEvent struct {
	Signature [20]byte
	Version   uint32
	VendorID  uint16
	DeviceID  uint16

	// Flags uint32
	HardwareRoT          bool
	InvalidState         bool
	UntrustedMeasurement bool
	InvalidMeasurment    bool
	LogUnavailable       bool
	FDOInvalidMeasurment bool
}

type FirmwareConfigEvent struct {
	Signature [20]byte
	// DataLength uint32

	// AMTStatus uint16
	AMTGloballyEnabled  bool
	MEBXPowerSet        bool
	AMTProvisioned      bool
	AMTProvisioningMode string
	ZeroTouch           bool
	KVM                 bool
	SerialOverLAN       bool
	USBRedirect         bool

	SecurePKISuffix          string
	CertificateHashAlgorithm string
	CertificateHash          [64]byte
}

func ParseInfoEvent(buf []byte) (*FirmwareInfoEvent, error) {
	if !bytes.Equal(buf[0:20], []byte(TPMEventSigInfo)) {
		return nil, ErrInvalidSignature
	}

	var info FirmwareInfoEvent
	rd := bytes.NewReader(buf)

	if _, err := rd.Read(info.Signature[:]); err != nil {
		return nil, err
	}
	if err := binary.Read(rd, binary.LittleEndian, &info.Version); err != nil {
		return nil, err
	}
	if err := binary.Read(rd, binary.LittleEndian, &info.VendorID); err != nil {
		return nil, err
	}
	if err := binary.Read(rd, binary.LittleEndian, &info.DeviceID); err != nil {
		return nil, err
	}

	var flags uint32
	if err := binary.Read(rd, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}
	info.HardwareRoT = flags&0b000001 != 0
	info.InvalidState = flags&0b000010 != 0
	info.UntrustedMeasurement = flags&0b000100 != 0
	info.InvalidMeasurment = flags&0b001000 != 0
	info.LogUnavailable = flags&0b010000 != 0
	info.FDOInvalidMeasurment = flags&0b100000 != 0

	return &info, nil
}

func parseTaggedEvent(rd io.Reader, index int, alg eventlog.HashAlg, ty eventlog.EventType) (eventlog.TPMEvent, error) {
	ev := Event{
		Event: eventlog.Event{
			Index: index,
			Alg:   alg,
			Type:  ty,
		},
	}
	if err := binary.Read(rd, binary.LittleEndian, &ev.EventDataType); err != nil {
		return nil, err
	}
	if err := binary.Read(rd, binary.LittleEndian, &ev.MeasuredEntityID); err != nil {
		return nil, err
	}
	var reserved uint16
	if err := binary.Read(rd, binary.LittleEndian, &reserved); err != nil {
		return nil, err
	}
	var sz uint32
	if err := binary.Read(rd, binary.LittleEndian, &sz); err != nil {
		return nil, err
	}
	ev.Data = make([]byte, sz)
	if _, err := rd.Read(ev.Data); err != nil {
		return nil, err
	}

	if ev.EventDataType == InitializeManifest {
		return InitializeManifestEvent{Event: ev}, nil
	} else if ev.EventDataType == ExtendManifest {
		return ExtendManifestEvent{Event: ev}, nil
	} else if ev.EventDataType == ManifestVersion && len(ev.Data) >= 22 {
		ev := ManifestVersionEvent{Event: ev}
		rd := bytes.NewReader(ev.Data)
		if err := binary.Read(rd, binary.LittleEndian, &ev.ManifestVersion.Version); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.LittleEndian, &ev.ManifestVersion.TCBSVN); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.LittleEndian, &ev.ManifestVersion.ARBSVN); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.LittleEndian, &ev.ManifestVersion.VCN); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.LittleEndian, &ev.ManifestVersion.VerificationStatus); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.LittleEndian, &ev.ManifestVersion.ManifestIdentifier); err != nil {
			return nil, err
		}
		return ev, nil
	} else if ev.EventDataType == ConfigurationData {
		switch ev.MeasuredEntityID {
		case SecurityParameters:
			if len(ev.Data) == 4 {
				ev := SecurityParametersEvent{Event: ev}
				var flags uint32
				if err := binary.Read(bytes.NewReader(ev.Data), binary.LittleEndian, &flags); err != nil {
					return nil, err
				}
				ev.SecurityParameters.SOCConfigLockFuse = flags&1 != 0
				ev.SecurityParameters.EndOfManufacturing = (flags>>1)&1 != 0
				ev.SecurityParameters.ManageabilityHardwareDisabled = (flags>>2)&1 != 0
				ev.SecurityParameters.BootSource = int((flags >> 3) & 1)
				ev.SecurityParameters.SPIRegionWriteLocked = (flags>>4)&1 != 0
				ev.SecurityParameters.SPIDescriptorLocked = (flags>>5)&1 != 0
				ev.SecurityParameters.RPMC_ENABLED = (flags>>6)&1 != 0
				return ev, nil
			}
		case OEMEnabledCapabilities:
			if len(ev.Data) == 4 {
				ev := OEMEnabledCapabilitiesEvent{Event: ev}
				var flags uint32
				if err := binary.Read(bytes.NewReader(ev.Data), binary.LittleEndian, &flags); err != nil {
					return nil, err
				}
				ev.OEMEnabledCapabilities = Features(flags)
				return ev, nil
			}
		case OperationModeID:
			if len(ev.Data) == 1 {
				return OperationModeEvent{
					Event:         ev,
					OperationMode: OperationMode(ev.Data[0]),
				}, nil
			}
		case SKUInformation:
			if len(ev.Data) == 1 {
				return SKUInformationEvent{
					Event:          ev,
					SKUInformation: ev.Data[0],
				}, nil
			}
		}
	}
	return ev, nil
}

func ParseMeasurmentEvent(event eventlog.NonHostInfoEvent) (*FirmwareMeasurmentEvent, error) {
	raw := event.RawEvent()
	if !bytes.Equal(raw.Data[0:20], []byte(TPMEventSigMeasurment)) {
		return nil, ErrInvalidSignature
	}

	var measure FirmwareMeasurmentEvent
	rd := bytes.NewReader(raw.Data)

	if _, err := rd.Read(measure.Signature[:]); err != nil {
		return nil, err
	}
	var numalg uint32
	var alg eventlog.HashAlg

	if err := binary.Read(rd, binary.LittleEndian, &numalg); err != nil {
		return nil, err
	}
	switch numalg {
	case 0:
		measure.ERHashAlgorithm = SHA1Algorithm
		alg = eventlog.HashSHA1
	case 2:
		measure.ERHashAlgorithm = SHA256Algorithm
		alg = eventlog.HashSHA256
	case 4:
		measure.ERHashAlgorithm = SHA384Algorithm
		alg = eventlog.HashSHA384
	default:
		measure.ERHashAlgorithm = UnknownAlgorithm
	}

	for {
		ev, err := parseTaggedEvent(rd, raw.Index, alg, raw.Type)
		if err != nil {
			if err == io.EOF {
				return &measure, nil
			}
			return nil, err
		}
		measure.Events = append(measure.Events, ev)
	}
}

func extendER(hash crypto.Hash, er []byte, data []byte) []byte {
	h := hash.New()
	if len(er) == 0 {
		er = make([]byte, h.Size())
	}
	h.Write(er)
	h.Write(data)
	return h.Sum(nil)
}

func ReplayER(alg eventlog.HashAlg, cometLake bool, events []eventlog.TPMEvent) []byte {
	var er []byte
	hash := alg.CryptoHash()
	sz := hash.New().Size()

	for _, event := range events {
		raw := event.RawEvent()
		if e, ok := event.(InitializeManifestEvent); cometLake && ok && e.MeasuredEntityID == IDLMManifest {
			if er == nil {
				er = raw.Data
			} else {
				return nil
			}
		} else {
			var padded []byte
			if len(raw.Data) < sz {
				padded = make([]byte, sz)
				copy(padded, raw.Data)
			} else {
				padded = raw.Data
			}
			er = extendER(hash, er, padded)
		}
	}

	in := bytes.NewReader(er)
	out := bytes.NewBuffer(nil)
	for {
		var reg uint32
		err := binary.Read(in, binary.LittleEndian, &reg)
		if err == io.EOF {
			return out.Bytes()
		} else if err != nil {
			return nil
		}
		binary.Write(out, binary.BigEndian, &reg)
	}
}

func ParseConfigEvent(buf []byte) (*FirmwareConfigEvent, error) {
	if !bytes.Equal(buf[0:20], []byte(TPMEventSigConfig)) {
		return nil, ErrInvalidSignature
	}

	var config FirmwareConfigEvent
	rd := bytes.NewReader(buf)

	if _, err := rd.Read(config.Signature[:]); err != nil {
		return nil, err
	}
	var dataLen uint32
	if err := binary.Read(rd, binary.LittleEndian, &dataLen); err != nil {
		return nil, err
	}
	if dataLen < 2+2+256+4+64 {
		return nil, io.EOF
	}

	var flags uint16
	if err := binary.Read(rd, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}

	config.AMTGloballyEnabled = flags&(1<<0) != 0
	config.MEBXPowerSet = flags&(1<<1) != 0
	config.AMTProvisioned = flags&(1<<2) != 0
	config.ZeroTouch = flags&(1<<4) != 0
	config.KVM = flags&(1<<8) != 0
	config.SerialOverLAN = flags&(1<<9) != 0
	config.USBRedirect = flags&(1<<10) != 0

	if flags&(1<<3) != 0 {
		config.AMTProvisioningMode = "enterprise"
	} else {
		config.AMTProvisioningMode = "none"
	}

	var suffix [256]byte
	if _, err := rd.Read(suffix[:]); err != nil {
		return nil, err
	}
	config.SecurePKISuffix = string(bytes.TrimRight(suffix[:], "\000"))

	if err := binary.Read(rd, binary.LittleEndian, &config.CertificateHashAlgorithm); err != nil {
		return nil, err
	}
	if _, err := rd.Read(config.CertificateHash[:]); err != nil {
		return nil, err
	}

	return &config, nil
}
