package api

import (
	"github.com/google/go-tpm/tpm2"
	"github.com/google/uuid"
)

var CurrentAPIVersion = "2"

var (
	// 3.3 Globally Defined Variables
	EFIGlobalVariable = uuid.Must(uuid.FromBytes([]byte{
		0x8B, 0xE4, 0xDF, 0x61, 0x93, 0xCA, 0x11, 0xd2, 0xAA, 0x0D, 0x00, 0xE0, 0x98, 0x03, 0x2B, 0x8C}))
	EFIImageSecurityDatabase = uuid.Must(uuid.FromBytes([]byte{
		0xD7, 0x19, 0xB2, 0xCB, 0x3D, 0x3A, 0x45, 0x96, 0xA3, 0xBC, 0xDA, 0xD0, 0x0E, 0x67, 0x65, 0x6F}))

	CPUIDExtendedFeatureFlags uint32 = 0x7
	CPUIDSGXCapabilities      uint32 = 0x12
	CPUIDSEV                  uint32 = 0x8000001f

	MSRSMBase             uint32 = 0x9e
	MSRMTRRCap            uint32 = 0xfe
	MSRSMRRPhysBase       uint32 = 0x1F2
	MSRSMRRPhysMask       uint32 = 0x1F3
	MSRFeatureControl     uint32 = 0x3A
	MSRPlatformID         uint32 = 0x17
	MSRIA32DebugInterface uint32 = 0xC80
	MSRK8Sys              uint32 = 0xc0010010
	MSREFER               uint32 = 0xc0000080
)

// Databases with UUID "EFIImageSecurityDatabase"
var (
	EFIImageSecurityDatabases = map[string]bool{
		"db":  true,
		"dbx": true,
		"dbt": true,
		"dbr": true,
	}
)

var (
	RootECC = KeyTemplate{
		Public: PublicKey(tpm2.Public{
			Type:    tpm2.AlgECC,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagDecrypt,
			AuthPolicy: []byte{},
			ECCParameters: &tpm2.ECCParams{
				Symmetric: &tpm2.SymScheme{
					Alg:     tpm2.AlgAES,
					KeyBits: 128,
					Mode:    tpm2.AlgCFB,
				},
				CurveID: tpm2.CurveNISTP256,
			},
		}),
		Label: "IMMUNE-GUARD-ROOT-KEY-V2",
	}

	RootRSA = KeyTemplate{
		Public: PublicKey(tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagDecrypt,
			AuthPolicy: []byte{},
			RSAParameters: &tpm2.RSAParams{
				Symmetric: &tpm2.SymScheme{
					Alg:     tpm2.AlgAES,
					KeyBits: 128,
					Mode:    tpm2.AlgCFB,
				},
				KeyBits:     2048,
				ExponentRaw: 0,
				ModulusRaw:  make([]byte, 256),
			},
		}),
		Label: "IMMUNE-GUARD-ROOT-KEY-V2",
	}

	QuoteECC = KeyTemplate{
		Public: PublicKey(tpm2.Public{
			Type:    tpm2.AlgECC,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagSign,
			AuthPolicy: []byte{},
			ECCParameters: &tpm2.ECCParams{
				Sign: &tpm2.SigScheme{
					Alg:   tpm2.AlgECDSA,
					Hash:  tpm2.AlgSHA256,
					Count: 0,
				},
				CurveID: tpm2.CurveNISTP256,
			},
		}),
		Label: "IMMUNE-GUARD-AIK-V2",
	}

	QuoteRSA = KeyTemplate{
		Public: PublicKey(tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagSign,
			AuthPolicy: []byte{},
			RSAParameters: &tpm2.RSAParams{
				Sign: &tpm2.SigScheme{
					Alg:  tpm2.AlgRSAPSS,
					Hash: tpm2.AlgSHA256,
				},
				KeyBits:     2048,
				ExponentRaw: 0,
				ModulusRaw:  make([]byte, 256),
			},
		}),
		Label: "IMMUNE-GUARD-AIK-V2",
	}
)
