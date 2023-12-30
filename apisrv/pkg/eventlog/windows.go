package eventlog

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"
)

var windowsEventNames = map[windowsEvent]string{
	trustBoundary:                   "TrustBoundary",
	elamAggregation:                 "ELAMAggregation",
	loadedModuleAggregation:         "LoadedModuleAggregation",
	trustpointAggregation:           "TrustpointAggregation",
	ksrAggregation:                  "KSRAggregation",
	ksrSignedMeasurementAggregation: "KSRSignedMeasurementAggregation",
	information:                     "Information",
	bootCounter:                     "BootCounter",
	transferControl:                 "TransferControl",
	applicationReturn:               "ApplicationReturn",
	bitlockerUnlock:                 "BitlockerUnlock",
	eventCounter:                    "EventCounter",
	counterID:                       "CounterID",
	morBitNotCancelable:             "MORBitNotCancelable",
	applicationSVN:                  "ApplicationSVN",
	svnChainStatus:                  "SVNChainStatus",
	morBitAPIStatus:                 "MORBitAPIStatus",
	bootDebugging:                   "BootDebugging",
	bootRevocationList:              "BootRevocationList",
	osKernelDebug:                   "OSKernelDebug",
	codeIntegrity:                   "CodeIntegrity",
	testSigning:                     "TestSigning",
	dataExecutionPrevention:         "DataExecutionPrevention",
	safeMode:                        "SafeMode",
	winPE:                           "WinPE",
	physicalAddressExtension:        "PhysicalAddressExtension",
	osDevice:                        "OSDevice",
	systemRoot:                      "SystemRoot",
	hypervisorLaunchType:            "HypervisorLaunchType",
	hypervisorPath:                  "HypervisorPath",
	hypervisorIOMMUPolicy:           "HypervisorIOMMUPolicy",
	hypervisorDebug:                 "HypervisorDebug",
	driverLoadPolicy:                "DriverLoadPolicy",
	siPolicy:                        "SIPolicy",
	hypervisorMMIONXPolicy:          "HypervisorMMIONXPolicy",
	hypervisorMSRFilterPolicy:       "HypervisorMSRFilterPolicy",
	vsmLaunchType:                   "VSMLaunchType",
	osRevocationList:                "OSRevocationList",
	vsmIDKInfo:                      "VSMIDKInfo",
	flightSigning:                   "FlightSigning",
	pagefileEncryptionEnabled:       "PagefileEncryptionEnabled",
	vsmIDKSInfo:                     "VSMIDKSInfo",
	hibernationDisabled:             "HibernationDisabled",
	dumpsDisabled:                   "DumpsDisabled",
	dumpEncryptionEnabled:           "DumpEncryptionEnabled",
	dumpEncryptionKeyDigest:         "DumpEncryptionKeyDigest",
	lsaISOConfig:                    "LSAISOConfig",
	noAuthority:                     "NoAuthority",
	authorityPubKey:                 "AuthorityPubKey",
	filePath:                        "FilePath",
	imageSize:                       "ImageSize",
	hashAlgorithmID:                 "HashAlgorithmID",
	authenticodeHash:                "AuthenticodeHash",
	authorityIssuer:                 "AuthorityIssuer",
	authoritySerial:                 "AuthoritySerial",
	imageBase:                       "ImageBase",
	authorityPublisher:              "AuthorityPublisher",
	authoritySHA1Thumbprint:         "AuthoritySHA1Thumbprint",
	imageValidated:                  "ImageValidated",
	moduleSVN:                       "ModuleSVN",
	quote:                           "Quote",
	quoteSignature:                  "QuoteSignature",
	aikID:                           "AIKID",
	aikPubDigest:                    "AIKPubDigest",
	elamKeyname:                     "ELAMKeyname",
	elamConfiguration:               "ELAMConfiguration",
	elamPolicy:                      "ELAMPolicy",
	elamMeasured:                    "ELAMMeasured",
	vbsVSMRequired:                  "VBSVSMRequired",
	vbsSecurebootRequired:           "VBSSecurebootRequired",
	vbsIOMMURequired:                "VBSIOMMURequired",
	vbsNXRequired:                   "VBSNXRequired",
	vbsMSRFilteringRequired:         "VBSMSRFilteringRequired",
	vbsMandatoryEnforcement:         "VBSMandatoryEnforcement",
	vbsHVCIPolicy:                   "VBSHVCIPolicy",
	vbsMicrosoftBootChainRequired:   "VBSMicrosoftBootChainRequired",
	ksrSignature:                    "KSRSignature",
}

