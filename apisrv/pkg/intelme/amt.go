// Intel AMT Implementation and reference guide and
// https://github.com/intel/lms/tree/e0ebda9d1e7884b51293b71c1bcda511a7942e1a/MEIClient/AMTHIClient/Include

package intelme

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	AMTGetProvisioningMode            = 0x04000008
	AMTGetProvisioningState           = 0x04000011
	AMTDisableAndClearAMT             = 0x04000088
	AMTGetRemoteAccessConnectionState = 0x04000046
	AMTGetSessionState                = 0x04000049
	AMTEnumerateHashHandles           = 0
	AMTGenerateRngSeed                = 0
	AMTGetCertificateHashEntry        = 0
	AMTGetCodeVersions                = 0x0400001a
	AMTGetSecurityParameters          = 0x0400001b
	AMTGetLocalSystemAccount          = 0
	AMTGetMESetupAuditRecord          = 0x04000050
	AMTGetAuditLogRecords             = 0x04000087
	AMTGetAuditLogSignature           = 0x0000008c
	AMTGetPID                         = 0x0400003b
	AMTGetPKIFQDNSuffix               = 0
	AMTGetProvisioningTLSMode         = 0x04000087
	AMTGetUUID                        = 0x0400005c
	AMTGetRngSeedStatus               = 0
	AMTGetZTECEnabledStatus           = 0
	AMTGetProvisioningServerOTP       = 0
	AMTGetZeroTouchEnabled            = 0x04000030
	AMTSetEHBCState                   = 0
	AMTGetEHBCState                   = 0
	AMTStartConfiguration             = 0
	AMTStartConfigurationEx           = 0
	AMTStartConfigurationHBased       = 0
	AMTStopConfiguration              = 0
	AMTUnprovision                    = 0
	AMTSetAMTOperationalState         = 0
	AMTChangeToAMT                    = 0
	AMTIsChangeToAMTEnabled           = 0
	AMTGetState                       = 0x01000001

	// Status codes
	AMTSuccess = 0

	// AMTGetSessionState session IDs
	redirectionSession = 0
	webUISession       = 2
	kvmSession         = 4
)

