package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"math/rand"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

func deviceKeyCreationDataHash(root tpm2.Public, key tpm2.Public, outsideInfo string) ([]byte, error) {
	rootQname, err := ComputeName(tpm2.HandleEndorsement, PublicKey(root))
	if err != nil {
		return []byte{}, err
	}

	rootName, err := ComputeName(PublicKey(root))
	if err != nil {
		return []byte{}, err
	}

	// encode expected TPMS_CREATION_DATA
	nameAlgHash, err := key.NameAlg.Hash()
	if err != nil {
		return []byte{}, err
	}

	expectedCreationData := tpm2.CreationData{
		PCRSelection: tpm2.PCRSelection{},
		// in part 2, table 212 it says "[..] pcrDigest.size shall be zero if
		// the pcrSelect list is empty." but the code in part 4 8.7.3.22 sets
		// pcrDigest to the hash of nothing.
		PCRDigest:           nameAlgHash.New().Sum([]byte{}),
		Locality:            1, // XXX: only applies to swtpm
		ParentNameAlg:       root.NameAlg,
		ParentName:          tpm2.Name(rootName),
		ParentQualifiedName: tpm2.Name(rootQname),
		OutsideInfo:         []byte(outsideInfo),
	}

	creationDataBlob, err := expectedCreationData.EncodeCreationData()
	if err != nil {
		return []byte{}, err
	}

	creationDataHasher := nameAlgHash.New()
	creationDataHasher.Write(creationDataBlob)

	return creationDataHasher.Sum([]byte{}), nil
}

func GenerateSignatureECC(rand *rand.Rand) tpm2.SignatureECC {
	return tpm2.SignatureECC{
		HashAlg: tpm2.AlgSHA256,
		R:       test.GenerateBigInt(rand, 256),
		S:       test.GenerateBigInt(rand, 256),
	}
}

func GenerateSignatureRSA(rand *rand.Rand) tpm2.SignatureRSA {
	return tpm2.SignatureRSA{
		HashAlg:   tpm2.AlgSHA256,
		Signature: test.GenerateBytes(rand, 256, 512),
	}
}

func GenerateSignature(rand *rand.Rand) tpm2.Signature {
	if rand.Intn(2) == 0 {
		ecc := GenerateSignatureECC(rand)
		return tpm2.Signature{
			Alg: tpm2.AlgECDSA,
			RSA: nil,
			ECC: &ecc,
		}
	} else {
		rsa := GenerateSignatureRSA(rand)
		return tpm2.Signature{
			Alg: tpm2.AlgRSAPSS,
			RSA: &rsa,
			ECC: nil,
		}
	}
}

func GenerateCertifyInfo(rand *rand.Rand) tpm2.CertifyInfo {
	return tpm2.CertifyInfo{
		Name:          GenerateName(rand),
		QualifiedName: GenerateName(rand),
	}
}

func GenerateCreationInfo(rand *rand.Rand) tpm2.CreationInfo {
	return tpm2.CreationInfo{
		Name:         GenerateName(rand),
		OpaqueDigest: tpmutil.U16Bytes(test.GenerateBytes(rand, 20, 21)),
	}
}

func GenerateQuoteInfo(rand *rand.Rand) tpm2.QuoteInfo {
	pcrsel := []tpm2.PCRSelection{
		{
			Hash: tpm2.AlgSHA256,
			PCRs: []int{1, 2, 3},
		}}
	signingHasher := sha256.New()
	for range pcrsel {
		signingHasher.Write(nil)
	}
	return tpm2.QuoteInfo{
		PCRSelection: pcrsel,
		PCRDigest:    signingHasher.Sum([]byte{}),
	}
}

func GenerateClockInfo(rand *rand.Rand) tpm2.ClockInfo {
	return tpm2.ClockInfo{
		Clock:        rand.Uint64(),
		ResetCount:   22,
		RestartCount: 11,
		Safe:         1,
	}
}