// ParseWinEvents parses a series of events to extract information about
// the bringup of Microsoft Windows. This information is not trustworthy
// unless the integrity of platform & bootloader events has already been
// established.
func ParseWinEvents(events []Event) (*WinEvents, error) {
	var (
		out = WinEvents{
			LoadedModules:   map[string]WinModuleLoad{},
			ELAM:            map[string]WinELAM{},
			TrustPointQuote: map[string]WinWBCLQuote{},
		}
		seenSeparator struct {
			PCR12 bool
			PCR13 bool
		}
	)

	for _, e := range events {
		if e.Index != 12 && e.Index != 13 && e.Index != 0xFFFFFFFF {
			continue
		}

		et, err := UntrustedParseEventType(uint32(e.Type))
		if err != nil {
			return nil, fmt.Errorf("unrecognised event type: %v", err)
		}

		digestVerify := e.DigestEquals(e.Data)

		switch e.Index {
		case 12: // 'early boot' events
			switch et {
			case EventTag:
				if seenSeparator.PCR12 {
					continue
				}
				s, err := ParseTaggedEventData(e.Data)
				if err != nil {
					return nil, fmt.Errorf("invalid tagged event structure at event %d: %w", e.Sequence, err)
				}
				if digestVerify != nil {
					return nil, fmt.Errorf("invalid digest for tagged event %d PCR(12): %w", e.Sequence, digestVerify)
				}
				if err := out.readWinEventBlock(s, e.Index); err != nil {
					return nil, fmt.Errorf("invalid SIPA events in event %d: %w", e.Sequence, err)
				}
			case Separator:
				if seenSeparator.PCR12 {
					return nil, fmt.Errorf("duplicate WBCL separator at event %d", e.Sequence)
				}
				seenSeparator.PCR12 = true
				if !bytes.Equal(e.Data, []byte("WBCL")) {
					return nil, fmt.Errorf("invalid WBCL separator data at event %d: %v", e.Sequence, e.Data)
				}
				if digestVerify != nil {
					return nil, fmt.Errorf("invalid separator digest at event %d: %v", e.Sequence, digestVerify)
				}

			default:
				return nil, fmt.Errorf("unexpected (PCR12) event type: %v", et)
			}
		case 13: // Post 'early boot' events
			switch et {
			case EventTag:
				if seenSeparator.PCR13 {
					continue
				}
				s, err := ParseTaggedEventData(e.Data)
				if err != nil {
					return nil, fmt.Errorf("invalid tagged event structure at event %d: %w", e.Sequence, err)
				}
				if digestVerify != nil {
					return nil, fmt.Errorf("invalid digest for tagged event %d PCR(13): %w", e.Sequence, digestVerify)
				}
				if err := out.readWinEventBlock(s, e.Index); err != nil {
					return nil, fmt.Errorf("invalid SIPA events in event %d: %w", e.Sequence, err)
				}
			case Separator:
				if seenSeparator.PCR13 {
					return nil, fmt.Errorf("duplicate WBCL separator at event %d", e.Sequence)
				}
				seenSeparator.PCR13 = true
				if !bytes.Equal(e.Data, []byte("WBCL")) {
					return nil, fmt.Errorf("invalid WBCL separator data at event %d: %v", e.Sequence, e.Data)
				}
				if digestVerify != nil {
					return nil, fmt.Errorf("invalid separator digest at event %d: %v", e.Sequence, digestVerify)
				}

			default:
				return nil, fmt.Errorf("unexpected (PCR13) event type: %v", et)
			}
		case 0xFFFFFFFF: // TPM12_PCR_TRUSTPOINT
			switch et {
			case NoAction:
				s, err := ParseTaggedEventData(e.Data)
				if err != nil {
					return nil, fmt.Errorf("invalid tagged event structure at event %d: %w", e.Sequence, err)
				}
				if digestVerify != nil {
					return nil, fmt.Errorf("invalid digest for tagged event %d PCR(-1): %w", e.Sequence, digestVerify)
				}
				if err := out.readOuterTrustPointAggregation(s); err != nil {
					return nil, fmt.Errorf("invalid SIPA events in event %d: %w", e.Sequence, err)
				}

			default:
				return nil, fmt.Errorf("unexpected (PCR_TRUSTPOINT) event type: %v", et)
			}
		}
	}
	return &out, nil
}