var (
	// An internal error in the Intel® AMT device has occurred
	ErrAMTInternal = errors.New("InternalError")

	// Intel® AMT device has not progressed far enough in its
	// initialization to process the command.
	ErrAMTNotReady = errors.New("NotReady")

	/**  Added For ME Tool Legacy Support - 14-Sep-2008  **/
	ErrAMTInvalidAMTMode = errors.New("InvalidAMTMode")

	// Length field of header is invalid.
	ErrAMTInvalidMessageLength = errors.New("InvalidMessageLength")

	// The requested hardware asset inventory table
	// checksum is not available.
	ErrAMTTableFingerprintNotAvailable = errors.New("TableFingerprintNotAvailable")

	// The Integrity Check Value field of the request
	// message sent by Intel® AMT enabled device is invalid.
	ErrAMTIntegrityCheckFailed = errors.New("IntegrityCheckFailed")

	// The specified ISV version is not supported
	ErrAMTUnsupportedIsvsVersion = errors.New("UnsupportedIsvsVersion")

	// The specified queried application is not registered.
	ErrAMTApplicationNotRegistered = errors.New("ApplicationNotRegistered")

	// Either an invalid name or a not previously registered
	//“Enterprise” name was specified
	ErrAMTInvalidRegistrationData = errors.New("InvalidRegistrationData")

	// The application handle provided in the request
	// message has never been allocated.
	ErrAMTApplicationDoesNotExist = errors.New("ApplicationDoesNotExist")

	// The requested number of bytes cannot be allocated in ISV storage.
	ErrAMTNotEnoughStorage = errors.New("NotEnoughStorage")

	// The specified name is invalid.
	ErrAMTInvalidName = errors.New("InvalidName")

	// The specified block does not exist.
	ErrAMTBlockDoesNotExist = errors.New("BlockDoesNotExist")

	// The specified byte offset is invalid.
	ErrAMTInvalidByteOffset = errors.New("InvalidByteOffset")

	// The specified byte count is invalid.
	ErrAMTInvalidByteCount = errors.New("InvalidByteCount")

	// The requesting application is not
	// permitted to request execution of the specified operation.
	ErrAMTNotPermitted = errors.New("NotPermitted")

	// The requesting application is not the owner of the block
	// as required for the requested operation.
	ErrAMTNotOwner = errors.New("NotOwner")

	// The specified block is locked by another application.
	ErrAMTBlockLockedByOther = errors.New("BlockLockedByOther")

	// The specified block is not locked.
	ErrAMTBlockNotLocked = errors.New("BlockNotLocked")

	// The specified group permission bits are invalid.
	ErrAMTInvalidGroupPermissions = errors.New("InvalidGroupPermissions")

	// The specified group does not exist.
	ErrAMTGroupDoesNotExist = errors.New("GroupDoesNotExist")

	// The specified member count is invalid.
	ErrAMTInvalidMemberCount = errors.New("InvalidMemberCount")

	// The request cannot be satisfied because a maximum
	// limit associated with the request has been reached.
	ErrAMTMaxLimitReached = errors.New("MaxLimitReached")

	// specified key algorithm is invalid.
	ErrAMTInvalidAuthType = errors.New("InvalidAuthType")

	// Authentication failed
	ErrAMTAuthenticationFailed = errors.New("AuthenticationFailed")

	// The specified DHCP mode is invalid.
	ErrAMTInvalidDHCPMode = errors.New("InvalidDHCPMode")

	// The specified IP address is not a valid IP unicast address.
	ErrAMTInvalidIPAddress = errors.New("InvalidIPAddress")

	// The specified domain name is not a valid domain name.
	ErrAMTInvalidDomainName = errors.New("InvalidDomainName")

	// Not Used
	ErrAMTUnsupportedVersion = errors.New("UnsupportedVersion")

	// The requested operation cannot be performed because a
	// prerequisite request message has not been received.
	ErrAMTRequestUnexpected = errors.New("RequestUnexpected")

	// Not Used
	ErrAMTInvalidTableType = errors.New("InvalidTableType")

	// The specified provisioning mode code is undefined.
	ErrAMTInvalidProvisioningState = errors.New("InvalidProvisioningState")

	// Not Used
	ErrAMTUnsupportedObject = errors.New("UnsupportedObject")

	// The specified time was not accepted by the Intel® AMT device
	// since it is earlier than the baseline time set for the device.
	ErrAMTInvalidTime = errors.New("InvalidTime")

	// StartingIndex is invalid.
	ErrAMTInvalidIndex = errors.New("InvalidIndex")

	// A parameter is invalid.
	ErrAMTInvalidParameter = errors.New("InvalidParameter")

	// An invalid netmask was supplied
	//(a valid netmask is an IP address in which all ‘1’s are before
	// the ‘0’ – e.g. FFFC0000h is valid, FF0C0000h is invalid).
	ErrAMTInvalidNetmask = errors.New("InvalidNetmask")

	// The operation failed because the Flash wear-out
	// protection mechanism prevented a write to an NVRAM sector.
	ErrAMTFlashWriteLimitExceeded = errors.New("FlashWriteLimitExceeded")

	// ME FW did not receive the entire image file.
	ErrAMTInvalidImageLength = errors.New("InvalidImageLength")

	// ME FW received an image file with an invalid signature.
	ErrAMTInvalidImageSignature = errors.New("InvalidImageSignature")

	// LME can not support the requested version.
	ErrAMTProposeAnotherVersion = errors.New("ProposeAnotherVersion")

	// The PID must be a 64 bit quantity made up of ASCII codes
	// of some combination of 8 characters –
	// capital alphabets (A–Z) and numbers (0–9).
	ErrAMTInvalidPIDFormat = errors.New("InvalidPIDFormat")

	// The PPS must be a 256 bit quantity made up of ASCII codes
	// of some combination of 32 characters –
	// capital alphabets (A–Z) and numbers (0–9).
	ErrAMTInvalidPPSFormat = errors.New("InvalidPPSFormat")

	// Full BIST test has been blocked
	ErrAMTBistCommandBlocked = errors.New("BistCommandBlocked")

	// A TCP/IP connection could not be opened on with the selected port.
	ErrAMTConnectionFailed = errors.New("ConnectionFailed")

	// Max number of connection reached.
	// LME can not open the requested connection.
	ErrAMTConnectionTooMany = errors.New("ConnectionTooMany")

	// Random key generation is in progress.
	ErrAMTRNGGenerationInProgress = errors.New("RNGGenerationInProgress")

	// A randomly generated key does not exist.
	ErrAMTRNGNotReady = errors.New("RNGNotReady")

	// Self-generated AMT certificate does not exist.
	ErrAMTCertificateNotReady = errors.New("CertificateNotReady")

	// Operetion disabled by policy
	ErrAMTDisabledByPolicy = errors.New("DisabledByPolicy")

	// This code establishes a dividing line between
	// status codes which are common to host interface and
	// network interface and status codes which are used by
	// network interface only.
	ErrAMTNetworkIfErrorBase = errors.New("NetworkIfErrorBase")

	// The OEM number specified in the remote control
	// command is not supported by the Intel® AMT device
	ErrAMTUnsupportedOEMNumber = errors.New("UnsupportedOEMNumber")

	// The boot option specified in the remote control command
	// is not supported by the Intel® AMT device
	ErrAMTUnsupportedBootOption = errors.New("UnsupportedBootOption")

	// The command specified in the remote control command
	// is not supported by the Intel® AMT device
	ErrAMTInvalidCommand = errors.New("InvalidCommand")

	// The special command specified in the remote control command
	// is not supported by the Intel® AMT device
	ErrAMTInvalidSpecialCommand = errors.New("InvalidSpecialCommand")

	// The handle specified in the command is invalid
	ErrAMTInvalidHandle = errors.New("InvalidHandle")

	// The password specified in the User ACL is invalid
	ErrAMTInvalidPassword = errors.New("InvalidPassword")

	// The realm specified in the User ACL is invalid
	ErrAMTInvalidRealm = errors.New("InvalidRealm")

	// The FPACL or EACL entry is used by an active
	// registration and cannot be removed or modified.
	ErrAMTStorageACLEntryInUse = errors.New("StorageACLEntryInUse")

	// Essential data is missing on CommitChanges command.
	ErrAMTDataMissing = errors.New("DataMissing")

	// The parameter specified is a duplicate of an existing value.
	// Returned for a case where duplicate entries are added to FPACL
	//(Factory Partner Allocation Control List) or EACL
	//(Enterprise Access Control List) lists.
	ErrAMTDuplicate = errors.New("Duplicate")

	// Event Log operation failed due to the current freeze status of the log.
	ErrAMTEventlogFrozen = errors.New("EventlogFrozen")

	// The device is missing private key material.
	ErrAMTPKIMissingKeys = errors.New("PKIMissingKeys")

	// The device is currently generating a keypair.
	// Caller may try repeating this operation at a later time.
	ErrAMTPKIGeneratingKeys = errors.New("PKIGeneratingKeys")

	// An invalid Key was entered.
	ErrAMTInvalidKey = errors.New("InvalidKey")

	// An invalid X.509 certificate was entered.
	ErrAMTInvalidCert = errors.New("InvalidCert")

	// Certificate Chain and Private Key do not match.
	ErrAMTCertKeyNotMatch = errors.New("CertKeyNotMatch")

	// The request cannot be satisfied because the maximum
	// number of allowed Kerberos domains has been reached.
	//(The domain is determined by the first 24 Bytes of the SID.)
	ErrAMTMaxKerbDomainReached = errors.New("MaxKerbDomainReached")

	// The requested configuration is unsupported
	ErrAMTUnsupported = errors.New("Unsupported")

	// A profile with the requested priority already exists
	ErrAMTInvalidPriority = errors.New("InvalidPriority")

	// Unable to find specified element
	ErrAMTNotFound = errors.New("NotFound")

	// Invalid User credentials
	ErrAMTInvalidCredentials = errors.New("InvalidCredentials")

	// Passphrase is invalid
	ErrAMTInvalidPassphrase = errors.New("InvalidPassphrase")

	// A certificate handle must be chosen before the
	// operation can be completed.
	ErrAMTNoAssociation = errors.New("NoAssociation")

	// The command is defined as Audit Log event and can not be
	// logged.
	ErrAMTAuditFail = errors.New("AuditFail")

	// One of the ME components is not ready for unprovisioning.
	ErrAMTBlockingComponent = errors.New("BlockingComponent")

	ErrAMTIPv6InterfaceDisabled     = errors.New("IPv6InterfaceDisabled")
	ErrAMTIPv6InterfaceDoesNotExist = errors.New("IPv6InterfaceDoesNotExist")

	errorMap = map[uint32]error{
		0x1:   ErrAMTInternal,
		0x2:   ErrAMTNotReady,
		0x3:   ErrAMTInvalidAMTMode,
		0x4:   ErrAMTInvalidMessageLength,
		0x5:   ErrAMTTableFingerprintNotAvailable,
		0x6:   ErrAMTIntegrityCheckFailed,
		0x7:   ErrAMTUnsupportedIsvsVersion,
		0x8:   ErrAMTApplicationNotRegistered,
		0x9:   ErrAMTInvalidRegistrationData,
		0xA:   ErrAMTApplicationDoesNotExist,
		0xB:   ErrAMTNotEnoughStorage,
		0xC:   ErrAMTInvalidName,
		0xD:   ErrAMTBlockDoesNotExist,
		0xE:   ErrAMTInvalidByteOffset,
		0xF:   ErrAMTInvalidByteCount,
		0x10:  ErrAMTNotPermitted,
		0x11:  ErrAMTNotOwner,
		0x12:  ErrAMTBlockLockedByOther,
		0x13:  ErrAMTBlockNotLocked,
		0x14:  ErrAMTInvalidGroupPermissions,
		0x15:  ErrAMTGroupDoesNotExist,
		0x16:  ErrAMTInvalidMemberCount,
		0x17:  ErrAMTMaxLimitReached,
		0x18:  ErrAMTInvalidAuthType,
		0x19:  ErrAMTAuthenticationFailed,
		0x1A:  ErrAMTInvalidDHCPMode,
		0x1B:  ErrAMTInvalidIPAddress,
		0x1C:  ErrAMTInvalidDomainName,
		0x1D:  ErrAMTUnsupportedVersion,
		0x1E:  ErrAMTRequestUnexpected,
		0x1F:  ErrAMTInvalidTableType,
		0x20:  ErrAMTInvalidProvisioningState,
		0x21:  ErrAMTUnsupportedObject,
		0x22:  ErrAMTInvalidTime,
		0x23:  ErrAMTInvalidIndex,
		0x24:  ErrAMTInvalidParameter,
		0x25:  ErrAMTInvalidNetmask,
		0x26:  ErrAMTFlashWriteLimitExceeded,
		0x27:  ErrAMTInvalidImageLength,
		0x28:  ErrAMTInvalidImageSignature,
		0x29:  ErrAMTProposeAnotherVersion,
		0x2A:  ErrAMTInvalidPIDFormat,
		0x2B:  ErrAMTInvalidPPSFormat,
		0x2C:  ErrAMTBistCommandBlocked,
		0x2D:  ErrAMTConnectionFailed,
		0x2E:  ErrAMTConnectionTooMany,
		0x2F:  ErrAMTRNGGenerationInProgress,
		0x30:  ErrAMTRNGNotReady,
		0x31:  ErrAMTCertificateNotReady,
		0x400: ErrAMTDisabledByPolicy,
		0x800: ErrAMTNetworkIfErrorBase,
		0x801: ErrAMTUnsupportedOEMNumber,
		0x802: ErrAMTUnsupportedBootOption,
		0x803: ErrAMTInvalidCommand,
		0x804: ErrAMTInvalidSpecialCommand,
		0x805: ErrAMTInvalidHandle,
		0x806: ErrAMTInvalidPassword,
		0x807: ErrAMTInvalidRealm,
		0x808: ErrAMTStorageACLEntryInUse,
		0x809: ErrAMTDataMissing,
		0x80A: ErrAMTDuplicate,
		0x80B: ErrAMTEventlogFrozen,
		0x80C: ErrAMTPKIMissingKeys,
		0x80D: ErrAMTPKIGeneratingKeys,
		0x80E: ErrAMTInvalidKey,
		0x80F: ErrAMTInvalidCert,
		0x810: ErrAMTCertKeyNotMatch,
		0x811: ErrAMTMaxKerbDomainReached,
		0x812: ErrAMTUnsupported,
		0x813: ErrAMTInvalidPriority,
		0x814: ErrAMTNotFound,
		0x815: ErrAMTInvalidCredentials,
		0x816: ErrAMTInvalidPassphrase,
		0x818: ErrAMTNoAssociation,
		0x81B: ErrAMTAuditFail,
		0x81C: ErrAMTBlockingComponent,
		2500:  ErrAMTIPv6InterfaceDisabled,
		2501:  ErrAMTIPv6InterfaceDoesNotExist,
	}
)