func GenerateAttestationCreation(rand *rand.Rand) tpm2.AttestationData {
	creation := GenerateCreationInfo(rand)
	return tpm2.AttestationData{
		Magic:                0xff544347,
		Type:                 tpm2.TagAttestCreation,
		QualifiedSigner:      GenerateName(rand),
		ExtraData:            tpmutil.U16Bytes(test.GenerateBytes(rand, 10, 20)),
		ClockInfo:            GenerateClockInfo(rand),
		FirmwareVersion:      42,
		AttestedCertifyInfo:  nil,
		AttestedQuoteInfo:    nil,
		AttestedCreationInfo: &creation,
	}
}

func GenerateAttestationData(rand *rand.Rand) tpm2.AttestationData {
	certify := GenerateCertifyInfo(rand)
	return tpm2.AttestationData{
		Magic:                0xff544347,
		Type:                 tpm2.TagAttestCertify,
		QualifiedSigner:      GenerateName(rand),
		ExtraData:            tpmutil.U16Bytes(test.GenerateBytes(rand, 10, 20)),
		ClockInfo:            GenerateClockInfo(rand),
		FirmwareVersion:      42,
		AttestedCertifyInfo:  &certify,
		AttestedQuoteInfo:    nil,
		AttestedCreationInfo: nil,
	}
}

//func (Attest) Generate(rand *rand.Rand, size int) reflect.Value {
//	return reflect.ValueOf(Attest(GenerateAttestationData(rand)))
//}

func GenerateName(rand *rand.Rand) tpm2.Name {
	hash := tpm2.HashValue{
		Alg:   tpm2.AlgSHA1,
		Value: tpmutil.U16Bytes(test.GenerateBytes(rand, 20, 21)),
	}
	nam := tpm2.Name{
		Digest: &hash,
		Handle: nil,
	}

	return nam
}

//func (Name) Generate(rand *rand.Rand, size int) reflect.Value {
//	return reflect.ValueOf(Name(GenerateName(rand)))
//}

func GenerateSymScheme(rand *rand.Rand) tpm2.SymScheme {
	modes := []tpm2.Algorithm{tpm2.AlgCTR, tpm2.AlgOFB, tpm2.AlgCBC, tpm2.AlgCFB, tpm2.AlgECB}
	bits := []uint16{128, 192, 256}

	return tpm2.SymScheme{
		Alg:     tpm2.AlgAES,
		KeyBits: bits[rand.Intn(len(bits))],
		Mode:    modes[rand.Intn(len(modes))],
	}
}

func GenerateSigScheme(rand *rand.Rand, ecc bool) tpm2.SigScheme {
	hashes := []tpm2.Algorithm{tpm2.AlgSHA256, tpm2.AlgSHA384, tpm2.AlgSHA512}
	eccs := []tpm2.Algorithm{tpm2.AlgECDSA, tpm2.AlgECDAA}
	rsas := []tpm2.Algorithm{tpm2.AlgRSAPSS, tpm2.AlgRSASSA}
	var alg tpm2.Algorithm

	if ecc {
		alg = eccs[rand.Intn(len(eccs))]
	} else {
		alg = rsas[rand.Intn(len(rsas))]
	}

	var count uint32

	if alg.UsesCount() {
		count = rand.Uint32()
	}

	return tpm2.SigScheme{
		Alg:   alg,
		Hash:  hashes[rand.Intn(len(hashes))],
		Count: count,
	}
}

func GenerateParamsRSA(rand *rand.Rand, prop tpm2.KeyProp) tpm2.RSAParams {
	var sym *tpm2.SymScheme
	var sig *tpm2.SigScheme

	bits := []uint16{2048, 4096}
	mod := make([]byte, bits[rand.Intn(len(bits))]/8)
	rand.Read(mod)

	if prop&tpm2.FlagDecrypt != 0 {
		s := GenerateSymScheme(rand)
		sym = &s
	} else {
		sym = nil
	}

	if prop&tpm2.FlagSign != 0 {
		s := GenerateSigScheme(rand, false)
		sig = &s
	} else {
		sig = nil
	}

	return tpm2.RSAParams{
		Symmetric:   sym,
		Sign:        sig,
		KeyBits:     uint16(len(mod) * 8),
		ExponentRaw: 0x10001,
		ModulusRaw:  mod,
	}
}