// ErrUnknownSIPAEvent is returned by parseSIPAEvent if the event type is
// not handled. Unlike other events in the TCG log, it is safe to skip
// unhandled SIPA events, as they are embedded within EventTag structures,
// and these structures should match the event digest.
var ErrUnknownSIPAEvent = errors.New("unknown event")

func (w *WinEvents) readBooleanInt64Event(header microsoftEventHeader, r *bytes.Reader) error {
	if header.Size != 8 {
		return fmt.Errorf("payload was %d bytes, want 8", header.Size)
	}
	var num uint64
	if err := binary.Read(r, binary.LittleEndian, &num); err != nil {
		return fmt.Errorf("reading u64: %w", err)
	}
	isSet := num != 0

	switch header.Type {
	// Boolean signals that latch off if the are ever false (ie: attributes
	// that represent a stronger security state when set).
	case dataExecutionPrevention:
		if isSet && w.DEPEnabled == TernaryUnknown {
			w.DEPEnabled = TernaryTrue
		} else if !isSet {
			w.DEPEnabled = TernaryFalse
		}
	}
	return nil
}

func (w *WinEvents) readBooleanByteEvent(header microsoftEventHeader, r *bytes.Reader) error {
	if header.Size != 1 {
		return fmt.Errorf("payload was %d bytes, want 1", header.Size)
	}
	var b byte
	if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
		return fmt.Errorf("reading byte: %w", err)
	}
	isSet := b != 0

	switch header.Type {
	// Boolean signals that latch on if they are ever true (ie: attributes
	// that represent a weaker security state when set).
	case osKernelDebug:
		w.KernelDebugEnabled = w.KernelDebugEnabled || isSet
	case bootDebugging:
		w.BootDebuggingEnabled = w.BootDebuggingEnabled || isSet
	case testSigning:
		w.TestSigningEnabled = w.TestSigningEnabled || isSet

	// Boolean signals that latch off if the are ever false (ie: attributes
	// that represent a stronger security state when set).
	case codeIntegrity:
		if isSet && w.CodeIntegrityEnabled == TernaryUnknown {
			w.CodeIntegrityEnabled = TernaryTrue
		} else if !isSet {
			w.CodeIntegrityEnabled = TernaryFalse
		}
	}
	return nil
}

func (w *WinEvents) readUint32(header microsoftEventHeader, r io.Reader) (uint32, error) {
	if header.Size != 4 {
		return 0, fmt.Errorf("integer size not uint32 (%d bytes)", header.Size)
	}

	data := make([]uint8, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return 0, fmt.Errorf("reading u32: %w", err)
	}
	i := binary.LittleEndian.Uint32(data)

	return i, nil
}

func (w *WinEvents) readUint64(header microsoftEventHeader, r io.Reader) (uint64, error) {
	if header.Size != 8 {
		return 0, fmt.Errorf("integer size not uint64 (%d bytes)", header.Size)
	}

	data := make([]uint8, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return 0, fmt.Errorf("reading u64: %w", err)
	}
	i := binary.LittleEndian.Uint64(data)

	return i, nil
}