type AMTHeader struct {
	MajorVersion uint8
	MinorVersion uint8
	Reserved     uint16
	Operation    uint32
	Length       uint32
}

type AMTDateTime struct {
	Year      uint16
	Month     uint16
	DayOfWeek uint16
	Day       uint16
	Hour      uint16
	Minute    uint16
	Second    uint16
}

const (
	responseHeaderLength = 1 + 1 + 2 + 4 + 4 + 4
)

func encodeAMT(op uint32, payload []byte) []byte {
	req := AMTHeader{
		MajorVersion: 1,
		MinorVersion: 1,
		Reserved:     0,
		Operation:    op,
		Length:       uint32(len(payload)),
	}

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, req)
	_ = binary.Write(buf, binary.LittleEndian, payload)
	return buf.Bytes()
}

func normalizeString(str string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	// The isOk filter is such that there is no need to chain to norm.NFC
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	// This Transformer could also trivially be applied as an io.Reader
	// or io.Writer filter to automatically do such filtering when reading
	// or writing data anywhere.
	str, _, _ = transform.String(t, str)
	return str
}

func decodeCodeVersionString(rd io.Reader) (string, error) {
	var slen uint16
	if err := binary.Read(rd, binary.LittleEndian, &slen); err != nil {
		return "", err
	}

	var str [20]byte
	if err := binary.Read(rd, binary.LittleEndian, &str); err != nil {
		return "", err
	}

	return string(str[:slen]), nil
}

