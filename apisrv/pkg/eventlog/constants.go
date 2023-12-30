package eventlog

type eventID uint32

const (
	Smbios                 eventID = 0x00
	BISCertificate         eventID = 0x01
	PostBIOSROM            eventID = 0x02
	EscdeventID            eventID = 0x03
	Cmos                   eventID = 0x04
	Nvram                  eventID = 0x05
	OptionROMExecute       eventID = 0x06
	OptionROMConfiguration eventID = 0x07
)

// EventType indicates what kind of data an event is reporting.
//
// https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClientSpecPlat_TPM_2p0_1p04_pub.pdf#page=103
type EventType uint32

// 	BIOS Events (TCG PC Client Specific Implementation Specification for Conventional BIOS 1.21)
const (
	PrebootCert          EventType = 0x00000000
	PostCode             EventType = 0x00000001
	unused               EventType = 0x00000002
	NoAction             EventType = 0x00000003
	Separator            EventType = 0x00000004
	Action               EventType = 0x00000005
	EventTag             EventType = 0x00000006
	SCRTMContents        EventType = 0x00000007
	SCRTMVersion         EventType = 0x00000008
	CpuMicrocode         EventType = 0x00000009
	PlatformConfigFlags  EventType = 0x0000000A
	TableOfDevices       EventType = 0x0000000B
	CompactHash          EventType = 0x0000000C
	Ipl                  EventType = 0x0000000D
	IplPartitionData     EventType = 0x0000000E
	NonhostCode          EventType = 0x0000000F
	NonhostConfig        EventType = 0x00000010
	NonhostInfo          EventType = 0x00000011
	OmitBootDeviceEvents EventType = 0x00000012
)

// EFI Events (TCG EFI Platform Specification Version 1.22)
const (
	EFIEventBase               EventType = 0x80000000
	EFIVariableDriverConfig    EventType = 0x80000001
	EFIVariableBoot            EventType = 0x80000002
	EFIBootServicesApplication EventType = 0x80000003
	EFIBootServicesDriver      EventType = 0x80000004
	EFIRuntimeServicesDriver   EventType = 0x80000005
	EFIGPTEvent                EventType = 0x80000006
	EFIAction                  EventType = 0x80000007
	EFIPlatformFirmwareBlob    EventType = 0x80000008
	EFIHandoffTables           EventType = 0x80000009
	EFIHCRTMEvent              EventType = 0x80000010
	EFIVariableAuthority       EventType = 0x800000e0
)

// TPM algorithms. See the TPM 2.0 specification section 6.3.
//
// https://trustedcomputinggroup.org/wp-content/uploads/TPM-Rev-2.0-Part-2-Structures-01.38.pdf#page=42
const (
	algSHA1   uint16 = 0x0004
	algSHA256 uint16 = 0x000B
)

const (
	// EFIMaxNameLen is the maximum accepted byte length for a name field.
	// This value should be larger than any reasonable value.
	EFIMaxNameLen = 2048
	// EFIMaxDataLen is the maximum size in bytes of a variable data field.
	// This value should be larger than any reasonable value.
	EFIMaxDataLen = 1024 * 1024 // 1 Megabyte.
)

type efiDevicePathType uint8

const (
	hardwareDevicePath  efiDevicePathType = 0x01
	acpiDevicePath      efiDevicePathType = 0x02
	messagingDevicePath efiDevicePathType = 0x03
	mediaDevicePath     efiDevicePathType = 0x04
	bbsDevicePath       efiDevicePathType = 0x05
	endDevicePath       efiDevicePathType = 0x7f
)

type hwDPType uint8

const (
	pciHwDevicePath        hwDPType = 0x01
	pccardHwDevicePath     hwDPType = 0x02
	mmioHwDevicePath       hwDPType = 0x03
	vendorHwDevicePath     hwDPType = 0x04
	controllerHwDevicePath hwDPType = 0x05
	bmcHwDevicePath        hwDPType = 0x06
)

type acpiDPType uint8

const (
	normalACPIDevicePath   acpiDPType = 0x01
	expandedACPIDevicePath acpiDPType = 0x02
	adrACPIDevicePath      acpiDPType = 0x03
)

type messagingDPType uint8