func (w *WinEvents) readBootCounter(header microsoftEventHeader, r *bytes.Reader) error {
	i, err := w.readUint64(header, r)
	if err != nil {
		return fmt.Errorf("boot counter: %v", err)
	}

	if w.BootCount > 0 && w.BootCount != i {
		return fmt.Errorf("conflicting values for boot counter: %d != %d", i, w.BootCount)
	}
	w.BootCount = i
	return nil
}

func (w *WinEvents) readEventCounter(header microsoftEventHeader, r *bytes.Reader) error {
	i, err := w.readUint64(header, r)
	if err != nil {
		return fmt.Errorf("event counter: %v", err)
	}

	if w.EventCount > 0 && w.EventCount != i {
		return fmt.Errorf("conflicting event for boot counter: %d != %d", i, w.EventCount)
	}
	w.EventCount = i
	return nil
}

func (w *WinEvents) readEventCounterId(header microsoftEventHeader, r *bytes.Reader) error {
	i, err := w.readUint64(header, r)
	if err != nil {
		return fmt.Errorf("event counter id: %v", err)
	}

	if w.EventCounterId > 0 && w.EventCounterId != i {
		return fmt.Errorf("conflicting values for event counter id: %d != %d", i, w.EventCounterId)
	}
	w.EventCounterId = uint64(i)
	return nil
}

func (w *WinEvents) readTransferControl(header microsoftEventHeader, r *bytes.Reader) error {
	i, err := w.readUint32(header, r)
	if err != nil {
		return fmt.Errorf("transfer control: %v", err)
	}

	// A transferControl event with a value of 1 indicates that bootmngr
	// launched WinLoad. A different (unknown) value is set if WinResume
	// is launched.
	w.ColdBoot = i == 0x1
	return nil
}

func (w *WinEvents) readBitlockerUnlock(header microsoftEventHeader, r *bytes.Reader, pcr int) error {
	if header.Size > 8 {
		return fmt.Errorf("bitlocker data too large (%d bytes)", header.Size)
	}
	data := make([]uint8, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return fmt.Errorf("reading u%d: %w", header.Size<<8, err)
	}
	i, n := binary.Uvarint(data)
	if n <= 0 {
		return fmt.Errorf("reading u%d: invalid varint", header.Size<<8)
	}

	if pcr == 13 {
		// The bitlocker status is duplicated across both PCRs. As such,
		// we prefer the earlier one, and bail here to prevent duplicate
		// records.
		return nil
	}

	w.BitlockerUnlocks = append(w.BitlockerUnlocks, BitlockerStatus(i))
	return nil
}

func (w *WinEvents) parseImageValidated(header microsoftEventHeader, r io.Reader) (bool, error) {
	if header.Size != 1 {
		return false, fmt.Errorf("payload was %d bytes, want 1", header.Size)
	}
	var num byte
	if err := binary.Read(r, binary.LittleEndian, &num); err != nil {
		return false, fmt.Errorf("reading u8: %w", err)
	}
	return num == 1, nil
}

func (w *WinEvents) parseHashAlgID(header microsoftEventHeader, r io.Reader) (WinCSPAlg, error) {
	i, err := w.readUint32(header, r)
	if err != nil {
		return 0, fmt.Errorf("hash algorithm ID: %v", err)
	}

	switch alg := WinCSPAlg(i & 0xff); alg {
	case WinAlgMD4, WinAlgMD5, WinAlgSHA1, WinAlgSHA256, WinAlgSHA384, WinAlgSHA512:
		return alg, nil
	default:
		return 0, fmt.Errorf("unknown algorithm ID: %x", i)
	}
}

func (w *WinEvents) parseAuthoritySerial(header microsoftEventHeader, r io.Reader) ([]byte, error) {
	if header.Size > 128 {
		return nil, fmt.Errorf("authority serial is too long (%d bytes)", header.Size)
	}
	data := make([]byte, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, fmt.Errorf("reading bytes: %w", err)
	}
	return data, nil
}