func decodeString(rd io.Reader) (string, error) {
	var slen uint8
	if err := binary.Read(rd, binary.LittleEndian, &slen); err != nil {
		return "", err
	}

	buf := make([]byte, int(slen))
	if err := binary.Read(rd, binary.LittleEndian, &buf); err != nil {
		return "", err
	}

	return normalizeString(string(buf)), nil
}

func decodeAMT(ctx context.Context, buf []byte, op uint32, response interface{}) error {
	var header AMTHeader
	var status uint32

	reader := bytes.NewReader(buf)
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		tel.Log(ctx).WithError(err).Error("Failed parsing response header")
		return ErrHeader
	}
	if header.MajorVersion != 1 || header.MinorVersion != 1 {
		tel.Log(ctx).WithField("received", fmt.Sprintf("%d.%d", header.MajorVersion, header.MinorVersion)).
			WithField("expected", "1.1").Error("Wrong reponse version")
		return ErrHeader
	}
	if header.Operation != (op | 0x0080_0000) {
		tel.Log(ctx).WithField("received", fmt.Sprintf("0x%08x", header.Operation)).
			WithField("expected", fmt.Sprintf("0x%08x", op|0x0080_0000)).Error("Wrong response code")
		return ErrHeader
	}
	if err := binary.Read(reader, binary.LittleEndian, &status); err != nil {
		tel.Log(ctx).WithError(err).Error("Failed parsing response header")
		return ErrHeader
	}
	if status != AMTSuccess {
		if err, ok := errorMap[status]; ok {
			tel.Log(ctx).WithField("status", status).WithError(err).Error("Response status not success")
		} else {
			tel.Log(ctx).WithField("status", status).Error("Response status not success")
		}
		return ErrHeader
	}

	if response == nil {
		return nil
	}
	if err := binary.Read(reader, binary.LittleEndian, response); err != nil {
		return err
	}

	return nil
}

type AMTDisableAndClearAMTRequest struct {
	Header AMTHeader
}

type AMTDisableAndClearAMTResponse struct {
}

type AMTEnumerateHashHandlesResponse struct {
	EntriesCount uint32
	// Handle [EntriesCount]uint32
}

type AMTMESetupAuditRecord struct {
	ProvisioningTLSMode uint8
	SecureDNS           uint8
	HostInitiated       uint8
	SelectedHashType    uint8
	//SelectedHashData    [64]byte
	//CACertificateSerials   [48]byte
	//AdditionalCASerials    uint8
	//IsOEMDefault           uint8
	//IsTimeValid            uint8
	//ProvisioningServerIP   [46]byte
	//TLSStartDate AMTDateTime
	//ProvisioningServerFQDN [1]byte // until EOF
}

type getAuditLogRecords struct {
	Total    uint32
	Returned uint32
	Bytes    uint32
}

const (
	httpDigestUsername = 0
	kerberosSID        = 1
	local              = 2
	kvmDefaultPort     = 3

	ipv4Address = 0
	ipv6Address = 1
	noAddress   = 2
)