const (
	atapiMessagingDevicePath      messagingDPType = 1
	scsiMessagingDevicePath       messagingDPType = 2
	fcMessagingDevicePath         messagingDPType = 3
	firewireMessagingDevicePath   messagingDPType = 4
	usbMessagingDevicePath        messagingDPType = 5
	i2oMessagingDevicePath        messagingDPType = 6
	infinibandMessagingDevicePath messagingDPType = 9
	vendorMessagingDevicePath     messagingDPType = 10
	macMessagingDevicePath        messagingDPType = 11
	ipv4MessagingDevicePath       messagingDPType = 12
	ipv6MessagingDevicePath       messagingDPType = 13
	uartMessagingDevicePath       messagingDPType = 14
	usbclassMessagingDevicePath   messagingDPType = 15
	usbwwidMessagingDevicePath    messagingDPType = 16
	lunMessagignDevicePath        messagingDPType = 17
	sataMessagingDevicePath       messagingDPType = 18
	iscsiMessagingDevicePath      messagingDPType = 19
	vlanMessagingDevicePath       messagingDPType = 20
	fcExMessagingDevicePath       messagingDPType = 21
	sasExMessagingDevicePath      messagingDPType = 22
	nvmMessagingDevicePath        messagingDPType = 23
	uriMessagingDevicePath        messagingDPType = 24
	ufsMessagingDevicePath        messagingDPType = 25
	sdMessagingDevicePath         messagingDPType = 26
	btMessagingDevicePath         messagingDPType = 27
	wifiMessagingDevicePath       messagingDPType = 28
	emmcMessagingDevicePath       messagingDPType = 29
	btleMessagingDevicePath       messagingDPType = 30
	dnsMessagingDevicePath        messagingDPType = 31
)

type mediaDPType uint8

const (
	hardDriveDevicePath     mediaDPType = 0x01
	cdDriveDevicePath       mediaDPType = 0x02
	vendorDevicePath        mediaDPType = 0x03
	filePathDevicePath      mediaDPType = 0x04
	mediaProtocolDevicePath mediaDPType = 0x05
	piwgFileDevicePath      mediaDPType = 0x06
	piwgVolumeDevicePath    mediaDPType = 0x07
	offsetDevicePath        mediaDPType = 0x08
	ramDiskDevicePath       mediaDPType = 0x09
)

type bbsDPType uint8

const (
	bbs101DevicePath bbsDPType = 0x01
)

type endDPType uint8

const (
	endThisDevicePath   endDPType = 0x01
	endEntireDevicePath endDPType = 0xff
)

type efiSignatureType uint8

const (
	mbr  efiSignatureType = 0x01
	guid efiSignatureType = 0x02
)

const (
	ebsInvocation = "Exit Boot Services Invocation"
	ebsSuccess    = "Exit Boot Services Returned with Success"
	ebsFailure    = "Exit Boot Services Returned with Failure"
)