func GenerateECPoint(rand *rand.Rand, cv elliptic.Curve, size uint) tpm2.ECPoint {
	pub, err := ecdsa.GenerateKey(cv, rand)
	if err != nil {
		panic(err)
	}

	return tpm2.ECPoint{
		XRaw: pub.X.Bytes(),
		YRaw: pub.Y.Bytes(),
	}
}

func GenerateKDFScheme(rand *rand.Rand) tpm2.KDFScheme {
	hashes := []tpm2.Algorithm{tpm2.AlgSHA256, tpm2.AlgSHA384, tpm2.AlgSHA512}

	return tpm2.KDFScheme{
		Alg:  tpm2.AlgKDF2,
		Hash: hashes[rand.Intn(len(hashes))],
	}
}

func GenerateParamsECC(rand *rand.Rand, prop tpm2.KeyProp) tpm2.ECCParams {
	var sym *tpm2.SymScheme
	var sig *tpm2.SigScheme
	var kdf *tpm2.KDFScheme

	curves := []tpm2.EllipticCurve{tpm2.CurveNISTP256}
	pub := GenerateECPoint(rand, elliptic.P256(), 32)

	if prop&tpm2.FlagDecrypt != 0 {
		s := GenerateSymScheme(rand)
		s = tpm2.SymScheme{
			Alg:     tpm2.AlgAES,
			KeyBits: 128,
			Mode:    tpm2.AlgCFB,
		}
		sym = &s
		//k := GenerateKDFScheme(rand)
		kdf = nil //&k
	} else {
		sym = nil
	}

	if prop&tpm2.FlagSign != 0 {
		//s := GenerateSigScheme(rand, true)
		s := tpm2.SigScheme{
			Alg:  tpm2.AlgECDSA,
			Hash: tpm2.AlgSHA256,
		}
		sig = &s
	} else {
		sig = nil
	}

	return tpm2.ECCParams{
		Symmetric: sym,
		Sign:      sig,
		CurveID:   curves[rand.Intn(len(curves))],
		KDF:       kdf,
		Point:     pub,
	}
}

func GeneratePublicRSA(rand *rand.Rand) tpm2.Public {
	var rsa *tpm2.RSAParams
	var ecc *tpm2.ECCParams
	var ty tpm2.Algorithm

	hashes := []tpm2.Algorithm{tpm2.AlgSHA256, tpm2.AlgSHA384, tpm2.AlgSHA512}
	prop := tpm2.KeyProp(rand.Uint32())
	policy := make([]byte, 32)
	rand.Read(policy)

	s := GenerateParamsRSA(rand, prop)
	rsa = &s
	ty = tpm2.AlgRSA

	return tpm2.Public{
		Type:                ty,
		NameAlg:             hashes[rand.Intn(len(hashes))],
		AuthPolicy:          policy,
		Attributes:          prop,
		RSAParameters:       rsa,
		ECCParameters:       ecc,
		SymCipherParameters: nil,
		KeyedHashParameters: nil,
	}
}

func GeneratePublicECC(rand *rand.Rand) tpm2.Public {
	var rsa *tpm2.RSAParams
	var ecc *tpm2.ECCParams
	var ty tpm2.Algorithm

	hashes := []tpm2.Algorithm{tpm2.AlgSHA256, tpm2.AlgSHA384, tpm2.AlgSHA512}
	prop := tpm2.KeyProp(rand.Uint32())
	policy := make([]byte, 32)
	rand.Read(policy)

	s := GenerateParamsECC(rand, prop)
	ecc = &s
	ty = tpm2.AlgECC

	return tpm2.Public{
		Type:                ty,
		NameAlg:             hashes[rand.Intn(len(hashes))],
		AuthPolicy:          policy,
		Attributes:          prop,
		RSAParameters:       rsa,
		ECCParameters:       ecc,
		SymCipherParameters: nil,
		KeyedHashParameters: nil,
	}
}