var (
	auditAppID = map[uint16]string{
		16: "security-admin",
		17: "rco",
		18: "redir-manager",
		19: "firmware-update-manager",
		20: "security-audit-log",
		21: "network-time",
		22: "network-admin",
		23: "storage-admin",
		24: "event-manager",
		25: "cb-manager",
		26: "agent-presence-manager",
		27: "wireless-config",
		28: "eac",
		29: "kvm",
		30: "user-opt-in",
		31: "bios-management",
		32: "screen-blanking",
		33: "watchdog",
	}

	auditEventID = map[uint16][]string{
		// security-admin
		16: {
			"amt-provisioning-started",
			"amt-provisioning-completed",
			"acl-entry-added",
			"acl-entry-modified",
			"acl-entry-removed",
			"acl-access-with-invalid-credentials",
			"acl-entry-enabled",
			"tls-state-changed",
			"tls-server-certificate-set",
			"tls-server-certificate-remove",
			"tls-trusted-root-certificate-added",
			"tls-trusted-root-certificate-removed",
			"tls-pre-shared-key-set",
			"kerberos-settings-modified",
			"kerberos-master-key-modified",
			"flash-wear-out-counters-reset",
			"power-package-modified",
			"set-realm-authentication-mode",
			"upgrade-client-to-admin",
			"unprovisioning-completed",
		},
		// rco
		17: {
			"performed-power-up",
			"performed-power-down",
			"performed-power-cycle",
			"performed-reset",
			"set-boot-options",
		},
		// redir-manager
		18: {
			"ide-r-session-opend",
			"ide-r-session-closed",
			"ide-r-enabled",
			"ide-r-disabled",
			"sol-session-opend",
			"sol-session-closed",
			"sol-enabled",
			"sol-disabled",
			"kvm-started",
			"kvm-ended",
			"kvm-enabled",
			"kvm-disabled",
			"vnc-pwd-failed-3-times",
		},
		// firmware-update-manager
		19: {
			"firmware-updated",
			"firmware-updated-failed",
		},
		// security-audit-log
		20: {
			"audit-log-cleared",
			"audit-log-modified",
			"audit-log-disabled",
			"audit-log-enabled",
			"audit-log-exported",
			"audit-log-recovery",
		},

		//const static unsigned short max-security-audit-recovery-reason = 2;
		//
		//const static std::string securityauditlogrecoveryreason[] = {"unknown",
		//														"migration failure",
		//														"initialization failure"};
		//
		//const static unsigned short max-interface-id-gen-type-strings= 3;
		//const static std::string interfaceidgentypestrings[] = {"random id", "intel id", "manual id", "invalid id"};
		// network-time
		21: {
			"amt-time-set",
		},
		// network-admin
		22: {
			"tcpip-parameters-set",
			"host-name-set",
			"domain-name-set",
			"vlan-parameters-set",
			"link-policy-set",
			"ipv6-params-set",
		},
		// storage-admin
		23: {
			"global-storage-attributes-set",
			"storage-eacl-modified",
			"storage-fpacl-modified",
			"storage-write-operation",
		},
		// event-manager
		24: {
			"alert-subscribed",
			"alert-unsubscribed",
			"event-log-cleared",
			"event-log-frozen",
		},
		// cb-manager
		25: {
			"cb-filter-added",
			"cb-filter-remove",
			"cb-policy-added",
			"cb-policy-remove",
			"cb-default-policy-set",
			"cb-heuristics-option-set",
			"cb-heuristics-state-cleared",
		},
		// agent-presence-manager
		26: {
			"agent-watchdog-added",
			"agent-watchdog-removed",
			"agent-watchdog-action-set",
		},
		// wireless-config
		27: {
			"wireless-profile-added",
			"wireless-profile-removed",
			"wireless-profile-updated",
			"wireless-profile-sync",
			"wireless-link-preference-changed",
			"wireless-uefi-profile-sync",
		},
		// eac
		28: {
			"eac-posture-signer-set",
			"eac-enabled",
			"eac-disabled",
			"eac-posture-state-update",
			"eac-set-options",
		},
		// kvm
		29: {
			"kvm-opt-in-enabled",
			"kvm-opt-in-disabled",
			"kvm-pwd-changed",
			"kvm-consent-succeeded",
			"kvm-consent-failed",
		},
		// user opt-in
		32: {
			"opt-in-policy-change",
			"send-consent-code-event",
			"start-opt-in-blocked-event",
		},
		// watchdog
		33: {
			"none",
			"wei-watchdog-reset-triggering-options-changed",
			"wei-watchdog-action-pairing-changed",
		},
	}
)

type AMTAuditLogRecord struct {
	AuditAppID    string
	EventID       string
	InitiatorType string

	// InitiatorType == HTTPDigestUsername
	Username *string

	// InitiatorType == KerberosSID
	UserInDomain *uint32
	Domain       *string

	Timestamp  time.Time
	NetAddress *net.IP

	// Extended data
	// firmware-update-manager -> firmware-updated
	PreviousFirmwareVersion []int
	FirmwareVersion         []int
	// firmware-update-manager -> firmware-update-failed
	UpdateFailureReason *int
}

func EncodeGetAMTAuditLogRecords() []byte {
	one := []byte{1, 0, 0, 0}
	return encodeAMT(AMTGetAuditLogRecords, one)
}