// SIPA event types
const (
	sipaTypeMask                    windowsEvent = 0x000f0000
	sipaContainer                   windowsEvent = 0x00010000
	sipaInformation                 windowsEvent = 0x00020000
	sipaError                       windowsEvent = 0x00030000
	sipaPreOsParameter              windowsEvent = 0x00040000
	sipaOSParameter                 windowsEvent = 0x00050000
	sipaAuthority                   windowsEvent = 0x00060000
	sipaLoadedModule                windowsEvent = 0x00070000
	sipaTrustPoint                  windowsEvent = 0x00080000
	sipaELAM                        windowsEvent = 0x00090000
	sipaVBS                         windowsEvent = 0x000a0000
	trustBoundary                   windowsEvent = 0x40010001
	elamAggregation                 windowsEvent = 0x40010002
	loadedModuleAggregation         windowsEvent = 0x40010003
	trustpointAggregation           windowsEvent = 0xC0010004
	ksrAggregation                  windowsEvent = 0x40010005
	ksrSignedMeasurementAggregation windowsEvent = 0x40010006
	information                     windowsEvent = 0x00020001
	bootCounter                     windowsEvent = 0x00020002
	transferControl                 windowsEvent = 0x00020003
	applicationReturn               windowsEvent = 0x00020004
	bitlockerUnlock                 windowsEvent = 0x00020005
	eventCounter                    windowsEvent = 0x00020006
	counterID                       windowsEvent = 0x00020007
	morBitNotCancelable             windowsEvent = 0x00020008
	applicationSVN                  windowsEvent = 0x00020009
	svnChainStatus                  windowsEvent = 0x0002000A
	morBitAPIStatus                 windowsEvent = 0x0002000B
	bootDebugging                   windowsEvent = 0x00040001
	bootRevocationList              windowsEvent = 0x00040002
	osKernelDebug                   windowsEvent = 0x00050001
	codeIntegrity                   windowsEvent = 0x00050002
	testSigning                     windowsEvent = 0x00050003
	dataExecutionPrevention         windowsEvent = 0x00050004
	safeMode                        windowsEvent = 0x00050005
	winPE                           windowsEvent = 0x00050006
	physicalAddressExtension        windowsEvent = 0x00050007
	osDevice                        windowsEvent = 0x00050008
	systemRoot                      windowsEvent = 0x00050009
	hypervisorLaunchType            windowsEvent = 0x0005000A
	hypervisorPath                  windowsEvent = 0x0005000B
	hypervisorIOMMUPolicy           windowsEvent = 0x0005000C
	hypervisorDebug                 windowsEvent = 0x0005000D
	driverLoadPolicy                windowsEvent = 0x0005000E
	siPolicy                        windowsEvent = 0x0005000F
	hypervisorMMIONXPolicy          windowsEvent = 0x00050010
	hypervisorMSRFilterPolicy       windowsEvent = 0x00050011
	vsmLaunchType                   windowsEvent = 0x00050012
	osRevocationList                windowsEvent = 0x00050013
	vsmIDKInfo                      windowsEvent = 0x00050020
	flightSigning                   windowsEvent = 0x00050021
	pagefileEncryptionEnabled       windowsEvent = 0x00050022
	vsmIDKSInfo                     windowsEvent = 0x00050023
	hibernationDisabled             windowsEvent = 0x00050024
	dumpsDisabled                   windowsEvent = 0x00050025
	dumpEncryptionEnabled           windowsEvent = 0x00050026
	dumpEncryptionKeyDigest         windowsEvent = 0x00050027
	lsaISOConfig                    windowsEvent = 0x00050028
	noAuthority                     windowsEvent = 0x00060001
	authorityPubKey                 windowsEvent = 0x00060002
	filePath                        windowsEvent = 0x00070001
	imageSize                       windowsEvent = 0x00070002
	hashAlgorithmID                 windowsEvent = 0x00070003
	authenticodeHash                windowsEvent = 0x00070004
	authorityIssuer                 windowsEvent = 0x00070005
	authoritySerial                 windowsEvent = 0x00070006
	imageBase                       windowsEvent = 0x00070007
	authorityPublisher              windowsEvent = 0x00070008
	authoritySHA1Thumbprint         windowsEvent = 0x00070009
	imageValidated                  windowsEvent = 0x0007000A
	moduleSVN                       windowsEvent = 0x0007000B
	quote                           windowsEvent = 0x80080001
	quoteSignature                  windowsEvent = 0x80080002
	aikID                           windowsEvent = 0x80080003
	aikPubDigest                    windowsEvent = 0x80080004
	elamKeyname                     windowsEvent = 0x00090001
	elamConfiguration               windowsEvent = 0x00090002
	elamPolicy                      windowsEvent = 0x00090003
	elamMeasured                    windowsEvent = 0x00090004
	vbsVSMRequired                  windowsEvent = 0x000A0001
	vbsSecurebootRequired           windowsEvent = 0x000A0002
	vbsIOMMURequired                windowsEvent = 0x000A0003
	vbsNXRequired                   windowsEvent = 0x000A0004
	vbsMSRFilteringRequired         windowsEvent = 0x000A0005
	vbsMandatoryEnforcement         windowsEvent = 0x000A0006
	vbsHVCIPolicy                   windowsEvent = 0x000A0007
	vbsMicrosoftBootChainRequired   windowsEvent = 0x000A0008
	ksrSignature                    windowsEvent = 0x000B0001
)

// EV_NO_ACTION is a special event type that indicates information to the parser
// instead of holding a measurement. For TPM 2.0, this event type is used to signal
// switching from SHA1 format to a variable length digest.
//
// https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClientSpecPlat_TPM_2p0_1p04_pub.pdf#page=110
const eventTypeNoAction = 0x03

const (
	wantMajor  = 2
	wantMinor  = 0
	wantErrata = 0
)
