package eventlog

import (
	"bytes"
	"encoding/asn1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/x509"
)

// GUIDs representing the contents of an UEFI_SIGNATURE_LIST.
var (
	hashSHA256SigGUID        = EFIGUID{0xc1c41626, 0x504c, 0x4092, [8]byte{0xac, 0xa9, 0x41, 0xf9, 0x36, 0x93, 0x43, 0x28}}
	hashSHA1SigGUID          = EFIGUID{0x826ca512, 0xcf10, 0x4ac9, [8]byte{0xb1, 0x87, 0xbe, 0x01, 0x49, 0x66, 0x31, 0xbd}}
	hashSHA224SigGUID        = EFIGUID{0x0b6e5233, 0xa65c, 0x44c9, [8]byte{0x94, 0x07, 0xd9, 0xab, 0x83, 0xbf, 0xc8, 0xbd}}
	hashSHA384SigGUID        = EFIGUID{0xff3e5307, 0x9fd0, 0x48c9, [8]byte{0x85, 0xf1, 0x8a, 0xd5, 0x6c, 0x70, 0x1e, 0x01}}
	hashSHA512SigGUID        = EFIGUID{0x093e0fae, 0xa6c4, 0x4f50, [8]byte{0x9f, 0x1b, 0xd4, 0x1e, 0x2b, 0x89, 0xc1, 0x9a}}
	keyRSA2048SigGUID        = EFIGUID{0x3c5766e8, 0x269c, 0x4e34, [8]byte{0xaa, 0x14, 0xed, 0x77, 0x6e, 0x85, 0xb3, 0xb6}}
	certRSA2048SHA256SigGUID = EFIGUID{0xe2b36190, 0x879b, 0x4a3d, [8]byte{0xad, 0x8d, 0xf2, 0xe7, 0xbb, 0xa3, 0x27, 0x84}}
	certRSA2048SHA1SigGUID   = EFIGUID{0x67f8444f, 0x8743, 0x48f1, [8]byte{0xa3, 0x28, 0x1e, 0xaa, 0xb8, 0x73, 0x60, 0x80}}
	certX509SigGUID          = EFIGUID{0xa5c059a1, 0x94e4, 0x4aa7, [8]byte{0x87, 0xb5, 0xab, 0x15, 0x5c, 0x2b, 0xf0, 0x72}}
	certHashSHA256SigGUID    = EFIGUID{0x3bd2a492, 0x96c0, 0x4079, [8]byte{0xb4, 0x20, 0xfc, 0xf9, 0x8e, 0xf1, 0x03, 0xed}}
	certHashSHA384SigGUID    = EFIGUID{0x7076876e, 0x80c2, 0x4ee6, [8]byte{0xaa, 0xd2, 0x28, 0xb3, 0x49, 0xa6, 0x86, 0x5b}}
	certHashSHA512SigGUID    = EFIGUID{0x446dbf63, 0x2502, 0x4cda, [8]byte{0xbc, 0xfa, 0x24, 0x65, 0xd2, 0xb0, 0xfe, 0x9d}}
)

var (
	// https://github.com/rhboot/shim/blob/20e4d9486fcae54ee44d2323ae342ffe68c920e6/lib/guid.c#L36
	// GUID used by the shim.
	shimLockGUID = EFIGUID{0x605dab50, 0xe046, 0x4300, [8]byte{0xab, 0xb6, 0x3d, 0xd8, 0x10, 0xdd, 0x8b, 0x23}}
	// "SbatLevel" encoded as UCS-2.
	shimSbatVarName = []uint16{0x53, 0x62, 0x61, 0x74, 0x4c, 0x65, 0x76, 0x65, 0x6c}
)

// ErrSigMissingGUID is returned if an EFI_SIGNATURE_DATA structure was parsed
// successfully, however was missing the SignatureOwner GUID. This case is
// handled specially as a workaround for a bug relating to authority events.
var ErrSigMissingGUID = errors.New("signature data was missing owner GUID")