func DecodeGetAMTAuditLogRecords(ctx context.Context, buf []byte) ([]AMTAuditLogRecord, error) {
	var hdr getAuditLogRecords
	err := decodeAMT(ctx, buf, AMTGetAuditLogRecords, &hdr)
	if err != nil {
		return nil, err
	}
	if hdr.Returned > 1024 {
		return nil, ErrFormat
	}

	rd := bytes.NewReader(buf[responseHeaderLength+12:])
	records := make([]AMTAuditLogRecord, hdr.Returned)
	for i := 0; i < int(hdr.Returned); i += 1 {
		rec := &records[i]
		var length, audit, event uint16
		var init uint8
		if err := binary.Read(rd, binary.LittleEndian, &length); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.BigEndian, &audit); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.BigEndian, &event); err != nil {
			return nil, err
		}
		if err := binary.Read(rd, binary.BigEndian, &init); err != nil {
			return nil, err
		}

		if a, ok := auditAppID[audit]; ok {
			rec.AuditAppID = a
		} else {
			rec.AuditAppID = "invalid"
		}
		if evs, ok := auditEventID[audit]; ok {
			if int(event) < len(evs) {
				rec.EventID = evs[int(event)]
			} else {
				rec.EventID = "invalid"
			}
		} else {
			rec.EventID = "invalid"
		}

		switch init {
		case httpDigestUsername:
			rec.InitiatorType = "http-digest-username"
			username, err := decodeString(rd)
			if err != nil {
				return nil, err
			}
			rec.Username = &username

		case kerberosSID:
			rec.InitiatorType = "kerberos"
			var userInDomain uint32
			if err := binary.Read(rd, binary.BigEndian, &userInDomain); err != nil {
				return nil, err
			}
			domain, err := decodeString(rd)
			if err != nil {
				return nil, err
			}
			rec.UserInDomain = &userInDomain
			rec.Domain = &domain

		case local:
			rec.InitiatorType = "local-user"
		case kvmDefaultPort:
			rec.InitiatorType = "kvm-user"
		default:
			rec.InitiatorType = "invalid"
		}

		var ts uint32
		if err := binary.Read(rd, binary.BigEndian, &ts); err != nil {
			return nil, err
		}
		rec.Timestamp = time.Unix(int64(ts), 0)

		var mctype uint8
		if err := binary.Read(rd, binary.BigEndian, &mctype); err != nil {
			return nil, err
		}
		addrStr, err := decodeString(rd)
		if err != nil {
			return nil, err
		}
		if len(addrStr) > 0 {
			switch mctype {
			case ipv4Address:
				fallthrough
			case ipv6Address:
				addr, _, err := net.ParseCIDR(addrStr)
				if err != nil {
					return nil, err
				}
				rec.NetAddress = &addr
			default:
			}
		}

		var edlen uint8
		if err := binary.Read(rd, binary.BigEndian, &edlen); err != nil {
			return nil, err
		}
		ed := make([]byte, int(edlen))
		if err := binary.Read(rd, binary.BigEndian, &ed); err != nil {
			return nil, err
		}
		edRd := bytes.NewReader(ed)

		switch rec.AuditAppID {
		case "security-admin":
			switch rec.EventID {
			case "amt-provisioning-completed":
				// XXX
			case "acl-entry-added":
				// XXX
			case "acl-entry-modified":
				// XXX
			case "acl-entry-removed":
				// XXX
			case "acl-access-with-invalid-credentials":
				// XXX
			case "acl-entry-enabled":
				// XXX
			case "tls-state-changed":
				// XXX
			case "tls-server-certificate-set":
				fallthrough
			case "tls-server-certificate-remove":
				fallthrough
			case "tls-trusted-root-certificate-added":
				fallthrough
			case "tls-trusted-root-certificate-removed":
				// XXX
			case "kerberos-settings-modified":
				// XXX
			case "power-package-modified":
				// XXX
			case "set-realm-authentication-mode":
				// XXX
			case "unprovisioning-completed":
				// XXX
			default:
			}
		case "rco":
			// XXX
		case "firmware-update-manager":
			switch rec.EventID {
			case "firmware-updated":
				var oldVer, newVer [4]uint16
				if err := binary.Read(edRd, binary.BigEndian, oldVer[:]); err != nil {
					return nil, err
				}
				if err := binary.Read(edRd, binary.BigEndian, newVer[:]); err != nil {
					return nil, err
				}
				rec.PreviousFirmwareVersion = []int{
					int(oldVer[0]),
					int(oldVer[1]),
					int(oldVer[2]),
					int(oldVer[3]),
				}
				rec.FirmwareVersion = []int{
					int(newVer[0]),
					int(newVer[1]),
					int(newVer[2]),
					int(newVer[3]),
				}
			case "firmware-update-failed":
				var reason uint32
				if err := binary.Read(edRd, binary.LittleEndian, &reason); err != nil {
					return nil, err
				}
				reasonInt := int(reason)
				rec.UpdateFailureReason = &reasonInt
			default:
			}
		case "security-audit-log":
			// XXX
		case "network-time":
			// XXX
		case "network-admin":
			// XXX
		case "storage-admin":
			// XXX
		case "event-manager":
			// XXX
		case "cb-manager":
			// XXX
		case "agent-presence-manager":
			// XXX
		case "wireless-config":
			// XXX
		case "eac":
			// XXX
		case "kvm":
			// XXX
		case "user-opt-in":
			// XXX
		case "bios-management":
			// XXX
		case "screen-blanking":
			// XXX
		case "watchdog":
			// XXX
		default:
		}
	}

	return records, nil
}

type AMTAuditLogSignature struct {
	TotalRecordCount        uint32
	StartLogTime            AMTDateTime
	EndLogTime              AMTDateTime
	SignatureGenerationTime AMTDateTime
	AuditLogHash            [64]byte
	AMTNonce                [20]byte
	UUID                    [16]byte
	FQDN                    [256]byte
	FWVersion               [4]uint16
	AMTSVN                  uint32
	SignatureMechanism      uint32
	Signature               [512]byte
	LengthsOfCertificates   [4]uint16
	Certificates            [3000]byte
}

func EncodeGetAMTAuditLogSignature() []byte {
	return encodeAMT(AMTGetAuditLogSignature, make([]byte, 20))
}

