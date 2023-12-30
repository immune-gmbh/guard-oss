package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var allAnnotations []string = []string{
	AnnHostname,
	AnnOSType,
	AnnCPUVendor,
	AnnNoSMBIOS,
	AnnInvalidSMBIOS,
	AnnSMBIOSType0Missing,
	AnnSMBIOSType0Dup,
	AnnSMBIOSType1Missing,
	AnnSMBIOSType1Dup,
	AnnNoEFI,
	AnnNoSecureBoot,
	AnnNoDeployedSecureBoot,
	AnnMissingEventLog,
	AnnModeInvalid,
	AnnPKMissing,
	AnnPKInvalid,
	AnnKEKMissing,
	AnnKEKInvalid,
	AnnDBMissing,
	AnnDBInvalid,
	AnnDBxMissing,
	AnnDBxInvalid,
	AnnNoTXTPubspace,
	AnnNoSGX,
	AnnSGXDisabled,
	AnnSGXCaps0Missing,
	AnnSGXCaps1Missing,
	AnnSGXCaps29Missing,
	AnnNoTPM,
	AnnNoTPMManufacturer,
	AnnInvalidTPMManufacturer,
	AnnNoTPMVendorID,
	AnnInvalidTPMVendorID,
	AnnNoTPMSpecVersion,
	AnnInvalidTPMSpecVersion,
	AnnEventLogMissing,
	AnnEventLogInvalid,
	AnnEventLogBad,
	AnnPCRInvalid,
	AnnPCRMissing,
	AnnNoSEV,
	AnnSEVDisabled,
	AnnPlatformStatusMissing,
	AnnPlatformStatusInvalid,
	AnnNoMEDevice,
	AnnNoMEAccess,
	AnnMEConfigSpaceInvalid,
	AnnMEVariantInvalid,
	AnnMEVersionMissing,
	AnnMEVersionInvalid,
	AnnMEFeaturesMissing,
	AnnMEFeaturesInvalid,
	AnnMEFWUPMissing,
	AnnMEFWUPInvalid,
	AnnMEBadChecksum,
	AnnMEBUPFailure,
	AnnMEFlashFailure,
	AnnMEFSCorruption,
	AnnMECPUInvalid,
	AnnMEUncatError,
	AnnMEImageError,
	AnnMEFatalError,
	AnnMEInvalidWorkingState,
	AnnMEM0Error,
	AnnMESPIDataInvalid,
	AnnMEUpdating,
	AnnMEHalted,
	AnnMEManufacturingMode,
	AnnMEUnlocked,
	AnnMEDebugMode,
	AnnMEResetRequest,
	AnnMEMBEXRequest,
	AnnMECPUReplaced,
	AnnMESafeBoot,
	AnnMEWasReset,
	AnnMEDisabled,
	AnnMEBooting,
	AnnMEInRecovery,
	AnnMESVNDisabled,
	AnnMESVNUpdated,
	AnnMESVNOutdated,
	AnnAMTSKUInvalid,
	AnnAMTVersionInvalid,
	AnnAMTVersionMissing,
	AnnAMTAuditLogInvalid,
	AnnAMTAuditLogMissing,
	AnnAMTAuditLogSignatureInvalid,
	AnnAMTTypeInvalid,
	AnnAMTTypeMissing,
	AnnTCCSMENoUpgrade,
	AnnTCCSMEDowngrade,
	AnnTCPlatformNoUpgrade,
	AnnTCFirmwareSetChanged,
	AnnTCUEFIConfigChanged,
	AnnTCUEFIBootChanged,
	AnnTCUEFINoExit,
	AnnTCGPTChanged,
	AnnTCUEFIKeysChanged,
	AnnTCUEFISecureBootOff,
	AnnTCUEFIdbxRemoved,
	AnnTCShimMokListsChanged,
	AnnTCGrub,
	AnnTCBootFailed,
	AnnTCNoEventlog,
	AnnTCInvalidEventlog,
	AnnTCDummyTPM,
	AnnTCEndorsementCertUnverified,
	AnnTCBootAggregate,
	AnnTCRuntimeMeasurements,
	AnnTCInvalidIMAlog,
	AnnTCUEFIdbxIncomplete,
	AnnTCNotInLVFS,
	AnnTCCSMEVersionVuln,
	AnnBRLY2021033,
	AnnBRLYESPecter,
	AnnBRLY2021040,
	AnnBRLY2021013,
	AnnBRLY2022004,
	AnnBRLYMoonBounceCOREDXE,
	AnnBRLY2021004,
	AnnBRLY2021024,
	AnnBRLY2022027,
	AnnBRLY2022014,
	AnnBRLY2021009,
	AnnBRLY2021036,
	AnnBRLY2021011,
	AnnBRLYThinkPwn,
	AnnBRLY2021042,
	AnnBRLY2021003,
	AnnBRLY2021035,
	AnnBRLY2021053,
	AnnBRLY2021037,
	AnnBRLYIntelBSSADFTINTELSA00525,
	AnnBRLY2021045,
	AnnBRLY2022011,
	AnnBRLY2021005,
	AnnBRLY2021012,
	AnnBRLY2021022,
	AnnBRLYLojaxSecDxe,
	AnnBRLYUsbRtCVE20175721,
	AnnBRLY2021027,
	AnnBRLY2021029,
	AnnBRLY2021043,
	AnnBRLY2021051,
	AnnBRLY2021008,
	AnnBRLY2021007,
	AnnBRLY2021018,
	AnnBRLY2021038,
	AnnBRLYUsbRtSwSmiCVE202012301,
	AnnBRLY2021010,
	AnnBRLY2021020,
	AnnBRLY2021034,
	AnnBRLY2021039,
	AnnBRLYMosaicRegressor,
	AnnBRLY2021028,
	AnnBRLY2022013,
	AnnBRLY2021006,
	AnnBRLY2021026,
	AnnBRLY2021021,
	AnnBRLY2021023,
	AnnBRLY2021019,
	AnnBRLY2022016,
	AnnBRLYRkloader,
	AnnBRLYUsbRtUsbSmiCVE202012301,
	AnnBRLY2021017,
	AnnBRLY2021014,
	AnnBRLY2021047,
	AnnBRLY2022028RsbStuffing,
	AnnBRLY2021031,
	AnnBRLY2021050,
	AnnBRLY2021015,
	AnnBRLYUsbRtINTELSA00057,
	AnnBRLY2021016,
	AnnBRLY2022015,
	AnnBRLY2021041,
	AnnBRLY2022009,
	AnnBRLY2022010,
	AnnBRLY2022012,
	AnnBRLY2021046,
	AnnBRLY2021032,
	AnnBRLY2021030,
	AnnBRLY2021025,
	AnnBRLY2021001,
	AnnTCESETNotRunning,
	AnnTCESETDisabled,
	AnnTCESETExcluded,
	AnnTCESETManipulated,
	AnnTCUnsecureWindowsBoot,
	AnnTSCEKMismatch,
	AnnTSCPCRMismatch,
	AnnInternalNoTSC,
	AnnInternalNoBinarly,
	AnnInternalNoIMA,
	AnnInternalNoELAM,
	AnnInternalNoESETLinux,
}