func (d EFIGUID) String() string {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, d.Data1); err != nil {
		return ""
	}
	if err := binary.Write(&buf, binary.BigEndian, d.Data2); err != nil {
		return ""
	}
	if err := binary.Write(&buf, binary.BigEndian, d.Data3); err != nil {
		return ""
	}
	if err := binary.Write(&buf, binary.BigEndian, d.Data4); err != nil {
		return ""
	}
	uuid := uuid.UUID{}
	if err := uuid.UnmarshalBinary(buf.Bytes()); err != nil {
		return ""
	}
	return uuid.String()
}

func (event baseEvent) RawEvent() Event {
	return event.Event
}

func parseStringData(b []byte) (string, error) {
	var buf []uint16
	for i := 0; i < len(b); i += 2 {
		if b[i+1] != 0x00 {
			buf = nil
			break
		}
		buf = append(buf, binary.LittleEndian.Uint16(b[i:]))
	}

	if buf != nil {
		return string(utf16.Decode(buf)), nil
	}

	if !utf8.Valid(b) {
		return "", errors.New("invalid UTF-8 string")
	}

	return string(b), nil
}

// ParseEfiVariableData parses byte array for UEFIVariableEvent
func ParseEfiVariableData(b []byte, parsedEvent *UEFIVariableEvent) error {
	var header UEFIVariableDataHeader
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.LittleEndian, &header)
	if err != nil {
		return err
	}
	if header.UnicodeNameLength < 1 {
		return fmt.Errorf("efi variable header unicode name length invalid")
	}
	if header.UnicodeNameLength > math.MaxUint16 {
		header.UnicodeNameLength = math.MaxUint16
	}
	unicodeName := make([]uint16, header.UnicodeNameLength)
	for i := 0; i < int(header.UnicodeNameLength); i++ {
		err = binary.Read(buf, binary.LittleEndian, &unicodeName[i])
		if err != nil {
			return err
		}
	}
	parsedEvent.VariableGUID = header.VariableName
	parsedEvent.VariableName = string(utf16.Decode(unicodeName))
	parsedEvent.VariableData = make([]byte, header.VariableDataLength)
	_, err = io.ReadFull(buf, parsedEvent.VariableData)
	return err
}

// Regular expression that matches UEFI "Boot####" variable names
var bootOption = regexp.MustCompile(`^Boot[0-9A-F]{4}$`)

// ParseEfiBootVariableData parses byte array for UefiBootVariableEvent
func ParseEfiBootVariableData(b []byte, parsedEvent *UEFIBootVariableEvent) error {
	var attributes uint32
	var dplength uint16
	var description []uint16
	var dpoffset int

	err := ParseEfiVariableData(b, &parsedEvent.UEFIVariableEvent)
	if err != nil {
		return err
	}
	// Return nil because of JSON representation
	if !bootOption.MatchString(parsedEvent.VariableName) {
		return nil
	}

	buf := bytes.NewBuffer(parsedEvent.VariableData)

	err = binary.Read(buf, binary.LittleEndian, &attributes)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &dplength)
	if err != nil {
		return err
	}

	// The device path starts after the null terminator in the UTF16 description
	for dpoffset = 6; dpoffset < len(parsedEvent.VariableData); dpoffset += 2 {
		var tmp uint16
		err = binary.Read(buf, binary.LittleEndian, &tmp)
		if err != nil {
			return err
		}

		description = append(description, tmp)
		if tmp == 0 {
			// Null terminator
			dpoffset += 2
			break
		}
	}

	parsedEvent.Description = string(utf16.Decode(description))

	// Verify that the structure is well formed
	if dpoffset+int(dplength) > len(parsedEvent.VariableData) {
		return fmt.Errorf("malformed boot variable")
	}

	parsedEvent.DevicePathRaw = make([]byte, dplength)
	err = binary.Read(buf, binary.LittleEndian, &parsedEvent.DevicePathRaw)
	if err != nil {
		return err
	}
	parsedEvent.DevicePath, err = efiDevicePath(parsedEvent.DevicePathRaw)
	if err != nil {
		return err
	}

	// Check whether there's any optional data
	optionaldatalen := len(parsedEvent.VariableData) - dpoffset - int(dplength)
	if optionaldatalen > 0 {
		parsedEvent.OptionalData = make([]byte, optionaldatalen)
		err = binary.Read(buf, binary.LittleEndian, &parsedEvent.OptionalData)
	}

	return err
}