func GeneratePublic(rand *rand.Rand) tpm2.Public {
	if rand.Uint32()&1 != 0 {
		return GeneratePublicRSA(rand)
	} else {
		return GeneratePublicECC(rand)
	}
}

func GenerateEK(rand *rand.Rand) tpm2.Public {
	hashes := []tpm2.Algorithm{tpm2.AlgSHA256, tpm2.AlgSHA384, tpm2.AlgSHA512}
	attr := tpm2.FlagDecrypt | tpm2.FlagFixedTPM
	policy := make([]byte, 32)
	rand.Read(policy)

	s := GenerateParamsRSA(rand, attr)

	return tpm2.Public{
		Type:          tpm2.AlgRSA,
		NameAlg:       hashes[rand.Intn(len(hashes))],
		AuthPolicy:    policy,
		Attributes:    attr,
		RSAParameters: &s,
	}
}

func GenerateIdentityKey(rand *rand.Rand) (tpm2.Public, tpm2.AttestationData) {
	attr := tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
		tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagDecrypt
	policy := make([]byte, 32)
	rand.Read(policy)

	s := GenerateParamsECC(rand, attr)

	pub := tpm2.Public{
		Type:          tpm2.AlgECC,
		NameAlg:       tpm2.AlgSHA256,
		AuthPolicy:    policy,
		Attributes:    attr,
		ECCParameters: &s,
	}
	name, err := pub.Name()
	if err != nil {
		panic(err)
	}
	signer, err := ComputeName(tpm2.HandleEndorsement, PublicKey(pub))
	if err != nil {
		panic(err)
	}

	cert := tpm2.CreationInfo{
		Name:         name,
		OpaqueDigest: tpmutil.U16Bytes(test.GenerateBytes(rand, 32, 64)),
	}
	attest := tpm2.AttestationData{
		Magic:                0xff544347,
		Type:                 tpm2.TagAttestCreation,
		QualifiedSigner:      tpm2.Name(signer),
		ExtraData:            tpmutil.U16Bytes([]byte{}),
		ClockInfo:            GenerateClockInfo(rand),
		FirmwareVersion:      42,
		AttestedCreationInfo: &cert,
	}

	return pub, attest
}

//func (PublicKey) Generate(rand *rand.Rand, size int) reflect.Value {
//	return reflect.ValueOf(PublicKey(GeneratePublic(rand)))
//}

func GenerateCertificateRSA(rand *rand.Rand) x509.Certificate {
	keyRef, err := rsa.GenerateKey(rand, 2048)
	if err != nil {
		panic("Failed to generate key")
	}

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Random Cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		PublicKeyAlgorithm:    x509.RSA,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand, &tmpl, &tmpl, &keyRef.PublicKey, keyRef)
	if err != nil {
		panic("Failed to sign certificate")
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic("Failed to parse certificate")
	}

	return *cert
}

func GenerateCertificateECC(rand *rand.Rand) x509.Certificate {
	keyRef, err := ecdsa.GenerateKey(elliptic.P256(), rand)
	if err != nil {
		panic("Failed to generate key")
	}

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Random Cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		PublicKeyAlgorithm:    x509.ECDSA,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand, &tmpl, &tmpl, &keyRef.PublicKey, keyRef)
	if err != nil {
		panic("Failed to sign certificate")
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic("Failed to parse certificate")
	}

	return *cert
}

func GenerateCertificate(rand *rand.Rand) x509.Certificate {
	return GenerateCertificateECC(rand)
}

//func (Certificate) Generate(rand *rand.Rand, size int) reflect.Value {
//	return reflect.ValueOf(Certificate(GenerateCertificate(rand)))
//}