func (w *WinEvents) parseAuthoritySHA1(header microsoftEventHeader, r io.Reader) ([]byte, error) {
	if header.Size > 20 {
		return nil, fmt.Errorf("authority thumbprint is too long (%d bytes)", header.Size)
	}
	data := make([]byte, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, fmt.Errorf("reading bytes: %w", err)
	}
	return data, nil
}

func (w *WinEvents) parseImageBase(header microsoftEventHeader, r io.Reader) (uint64, error) {
	if header.Size != 8 {
		return 0, fmt.Errorf("payload was %d bytes, want 8", header.Size)
	}
	var num uint64
	if err := binary.Read(r, binary.LittleEndian, &num); err != nil {
		return 0, fmt.Errorf("reading u64: %w", err)
	}
	return num, nil
}

func (w *WinEvents) parseAuthenticodeHash(header microsoftEventHeader, r io.Reader) ([]byte, error) {
	if header.Size > 32 {
		return nil, fmt.Errorf("authenticode hash data exceeds the size of any valid hash (%d bytes)", header.Size)
	}
	data := make([]byte, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, fmt.Errorf("reading bytes: %w", err)
	}
	return data, nil
}

func (w *WinEvents) readLoadedModuleAggregation(rdr *bytes.Reader, header microsoftEventHeader) error {
	var (
		r                   = &io.LimitedReader{R: rdr, N: int64(header.Size)}
		codeHash            []byte
		imgBase, imgSize    uint64
		fPath               string
		algID               WinCSPAlg
		imgValidated        bool
		aIssuer, aPublisher string
		aSerial, aSHA1      []byte
	)

	for r.N > 0 {
		var h microsoftEventHeader
		if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
			return fmt.Errorf("parsing LMA sub-event: %v", err)
		}
		if int64(h.Size) > r.N {
			return fmt.Errorf("LMA sub-event is larger than available data: %d > %d", h.Size, r.N)
		}

		var err error
		switch h.Type {
		case imageBase:
			if imgBase != 0 {
				return errors.New("duplicate image base data in LMA event")
			}
			if imgBase, err = w.parseImageBase(h, r); err != nil {
				return err
			}
		case authenticodeHash:
			if codeHash != nil {
				return errors.New("duplicate authenticode hash structure in LMA event")
			}
			if codeHash, err = w.parseAuthenticodeHash(h, r); err != nil {
				return err
			}
		case filePath:
			if fPath != "" {
				return errors.New("duplicate file path in LMA event")
			}
			if fPath, err = w.parseUTF16(h, r); err != nil {
				return err
			}
		case imageSize:
			if imgSize != 0 {
				return errors.New("duplicate image size in LMA event")
			}
			if imgSize, err = w.readUint64(h, r); err != nil {
				return err
			}
		case hashAlgorithmID:
			if algID != 0 {
				return errors.New("duplicate hash algorithm ID in LMA event")
			}
			if algID, err = w.parseHashAlgID(h, r); err != nil {
				return err
			}
		case imageValidated:
			if imgValidated {
				return errors.New("duplicate image validated field in LMA event")
			}
			if imgValidated, err = w.parseImageValidated(h, r); err != nil {
				return err
			}
		case authorityIssuer:
			if aIssuer != "" {
				return errors.New("duplicate authority issuer in LMA event")
			}
			if aIssuer, err = w.parseUTF16(h, r); err != nil {
				return err
			}
		case authorityPublisher:
			if aPublisher != "" {
				return errors.New("duplicate authority publisher in LMA event")
			}
			if aPublisher, err = w.parseUTF16(h, r); err != nil {
				return err
			}
		case authoritySerial:
			if aSerial != nil {
				return errors.New("duplicate authority serial in LMA event")
			}
			if aSerial, err = w.parseAuthoritySerial(h, r); err != nil {
				return err
			}
		case authoritySHA1Thumbprint:
			if aSHA1 != nil {
				return errors.New("duplicate authority SHA1 thumbprint in LMA event")
			}
			if aSHA1, err = w.parseAuthoritySHA1(h, r); err != nil {
				return err
			}
		case moduleSVN:
			// Ignore - consume value.
			b := make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown event in LMA aggregation: %v", h.Type)
		}
	}

	var iBase []uint64
	if imgBase != 0 {
		iBase = []uint64{imgBase}
	}

	l := WinModuleLoad{
		FilePath:           fPath,
		AuthenticodeHash:   codeHash,
		ImageBase:          iBase,
		ImageSize:          imgSize,
		ImageValidated:     imgValidated,
		HashAlgorithm:      algID,
		AuthorityIssuer:    aIssuer,
		AuthorityPublisher: aPublisher,
		AuthoritySerial:    aSerial,
		AuthoritySHA1:      aSHA1,
	}
	hashHex := hex.EncodeToString(l.AuthenticodeHash)
	l.ImageBase = append(l.ImageBase, w.LoadedModules[hashHex].ImageBase...)
	w.LoadedModules[hashHex] = l
	return nil
}