// ParseEfiImageLoadEvent parses byte array for UEFIImageLoadEvent
func ParseEfiImageLoadEvent(b []byte, parsedEvent *UEFIImageLoadEvent) error {
	var devicePathLength uint64
	buf := bytes.NewBuffer(b)

	err := binary.Read(buf, binary.LittleEndian, &parsedEvent.ImageLocationInMemory)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &parsedEvent.ImageLengthInMemory)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &parsedEvent.ImageLinkTimeAddress)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &devicePathLength)
	if err != nil {
		return err
	}
	parsedEvent.DevicePathRaw = make([]byte, devicePathLength)
	err = binary.Read(buf, binary.LittleEndian, &parsedEvent.DevicePathRaw)
	if err != nil {
		return err
	}
	parsedEvent.DevicePath, err = efiDevicePath(parsedEvent.DevicePathRaw)

	return err
}

// ParseUefiGPTEvent parses byte array for UefiGPTEvent
func ParseUefiGPTEvent(b []byte, parsedEvent *UEFIGPTEvent) error {
	r := bytes.NewReader(b)
	err := binary.Read(r, binary.LittleEndian, &parsedEvent.UEFIPartitionHeader)
	if err != nil {
		return err
	}

	var numPartitions uint64
	err = binary.Read(r, binary.LittleEndian, &numPartitions)
	if err != nil {
		return err
	}

	if numPartitions*uint64(parsedEvent.UEFIPartitionHeader.SizeOfPartitionEntry) > uint64(r.Len()) {
		err = fmt.Errorf("(numPartitions * SizeOfPartitionEntry) > b.Len(), %d > %d", numPartitions*uint64(parsedEvent.UEFIPartitionHeader.SizeOfPartitionEntry), r.Len())
		return err
	}

	for i := uint64(0); i < numPartitions; i++ {
		r.Seek(int64(100+i*uint64(parsedEvent.UEFIPartitionHeader.SizeOfPartitionEntry)), io.SeekStart)
		var partition EFIPartition
		err = binary.Read(r, binary.LittleEndian, &partition)
		if err != nil {
			return err
		}
		parsedEvent.Partitions = append(parsedEvent.Partitions, partition)
	}

	return nil
}

// ParseUefiPlatformFirmwareBlobEvent parses byte array for UEFIPlatformFirmwareBlobEvent
func ParseUefiPlatformFirmwareBlobEvent(b []byte, parsedEvent *UEFIPlatformFirmwareBlobEvent) error {
	if len(b) != 16 {
		return fmt.Errorf("unexpected length for a platform firmware blob event: %d", len(b))
	}
	parsedEvent.BlobBase = binary.LittleEndian.Uint64(b)
	parsedEvent.BlobLength = binary.LittleEndian.Uint64(b[8:])
	return nil
}

// ParseEFIGUID parses from io.Reader for EFIGUID
func ParseEFIGUID(r io.Reader) (ret EFIGUID, err error) {
	if err = binary.Read(r, binary.LittleEndian, &ret.Data1); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &ret.Data2); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &ret.Data3); err != nil {
		return
	}
	_, err = r.Read(ret.Data4[:])
	return
}

// ParseUefiHandoffTableEvent parses byte array for UEFIHandoffTableEvent
func ParseUefiHandoffTableEvent(b []byte, parsedEvent *UEFIHandoffTableEvent) error {
	r := bytes.NewReader(b)
	var numTables uint64
	err := binary.Read(r, binary.LittleEndian, &numTables)
	if err != nil {
		return err
	}
	parsedEvent.Tables = make([]efiConfigurationTable, numTables)
	for i := uint64(0); i < numTables; i++ {
		if parsedEvent.Tables[i].VendorGUID, err = ParseEFIGUID(r); err != nil {
			err = fmt.Errorf("TableEntry[%d]: %v", i, err)
			return err
		}
		if err = binary.Read(r, binary.LittleEndian, &parsedEvent.Tables[i].VendorTable); err != nil {
			err = fmt.Errorf("TableEntry[%d]: %v", i, err)
			return err
		}
	}
	return err
}