func DecodeGetAMTAuditLogSignature(ctx context.Context, buf []byte) (*AMTAuditLogSignature, error) {
	var rec AMTAuditLogSignature
	err := decodeAMT(ctx, buf, AMTGetAuditLogSignature, &rec)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func EncodeGetAMTMESetupAuditRecord() []byte {
	return encodeAMT(AMTGetMESetupAuditRecord, nil)
}

func DecodeGetAMTMESetupAuditRecord(ctx context.Context, buf []byte) (*AMTMESetupAuditRecord, error) {
	var rec AMTMESetupAuditRecord
	err := decodeAMT(ctx, buf, AMTGetMESetupAuditRecord, &rec)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func EncodeGetAMTPID() []byte {
	return encodeAMT(AMTGetPID, nil)
}

func DecodeGetAMTPID(ctx context.Context, buf []byte) ([]byte, error) {
	var pid [8]byte
	if err := decodeAMT(ctx, buf, AMTGetPID, &pid); err != nil {
		return nil, err
	}

	return pid[:], nil
}

const (
	NoMode         = "none"
	EnterpriseMode = "enterprise"
	InvalidMode    = "invalid"
)

type amtProvisioningModeRaw struct {
	Mode   uint32
	Legacy uint8
}

type AMTProvisioningMode struct {
	Mode   string
	Legacy bool
}

func EncodeGetAMTProvisioningMode() []byte {
	return encodeAMT(AMTGetProvisioningMode, nil)
}

func DecodeGetAMTProvisioningMode(ctx context.Context, buf []byte) (*AMTProvisioningMode, error) {
	var raw amtProvisioningModeRaw
	err := decodeAMT(ctx, buf, AMTGetProvisioningMode, &raw)
	if err != nil {
		return nil, err
	}

	var modes AMTProvisioningMode

	switch raw.Mode {
	case 0:
		modes.Mode = NoMode
	case 1:
		modes.Mode = EnterpriseMode
	default:
		modes.Mode = InvalidMode
	}
	modes.Legacy = raw.Legacy == 1

	return &modes, nil
}

const (
	// Provisioning state
	PreProvisioning  = 0
	InProvisioning   = 1
	PostProvisioning = 2
)

type securityParametersRaw struct {
	EnterpriseMode          uint8
	TLSEnabled              uint8
	HWCryptoEnabled         uint8
	ProvisioningState       uint32
	NetworkInterfaceEnabled uint8
	SOLEnabled              uint8
	IDEREnabled             uint8
	FWUpdateEnabled         uint8
	LinkIsUp                uint8
	KvmEnabled              uint8
	Reserved                [7]byte
}

type AMTSecurityParameters struct {
	EnterpriseMode    bool
	TLS               bool
	HardwareCrypto    bool
	ProvisioningState string
	NetworkInterface  bool
	SerialOverLAN     bool
	IDERedirect       bool
	FirmwareUpdate    bool
	NetworkLink       bool
	KVM               bool
}

func EncodeGetAMTSecurityParameters() []byte {
	return encodeAMT(AMTGetSecurityParameters, nil)
}

func DecodeGetAMTSecurityParameters(ctx context.Context, buf []byte) (*AMTSecurityParameters, error) {
	var raw securityParametersRaw
	err := decodeAMT(ctx, buf, AMTGetSecurityParameters, &raw)
	if err != nil {
		return nil, err
	}

	var state string
	switch raw.ProvisioningState {
	case PreProvisioning:
		state = "pre"
	case InProvisioning:
		state = "in"
	case PostProvisioning:
		state = "post"
	default:
		state = "unknown"
	}

	params := AMTSecurityParameters{
		EnterpriseMode:    raw.EnterpriseMode == 1,
		TLS:               raw.TLSEnabled == 1,
		HardwareCrypto:    raw.HWCryptoEnabled == 1,
		NetworkInterface:  raw.NetworkInterfaceEnabled == 1,
		SerialOverLAN:     raw.SOLEnabled == 1,
		IDERedirect:       raw.IDEREnabled == 1,
		FirmwareUpdate:    raw.FWUpdateEnabled == 1,
		NetworkLink:       raw.LinkIsUp == 1,
		KVM:               raw.KvmEnabled == 1,
		ProvisioningState: state,
	}
	return &params, nil
}

func EncodeGetAMTUUID() []byte {
	return encodeAMT(AMTGetUUID, nil)
}

func DecodeGetAMTUUID(ctx context.Context, buf []byte) (uuid.UUID, error) {
	var raw [16]byte
	err := decodeAMT(ctx, buf, AMTGetUUID, &raw)
	if err != nil {
		return uuid.UUID{}, err
	}

	return uuid.FromBytes(raw[:])
}

const (
	ResetNone      = 0
	ResetME        = 1
	ResetGlobal    = 2
	ResetException = 3
)

var (
	linkStateUUID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
)

type amtStateRaw struct {
	UUID            [16]byte
	StateLength     uint32
	LinkStatus      uint8
	Reserved        uint8
	CryptoFuse      uint8
	FlashProtection uint8
	LastMEResetType uint8
}

type AMTState struct {
	UUID            uuid.UUID
	Link            bool
	CryptoFuse      bool
	FlashProtection bool
	LastResetType   int
}

func EncodeGetAMTState() []byte {
	return encodeAMT(AMTGetState, linkStateUUID)
}

func DecodeGetAMTState(ctx context.Context, buf []byte) (*AMTState, error) {
	var raw amtStateRaw
	err := decodeAMT(ctx, buf, AMTGetState, &raw)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(raw.UUID[:], linkStateUUID[:]) {
		return nil, ErrFormat
	}
	if raw.StateLength != 5 {
		return nil, ErrFormat
	}

	uu, err := uuid.FromBytes(raw.UUID[:])
	if err != nil {
		return nil, err
	}
	state := AMTState{
		UUID:            uu,
		Link:            raw.LinkStatus == 1,
		CryptoFuse:      raw.CryptoFuse == 1,
		FlashProtection: raw.FlashProtection == 1,
		LastResetType:   int(raw.LastMEResetType),
	}

	return &state, nil
}

func EncodeGetAMTCodeVersions() []byte {
	return encodeAMT(AMTGetCodeVersions, nil)
}

type AMTCodeVersions struct {
	BIOS     string
	Versions map[string]string
}

func (ver *AMTCodeVersions) ParseVersion(name string) ([]int, error) {
	val, ok := ver.Versions[name]
	if !ok {
		return nil, errors.New("not found")
	}

	comps := strings.Split(val, ".")
	if len(comps) == 0 {
		return nil, errors.New("empty value")
	}

	ret := make([]int, len(comps))
	for i, c := range comps {
		v, err := strconv.ParseInt(c, 10, 64)
		if err != nil {
			return nil, err
		}
		ret[i] = int(v)
	}

	return ret, nil
}

func DecodeGetAMTCodeVersions(ctx context.Context, buf []byte) (*AMTCodeVersions, error) {
	err := decodeAMT(ctx, buf, AMTGetCodeVersions, nil)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(buf[responseHeaderLength:])
	var bios [65]byte
	if err := binary.Read(reader, binary.LittleEndian, &bios); err != nil {
		tel.Log(ctx).WithError(err).Error("AMT CodeVersions BIOS field")
		return nil, ErrFormat
	}

	var num uint32
	if err := binary.Read(reader, binary.LittleEndian, &num); err != nil || num > 128 {
		tel.Log(ctx).WithError(err).Error("AMT CodeVersions num versions field")
		return nil, ErrFormat
	}

	vers := AMTCodeVersions{
		BIOS:     normalizeString(string(bios[:])),
		Versions: make(map[string]string),
	}

	for i := 0; i < int(num); i += 1 {
		key, err := decodeCodeVersionString(reader)
		if err != nil {
			tel.Log(ctx).WithError(err).WithField("iteration", i).Error("AMT CodeVersions versions key")
			return nil, ErrFormat
		}
		value, err := decodeCodeVersionString(reader)
		if err != nil {
			tel.Log(ctx).WithError(err).WithField("iteration", i).Error("AMT CodeVersions versions value")
			return nil, ErrFormat
		}

		vers.Versions[key] = value
	}

	return &vers, nil
}

func EncodeGetAMTZeroTouchEnabled() []byte {
	return encodeAMT(AMTGetZeroTouchEnabled, nil)
}

func DecodeGetAMTZeroTouchEnabled(ctx context.Context, buf []byte) (bool, error) {
	var bit uint8
	err := decodeAMT(ctx, buf, AMTGetZeroTouchEnabled, &bit)
	if err != nil {
		return false, err
	}

	return bit != 0, nil
}

type amtGetRedirectionSessionStateRaw struct {
	Request       uint32
	IDERedirect   uint8
	SerialOverLAN uint8
	Reserved      uint8
}

type AMTRedirectionSessionState struct {
	IDERedirect   bool
	SerialOverLAN bool
}

func EncodeGetAMTRedirectionSessionState() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(redirectionSession))
	return encodeAMT(AMTGetSessionState, buf.Bytes())
}