// parseUTF16 decodes data representing a UTF16 string. It is assumed the
// caller has validated that the data size is within allowable bounds.
func (w *WinEvents) parseUTF16(header microsoftEventHeader, r io.Reader) (string, error) {
	data := make([]uint16, header.Size/2)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(utf16.Decode(data)), "\x00"), nil
}

func (w *WinEvents) readELAMAggregation(rdr *bytes.Reader, header microsoftEventHeader) error {
	var (
		r          = &io.LimitedReader{R: rdr, N: int64(header.Size)}
		driverName string
		measured   []byte
		policy     []byte
		config     []byte
	)

	for r.N > 0 {
		var h microsoftEventHeader
		if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
			return fmt.Errorf("parsing ELAM aggregation sub-event: %v", err)
		}
		if int64(h.Size) > r.N {
			return fmt.Errorf("ELAM aggregation sub-event is larger than available data: %d > %d", h.Size, r.N)
		}

		var err error
		switch h.Type {
		case elamKeyname:
			if driverName != "" {
				return errors.New("duplicate driver name in ELAM aggregation event")
			}
			if driverName, err = w.parseUTF16(h, r); err != nil {
				return fmt.Errorf("parsing ELAM driver name: %v", err)
			}
		case elamMeasured:
			if measured != nil {
				return errors.New("duplicate measured data in ELAM aggregation event")
			}
			measured = make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &measured); err != nil {
				return fmt.Errorf("reading ELAM measured value: %v", err)
			}
		case elamPolicy:
			if policy != nil {
				return errors.New("duplicate policy data in ELAM aggregation event")
			}
			policy = make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &policy); err != nil {
				return fmt.Errorf("reading ELAM policy value: %v", err)
			}
		case elamConfiguration:
			if config != nil {
				return errors.New("duplicate config data in ELAM aggregation event")
			}
			config = make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &config); err != nil {
				return fmt.Errorf("reading ELAM config value: %v", err)
			}
		default:
			return fmt.Errorf("unknown event in LMA aggregation: %v", h.Type)
		}
	}

	if driverName == "" {
		return errors.New("ELAM driver name not specified")
	}
	w.ELAM[driverName] = WinELAM{
		Measured: measured,
		Config:   config,
		Policy:   policy,
	}
	return nil
}