// ParseOptionROMConfig parses byte array for OptionROMConfigEvent
func ParseOptionROMConfig(b []byte, parsedEvent *OptionROMConfigEvent) error {
	r := bytes.NewReader(b)
	var dummy uint16
	if err := binary.Read(r, binary.LittleEndian, &dummy); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &parsedEvent.PFA); err != nil {
		return err
	}
	_, err := io.ReadFull(r, parsedEvent.OptionROMStruct)
	return err
}

// ParseEvents parses events for TPMEvent array
func ParseEvents(events []Event) ([]TPMEvent, error) {
	var parsedEvents []TPMEvent

	for _, event := range events {
		buf := bytes.NewBuffer(event.Data)
		var err error
		switch event.Type {
		case PrebootCert: // 0x00
			var parsedEvent PrebootCertEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case PostCode: // 0x01
			var parsedEvent PostEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case NoAction: // 0x03
			var parsedEvent NoActionEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case Separator: // 0x04
			var parsedEvent SeparatorEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case Action: // 0x05
			var parsedEvent ActionEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EventTag: // 0x06
			var parsedEvent EventTagEvent
			var eventSize uint32
			parsedEvent.EventData = event.Data[8:]
			parsedEvent.Event = event
			if err := binary.Read(buf, binary.LittleEndian, &parsedEvent.EventID); err != nil {
				parsedEvent.Err = err
				parsedEvents = append(parsedEvents, parsedEvent)
				continue
			}
			if err := binary.Read(buf, binary.LittleEndian, &eventSize); err != nil {
				parsedEvent.Err = err
				parsedEvents = append(parsedEvents, parsedEvent)
				continue
			}
			switch parsedEvent.EventID {
			case OptionROMConfiguration:
				var OptionROMConfigEvent OptionROMConfigEvent
				OptionROMConfigEvent.Event = event
				err := ParseOptionROMConfig(parsedEvent.EventData, &OptionROMConfigEvent)
				if err != nil {
					parsedEvent.Err = err
					parsedEvents = append(parsedEvents, parsedEvent)
					continue
				}
				OptionROMConfigEvent.Event = event
				parsedEvents = append(parsedEvents, OptionROMConfigEvent)
			default:
				var MicrosoftEvent MicrosoftBootEvent
				// Pass the raw event data including the header
				err := parseMicrosoftEvent(parsedEvent.Data, &MicrosoftEvent)
				if err != nil {
					parsedEvent.Err = err
					parsedEvents = append(parsedEvents, parsedEvent)
					continue
				}
				MicrosoftEvent.Event = event
				parsedEvents = append(parsedEvents, MicrosoftEvent)
			}
		case SCRTMContents: // 0x07
			var parsedEvent CRTMContentEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case SCRTMVersion: // 0x08
			var parsedEvent CRTMEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case CpuMicrocode: // 0x09
			var parsedEvent MicrocodeEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case PlatformConfigFlags: // 0x0a
			var parsedEvent PlatformConfigFlagsEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case TableOfDevices: // 0x0b
			var parsedEvent TableOfDevicesEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case CompactHash: // 0x0c
			var parsedEvent CompactHashEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case Ipl: // 0x0d
			var parsedEvent IPLEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case IplPartitionData: // 0x0e
			var parsedEvent IPLPartitionEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case NonhostCode: // 0x0f
			var parsedEvent NonHostCodeEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case NonhostConfig: // 0x10
			var parsedEvent NonHostConfigEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case NonhostInfo: // 0x11
			var parsedEvent NonHostInfoEvent
			parsedEvent.Event = event
			parsedEvents = append(parsedEvents, parsedEvent)
		case OmitBootDeviceEvents: // 0x12
			var parsedEvent OmitBootDeviceEventsEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIVariableDriverConfig: // 0x80000001
			var parsedEvent UEFIVariableDriverConfigEvent
			parsedEvent.Event = event
			err = ParseEfiVariableData(event.Data, &parsedEvent.UEFIVariableEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIVariableBoot: // 0x80000002
			var parsedEvent UEFIBootVariableEvent
			parsedEvent.Event = event
			err = ParseEfiBootVariableData(event.Data, &parsedEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIBootServicesApplication: // 0x80000003
			var parsedEvent UEFIBootServicesApplicationEvent
			parsedEvent.Event = event
			err = ParseEfiImageLoadEvent(event.Data, &parsedEvent.UEFIImageLoadEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIBootServicesDriver: // 0x80000004
			var parsedEvent UEFIBootServicesDriverEvent
			parsedEvent.Event = event
			err = ParseEfiImageLoadEvent(event.Data, &parsedEvent.UEFIImageLoadEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIAction: // 0x80000005
			var parsedEvent UEFIActionEvent
			parsedEvent.Event = event
			parsedEvent.Message, err = parseStringData(event.Data)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIRuntimeServicesDriver: // 0x80000006
			var parsedEvent UEFIRuntimeServicesDriverEvent
			parsedEvent.Event = event
			err = ParseEfiImageLoadEvent(event.Data, &parsedEvent.UEFIImageLoadEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIGPTEvent: // 0x80000007
			var parsedEvent UEFIGPTEvent
			parsedEvent.Event = event
			err = ParseUefiGPTEvent(event.Data, &parsedEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIPlatformFirmwareBlob: // 0x80000008
			var parsedEvent UEFIPlatformFirmwareBlobEvent
			parsedEvent.Event = event
			err = ParseUefiPlatformFirmwareBlobEvent(event.Data, &parsedEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIHandoffTables:
			var parsedEvent UEFIHandoffTableEvent
			parsedEvent.Event = event
			err = ParseUefiHandoffTableEvent(event.Data, &parsedEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		case EFIVariableAuthority:
			var parsedEvent UEFIVariableAuthorityEvent
			parsedEvent.Event = event
			err = ParseEfiVariableData(event.Data, &parsedEvent.UEFIVariableEvent)
			if err != nil {
				parsedEvent.Err = err
			}
			parsedEvents = append(parsedEvents, parsedEvent)
		}
	}

	return parsedEvents, nil
}

// ParseEfiSignature parses byte array for EFI signatures in x509 format
func ParseEfiSignature(b []byte) ([]x509.Certificate, error) {
	certificates := []x509.Certificate{}

	if len(b) < 16 {
		return nil, fmt.Errorf("invalid signature: buffer smaller than header (%d < %d)", len(b), 16)
	}

	buf := bytes.NewReader(b)
	signature := EFISignatureData{}
	signature.SignatureData = make([]byte, len(b)-16)

	if err := binary.Read(buf, binary.LittleEndian, &signature.SignatureOwner); err != nil {
		return certificates, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &signature.SignatureData); err != nil {
		return certificates, err
	}

	cert, err := x509.ParseCertificate(signature.SignatureData)
	if err == nil {
		certificates = append(certificates, *cert)
	} else {
		// A bug in shim may cause an event to be missing the SignatureOwner GUID.
		// We handle this, but signal back to the caller using ErrSigMissingGUID.
		if _, isStructuralErr := err.(asn1.StructuralError); isStructuralErr {
			var err2 error
			cert, err2 = x509.ParseCertificate(b)
			if err2 == nil {
				certificates = append(certificates, *cert)
				err = ErrSigMissingGUID
			}
		}
	}
	return certificates, err
}

// ParseEfiSignatureList parses a EFI_SIGNATURE_LIST structure.
// The structure and related GUIDs are defined at:
// https://uefi.org/sites/default/files/resources/UEFI_Spec_2_8_final.pdf#page=1790
func ParseEfiSignatureList(b []byte) ([]x509.Certificate, [][]byte, error) {
	if len(b) < 28 {
		// Being passed an empty signature list here appears to be valid
		return nil, nil, nil
	}
	signatures := efiSignatureList{}
	buf := bytes.NewReader(b)
	certificates := []x509.Certificate{}
	hashes := [][]byte{}

	for buf.Len() > 0 {
		err := binary.Read(buf, binary.LittleEndian, &signatures.Header)
		if err != nil {
			return nil, nil, err
		}

		if signatures.Header.SignatureHeaderSize > EFIMaxDataLen {
			return nil, nil, fmt.Errorf("signature header too large: %d > %d", signatures.Header.SignatureHeaderSize, EFIMaxDataLen)
		}
		if signatures.Header.SignatureListSize > EFIMaxDataLen {
			return nil, nil, fmt.Errorf("signature list too large: %d > %d", signatures.Header.SignatureListSize, EFIMaxDataLen)
		}

		signatureType := signatures.Header.SignatureType
		switch signatureType {
		case certX509SigGUID: // X509 certificate
			for sigOffset := 0; uint32(sigOffset) < signatures.Header.SignatureListSize-28; {
				signature := efiSignatureData{}
				signature.SignatureData = make([]byte, signatures.Header.SignatureSize-16)
				err := binary.Read(buf, binary.LittleEndian, &signature.SignatureOwner)
				if err != nil {
					return nil, nil, err
				}
				err = binary.Read(buf, binary.LittleEndian, &signature.SignatureData)
				if err != nil {
					return nil, nil, err
				}
				cert, err := x509.ParseCertificate(signature.SignatureData)
				if err != nil {
					return nil, nil, err
				}
				sigOffset += int(signatures.Header.SignatureSize)
				certificates = append(certificates, *cert)
			}
		case hashSHA256SigGUID: // SHA256
			for sigOffset := 0; uint32(sigOffset) < signatures.Header.SignatureListSize-28; {
				signature := efiSignatureData{}
				signature.SignatureData = make([]byte, signatures.Header.SignatureSize-16)
				err := binary.Read(buf, binary.LittleEndian, &signature.SignatureOwner)
				if err != nil {
					return nil, nil, err
				}
				err = binary.Read(buf, binary.LittleEndian, &signature.SignatureData)
				if err != nil {
					return nil, nil, err
				}
				hashes = append(hashes, signature.SignatureData)
				sigOffset += int(signatures.Header.SignatureSize)
			}
		case keyRSA2048SigGUID:
			err = errors.New("unhandled RSA2048 key")
		case certRSA2048SHA256SigGUID:
			err = errors.New("unhandled RSA2048-SHA256 key")
		case hashSHA1SigGUID:
			err = errors.New("unhandled SHA1 hash")
		case certRSA2048SHA1SigGUID:
			err = errors.New("unhandled RSA2048-SHA1 key")
		case hashSHA224SigGUID:
			err = errors.New("unhandled SHA224 hash")
		case hashSHA384SigGUID:
			err = errors.New("unhandled SHA384 hash")
		case hashSHA512SigGUID:
			err = errors.New("unhandled SHA512 hash")
		case certHashSHA256SigGUID:
			err = errors.New("unhandled X509-SHA256 hash metadata")
		case certHashSHA384SigGUID:
			err = errors.New("unhandled X509-SHA384 hash metadata")
		case certHashSHA512SigGUID:
			err = errors.New("unhandled X509-SHA512 hash metadata")
		default:
			err = fmt.Errorf("unhandled signature type %s", signatureType)
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return certificates, hashes, nil
}

// UEFIVariableAuthority describes the contents of a UEFI variable authority
// event.
type UEFIVariableAuthority struct {
	Certs []x509.Certificate
}

// ParseUEFIVariableAuthority parses the data section of an event structured as
// a UEFI variable authority.
//
// https://uefi.org/sites/default/files/resources/UEFI_Spec_2_8_final.pdf#page=1789
func ParseUEFIVariableAuthority(v UEFIVariableData) (UEFIVariableAuthority, error) {
	// Skip parsing new SBAT section logged by shim.
	// See https://github.com/rhboot/shim/blob/main/SBAT.md for more.
	if v.Header.VariableName == shimLockGUID && unicodeNameEquals(v, shimSbatVarName) {
		//https://github.com/rhboot/shim/blob/20e4d9486fcae54ee44d2323ae342ffe68c920e6/include/sbat.h#L9-L12
		return UEFIVariableAuthority{}, nil
	}
	certs, err := ParseEfiSignature(v.VariableData)
	return UEFIVariableAuthority{Certs: certs}, err
}

func unicodeNameEquals(v UEFIVariableData, comp []uint16) bool {
	if len(v.UnicodeName) != len(comp) {
		return false
	}
	for i, v := range v.UnicodeName {
		if v != comp[i] {
			return false
		}
	}
	return true
}

// UntrustedParseEventType returns the event type indicated by
// the provided value.
func UntrustedParseEventType(et uint32) (EventType, error) {
	// "The value associated with a UEFI specific platform event type MUST be in
	// the range between 0x80000000 and 0x800000FF, inclusive."
	if (et < 0x80000000 && et > 0x800000FF) || (et < 0x0 && et > 0x12) {
		return EventType(0), fmt.Errorf("event type not between [0x0, 0x12] or [0x80000000, 0x800000FF]: got %#x", et)
	}
	if _, ok := eventTypeNames[EventType(et)]; !ok {
		return EventType(0), fmt.Errorf("unknown event type %#x", et)
	}
	return EventType(et), nil
}

// ParseTaggedEventData parses a TCG_PCClientTaggedEventStruct structure.
func ParseTaggedEventData(d []byte) (*TaggedEventData, error) {
	var (
		r      = bytes.NewReader(d)
		header struct {
			ID      uint32
			DataLen uint32
		}
	)
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	if int(header.DataLen) > len(d) {
		return nil, fmt.Errorf("tagged event len (%d bytes) larger than data length (%d bytes)", header.DataLen, len(d))
	}

	out := TaggedEventData{
		ID:   header.ID,
		Data: make([]byte, header.DataLen),
	}
	return &out, binary.Read(r, binary.LittleEndian, &out.Data)
}

// VarName on *UEFIVariableData instance decodes variable name
func (v *UEFIVariableData) VarName() string {
	return string(utf16.Decode(v.UnicodeName))
}

// SignatureData on *UEFIVariableData instance decodes signature data
func (v *UEFIVariableData) SignatureData() (certs []x509.Certificate, hashes [][]byte, err error) {
	return ParseEfiSignatureList(v.VariableData)
}

// ParseUEFIVariableData parses the data section of an event structured as
// a UEFI variable.
//
// https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClient_Specific_Platform_Profile_for_TPM_2p0_1p04_PUBLIC.pdf#page=100
func ParseUEFIVariableData(r io.Reader) (ret UEFIVariableData, err error) {
	err = binary.Read(r, binary.LittleEndian, &ret.Header)
	if err != nil {
		return
	}
	if ret.Header.UnicodeNameLength > EFIMaxNameLen {
		return UEFIVariableData{}, fmt.Errorf("unicode name too long: %d > %d", ret.Header.UnicodeNameLength, EFIMaxNameLen)
	}
	ret.UnicodeName = make([]uint16, ret.Header.UnicodeNameLength)
	for i := 0; uint64(i) < ret.Header.UnicodeNameLength; i++ {
		err = binary.Read(r, binary.LittleEndian, &ret.UnicodeName[i])
		if err != nil {
			return
		}
	}
	if ret.Header.VariableDataLength > EFIMaxDataLen {
		return UEFIVariableData{}, fmt.Errorf("variable data too long: %d > %d", ret.Header.VariableDataLength, EFIMaxDataLen)
	}
	ret.VariableData = make([]byte, ret.Header.VariableDataLength)
	_, err = io.ReadFull(r, ret.VariableData)
	return
}