func GenerateDeviceKey(root tpm2.Public, rand *rand.Rand) (tpm2.Public, tpm2.AttestationData, tpm2.Signature, *ecdsa.PrivateKey) {
	attr := tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
		tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagSign
	policy := make([]byte, 32)
	rand.Read(policy)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand)
	if err != nil {
		panic(err)
	}

	s := tpm2.ECCParams{
		Symmetric: nil,
		Sign: &tpm2.SigScheme{
			Alg:  tpm2.AlgECDSA,
			Hash: tpm2.AlgSHA256,
		},
		CurveID: tpm2.CurveNISTP256,
		KDF:     nil,
		Point: tpm2.ECPoint{
			XRaw: key.X.Bytes(),
			YRaw: key.Y.Bytes(),
		},
	}

	pub := tpm2.Public{
		Type:          tpm2.AlgECC,
		NameAlg:       tpm2.AlgSHA256,
		AuthPolicy:    policy,
		Attributes:    attr,
		ECCParameters: &s,
	}
	name, err := pub.Name()
	if err != nil {
		panic(err)
	}
	signer, err := ComputeName(tpm2.HandleEndorsement, PublicKey(root), PublicKey(pub))
	if err != nil {
		panic(err)
	}
	od, err := deviceKeyCreationDataHash(root, pub, "IMMUNE-GUARD-AIK-V2")
	if err != nil {
		panic(err)
	}
	cert := tpm2.CreationInfo{
		Name:         name,
		OpaqueDigest: od,
	}
	attest := tpm2.AttestationData{
		Magic:                0xff544347,
		Type:                 tpm2.TagAttestCreation,
		QualifiedSigner:      tpm2.Name(signer),
		ExtraData:            nil,
		ClockInfo:            GenerateClockInfo(rand),
		FirmwareVersion:      42,
		AttestedCreationInfo: &cert,
	}

	attestBlob, err := attest.Encode()
	if err != nil {
		panic(err)
	}

	nameAlgHash, err := pub.NameAlg.Hash()
	if err != nil {
		panic(err)
	}

	attestHasher := nameAlgHash.New()
	attestHasher.Write(attestBlob)
	attestHash := attestHasher.Sum([]byte{})

	r, ss, err := ecdsa.Sign(rand, key, attestHash)
	if err != nil {
		panic(err)
	}

	sig := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{
			HashAlg: tpm2.AlgSHA256,
			R:       r,
			S:       ss,
		},
	}

	return pub, attest, sig, key
}

func GenerateQuote(root tpm2.Public, pub tpm2.Public, key *ecdsa.PrivateKey, extraData tpmutil.U16Bytes, rand *rand.Rand) (tpm2.AttestationData, tpm2.Signature) {
	nam, err := ComputeName(tpm2.HandleEndorsement, PublicKey(root), PublicKey(pub))
	if err != nil {
		panic(err)
	}

	quote := GenerateQuoteInfo(rand)
	att := tpm2.AttestationData{
		Magic:                0xff544347,
		Type:                 tpm2.TagAttestQuote,
		QualifiedSigner:      tpm2.Name(nam),
		ClockInfo:            GenerateClockInfo(rand),
		FirmwareVersion:      42,
		ExtraData:            extraData,
		AttestedCertifyInfo:  nil,
		AttestedQuoteInfo:    &quote,
		AttestedCreationInfo: nil,
	}
	attestBlob, err := att.Encode()
	if err != nil {
		panic(err)
	}

	nameAlgHash, err := pub.NameAlg.Hash()
	if err != nil {
		panic(err)
	}

	quoteHasher := nameAlgHash.New()
	quoteHasher.Write(attestBlob)
	quoteHash := quoteHasher.Sum([]byte{})

	r, ss, err := ecdsa.Sign(rand, key, quoteHash)
	if err != nil {
		panic(err)
	}

	sig := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{
			HashAlg: tpm2.AlgSHA256,
			R:       r,
			S:       ss,
		},
	}

	return att, sig
}