func DecodeGetAMTRedirectionSessionState(ctx context.Context, buf []byte) (*AMTRedirectionSessionState, error) {
	var raw amtGetRedirectionSessionStateRaw
	err := decodeAMT(ctx, buf, AMTGetSessionState, &raw)
	if err != nil {
		return nil, err
	}

	state := AMTRedirectionSessionState{
		IDERedirect:   raw.IDERedirect == 1,
		SerialOverLAN: raw.SerialOverLAN == 1,
	}

	return &state, nil
}

type amtGetKVMSessionStateRaw struct {
	Request uint32
	Status  uint32
}

type AMTKVMSessionState struct {
	IsActive   bool
	WaitForOpt bool
}

func EncodeGetAMTKVMSessionState() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(kvmSession))
	return encodeAMT(AMTGetSessionState, buf.Bytes())
}

func DecodeGetAMTKVMSessionState(ctx context.Context, buf []byte) (*AMTKVMSessionState, error) {
	var raw amtGetKVMSessionStateRaw
	err := decodeAMT(ctx, buf, AMTGetSessionState, &raw)
	if err != nil {
		return nil, err
	}

	state := AMTKVMSessionState{
		IsActive:   (raw.Status & 1) != 0,
		WaitForOpt: (raw.Status & 0b10) != 0,
	}

	return &state, nil
}

type amtGetWebUIStateRaw struct {
	Request uint32
	Enabled uint32
}

func EncodeGetAMTWebUIState() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(webUISession))
	return encodeAMT(AMTGetSessionState, buf.Bytes())
}

func DecodeGetAMTWebUIState(ctx context.Context, buf []byte) (bool, error) {
	var raw amtGetWebUIStateRaw
	err := decodeAMT(ctx, buf, AMTGetSessionState, &raw)
	if err != nil {
		return false, err
	}

	return raw.Enabled == 1, nil
}

type amtGetRemoteAccessConnectionStateRaw struct {
	NetworkConnectionStatus       uint32
	RemoteAccessConnectionStatus  uint32
	RemoteAccessConnectionTrigger uint32
}

type AMTRemoteAccessConnectionState struct {
	Network string
	Status  string
	Trigger string
}

func EncodeGetAMTRemoteAccessConnectionState() []byte {
	return encodeAMT(AMTGetRemoteAccessConnectionState, nil)
}

func DecodeGetAMTRemoteAccessConnectionState(ctx context.Context, buf []byte) (*AMTRemoteAccessConnectionState, error) {
	var raw amtGetRemoteAccessConnectionStateRaw
	err := decodeAMT(ctx, buf, AMTGetRemoteAccessConnectionState, &raw)
	if err != nil {
		return nil, err
	}

	var state AMTRemoteAccessConnectionState

	switch raw.NetworkConnectionStatus {
	case 0:
		state.Network = "direct"
	case 1:
		state.Network = "vpn"
	case 2:
		state.Network = "outside"
	default:
		state.Network = "unknown"
	}
	switch raw.RemoteAccessConnectionStatus {
	case 0:
		state.Status = "not-connected"
	case 1:
		state.Status = "connecting"
	case 2:
		state.Status = "connected"
	default:
		state.Status = "invalid"
	}
	switch raw.RemoteAccessConnectionTrigger {
	case 0:
		state.Trigger = "user-initiated"
	case 1:
		state.Trigger = "alert"
	case 2:
		state.Trigger = "provisioning"
	case 3:
		state.Trigger = "periodic"
	default:
		state.Trigger = "invalid"
	}

	return &state, nil
}