func TestIsFatal(t *testing.T) {
	for _, ann := range allAnnotations {
		if strings.HasPrefix(string(ann), "brly-") {
			continue
		}
		_, ok := AnnFatal[AnnotationID(ann)]
		assert.Truef(t, ok, "%s in AnnFatal", ann)
	}
}

func TestCategory(t *testing.T) {
	for _, ann := range allAnnotations {
		if strings.HasPrefix(string(ann), "internal-") {
			continue
		}
		if strings.HasPrefix(string(ann), "me-") {
			continue
		}
		if strings.HasPrefix(string(ann), "amt-") {
			continue
		}
		if strings.HasPrefix(string(ann), "sgx-") {
			continue
		}
		if strings.HasPrefix(string(ann), "sev-") {
			continue
		}
		if strings.HasPrefix(string(ann), "tpm-") {
			continue
		}
		if strings.HasPrefix(string(ann), "uefi-") {
			continue
		}
		if strings.HasPrefix(string(ann), "txt-") {
			continue
		}
		if strings.HasPrefix(string(ann), "smbios-") {
			continue
		}
		if strings.HasPrefix(string(ann), "host-") {
			continue
		}
		assert.NotEmptyf(t, annotationCategory(AnnotationID(ann)), "%s has a category", ann)
	}
}