func (w *WinEvents) readInnerTrustPointAggregation(rdr *bytes.Reader, dataLen uint32) error {
	var (
		r          = &io.LimitedReader{R: rdr, N: int64(dataLen)}
		aikName    string
		aikPubHash []byte
		quoteBlob  []byte
		quoteSig   []byte
	)

	for r.N > 0 {
		var h microsoftEventHeader
		if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
			return fmt.Errorf("parsing TrustPoint aggregation sub-event: %v", err)
		}
		if int64(h.Size) > r.N {
			return fmt.Errorf("TrustPoint aggregation sub-event is larger than available data: %d > %d", h.Size, r.N)
		}

		var err error
		switch h.Type {
		case aikID:
			if aikName != "" {
				return errors.New("duplicate AIK id in inner TrustPoint aggregation event")
			}
			if aikName, err = w.parseUTF16(h, r); err != nil {
				return fmt.Errorf("parsing AIK id: %v", err)
			}
		case aikPubDigest:
			if aikPubHash != nil {
				return errors.New("duplicate AIK public digest in inner TrustPoint aggregation event")
			}
			aikPubHash = make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &aikPubHash); err != nil {
				return fmt.Errorf("reading AIK public digest: %v", err)
			}
		case quote:
			if quoteBlob != nil {
				return errors.New("duplicate quote blob in inner TrustPoint aggregation event")
			}
			quoteBlob = make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &quoteBlob); err != nil {
				return fmt.Errorf("reading quote blob: %v", err)
			}
		case quoteSignature:
			if quoteSig != nil {
				return errors.New("duplicate config data in ELAM aggregation event")
			}
			quoteSig = make([]byte, h.Size)
			if err := binary.Read(r, binary.LittleEndian, &quoteSig); err != nil {
				return fmt.Errorf("reading quote signature: %v", err)
			}
		default:
			return fmt.Errorf("unknown event in TrustPoint aggregation: %v", h.Type)
		}
	}

	if aikName == "" {
		return errors.New("AIK id not specified")
	}
	w.TrustPointQuote[aikName] = WinWBCLQuote{
		AIKPubDigest:   aikPubHash,
		Quote:          quoteBlob,
		QuoteSignature: quoteSig,
	}
	return nil
}

func (w *WinEvents) readOuterTrustPointAggregation(evt *TaggedEventData) error {
	r := bytes.NewReader(evt.Data)
	r0 := bytes.NewReader(evt.Data)

	for r.Len() > 0 {
		var h microsoftEventHeader
		if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
			return fmt.Errorf("parsing outer TrustPoint aggregation sub-event: %v", err)
		}
		if int64(h.Size) > int64(r.Len()) {
			return fmt.Errorf("outer TrustPoint aggregation sub-event is larger than available data: %d > %d", h.Size, r.Len())
		}

		var err error
		switch h.Type {
		case trustpointAggregation: // the trust point contains an array of quotes
			if err = w.readInnerTrustPointAggregation(r, h.Size); err != nil {
				return fmt.Errorf("parsing inner TrustPoint aggregation (quote): %v", err)
			}
		default: // the trust point contains just a single quote
			if err = w.readInnerTrustPointAggregation(r0, uint32(len(evt.Data))); err != nil {
				return fmt.Errorf("parsing inner TrustPoint aggregation (quote): %v", err)
			}
			return nil
		}
	}

	return nil
}

func (w *WinEvents) readSIPAEvent(r *bytes.Reader, pcr int) error {
	var header microsoftEventHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return err
	}

	switch header.Type {
	case elamAggregation:
		return w.readELAMAggregation(r, header)
	case loadedModuleAggregation:
		return w.readLoadedModuleAggregation(r, header)
	case bootCounter:
		return w.readBootCounter(header, r)
	case eventCounter:
		return w.readEventCounter(header, r)
	case counterID:
		return w.readEventCounterId(header, r)
	case bitlockerUnlock:
		return w.readBitlockerUnlock(header, r, pcr)
	case transferControl:
		return w.readTransferControl(header, r)

	case osKernelDebug, codeIntegrity, bootDebugging, testSigning: // Parse boolean values.
		return w.readBooleanByteEvent(header, r)
	case dataExecutionPrevention: // Parse booleans represented as uint64's.
		return w.readBooleanInt64Event(header, r)

	default:
		// Event type was not handled, consume the data.
		if int(header.Size) > r.Len() {
			return fmt.Errorf("event data len (%d bytes) larger than event length (%d bytes)", header.Size, r.Len())
		}
		tmp := make([]byte, header.Size)
		if err := binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return fmt.Errorf("reading unknown data section of length %d: %w", header.Size, err)
		}

		return ErrUnknownSIPAEvent
	}
}

// readWinEventBlock extracts boot configuration from SIPA events contained in
// the given tagged event.
func (w *WinEvents) readWinEventBlock(evt *TaggedEventData, pcr int) error {
	r := bytes.NewReader(evt.Data)

	// All windows information should be sub events in an enclosing SIPA
	// container event.
	if (windowsEvent(evt.ID) & sipaTypeMask) != sipaContainer {
		return fmt.Errorf("expected container event, got %v", windowsEvent(evt.ID))
	}

	for r.Len() > 0 {
		if err := w.readSIPAEvent(r, pcr); err != nil {
			if errors.Is(err, ErrUnknownSIPAEvent) {
				// Unknown SIPA events are okay as all TCG events are verifiable.
				continue
			}
			return err
		}
	}
	return nil
}

func (e windowsEvent) String() string {
	if s, ok := windowsEventNames[e]; ok {
		return s
	}
	return fmt.Sprintf("windowsEvent(%#v)", uint32(e))
}

func parseMicrosoftEvent(b []byte, parsedEvent *MicrosoftBootEvent) error {
	events, err := parseMicrosoftEventContainer(b)
	if err != nil {
		return err
	}
	parsedEvent.Events = events
	return nil
}

func parseMicrosoftEventContainer(b []byte) ([]MicrosoftEvent, error) {
	var header microsoftEventHeader
	var events []MicrosoftEvent
	r := bytes.NewReader(b)
	for {
		err := binary.Read(r, binary.LittleEndian, &header)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unable to read Windows event: %v", err)
		}
		data := make([]byte, header.Size)
		if err = binary.Read(r, binary.LittleEndian, &data); err != nil {
			return nil, fmt.Errorf("unable to read Windows event: %v", err)
		}
		if (header.Type & sipaTypeMask) == sipaContainer {
			ret, err := parseMicrosoftEventContainer(data)
			if err != nil {
				return nil, fmt.Errorf("unable to parse Windows event container: %v", err)
			}
			events = append(events, ret...)
			continue
		}
		buf := bytes.NewBuffer(data)
		switch header.Type {
		case filePath, authorityIssuer, authorityPublisher, systemRoot, elamKeyname:
			var Event MicrosoftStringEvent
			var utf16data []uint16
			for i := 0; i < int(header.Size)/2; i++ {
				var tmp uint16
				if err = binary.Read(buf, binary.LittleEndian, &tmp); err != nil {
					return nil, fmt.Errorf("unable to read Windows event data: %v", err)
				}
				utf16data = append(utf16data, tmp)
			}

			Event.Message = string(utf16.Decode(utf16data))
			Event.Type = header.Type
			events = append(events, Event)
		case bootRevocationList:
			var Event MicrosoftRevocationEvent
			var digestLen uint32

			if err = binary.Read(buf, binary.LittleEndian, &Event.CreationTime); err != nil {
				return nil, fmt.Errorf("unable to read Windows revocation list event: %v", err)
			}
			if err = binary.Read(buf, binary.LittleEndian, &digestLen); err != nil {
				return nil, fmt.Errorf("unable to read Windows revocation list event: %v", err)
			}
			if err = binary.Read(buf, binary.LittleEndian, &Event.HashAlgorithm); err != nil {
				return nil, fmt.Errorf("unable to read Windows revocation list event: %v", err)
			}
			Event.Digest = make([]byte, digestLen)
			if err = binary.Read(buf, binary.LittleEndian, &Event.Digest); err != nil {
				return nil, fmt.Errorf("unable to read Windows revocation list event: %v", err)
			}

			Event.Type = header.Type
			events = append(events, Event)
		default:
			var Event MicrosoftDataEvent
			Event.Type = header.Type
			Event.Data = data
			events = append(events, Event)
		}
	}
	return events, nil
}
