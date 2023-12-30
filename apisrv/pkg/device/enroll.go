package device

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/go-tpm/tpm2"
	google "github.com/google/go-tpm/tpm2/credactivation"
	"github.com/jackc/pgx/v4"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	cred "github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/configuration"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

func randomBuffer(bytes int) ([]byte, error) {
	var value []byte = make([]byte, bytes)

	l, err := rand.Read(value)
	if l != bytes || err != nil {
		return []byte{}, err
	}

	return value, nil
}

func makeCredential(key *api.PublicKey, ek *rsa.PublicKey, ekNameAlg tpm2.Algorithm, secret []byte) ([]byte, []byte, []byte, []byte, error) {

	credkey, err := randomBuffer(16)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, []byte{}, err
	}

	nonce, err := randomBuffer(12)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, []byte{}, err
	}

	block, err := aes.NewCipher(credkey)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, []byte{}, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, []byte{}, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, secret, nil)

	name, err := key.Name()
	if err != nil {
		return []byte{}, []byte{}, []byte{}, []byte{}, err
	}

	// TPM2_ActivateCredential
	// 		activateHandle = Device key (AIK)
	//  	keyHandle = EK (protector)
	//  	creationData = encrypted X.509 certificate
	//  	secret = EK-encrypted seed for creationData
	// Generate
	//		aik = Device key (AIK)
	//		pub = EK (protector)
	//		pubNameAlg = EK Name algorithm
	//		symBlockSize
	//		secret = unencrypted secret
	encCred, encSecret, err := google.Generate(name.Digest, ek, 16, credkey)

	if err != nil {
		return []byte{}, []byte{}, []byte{}, []byte{}, err
	} else {
		return encCred[2:], encSecret, ciphertext, nonce, err
	}
}

func verifyKeyAgainstTemplate(key *api.PublicKey, template *api.KeyTemplate) error {
	if key.Type != template.Public.Type {
		return fmt.Errorf("key type")
	}

	if key.NameAlg != template.Public.NameAlg {
		return fmt.Errorf("name algorithm")
	}

	if key.Attributes != template.Public.Attributes {
		return fmt.Errorf("key attribute flags")
	}

	if key.ECCParameters != nil && template.Public.ECCParameters != nil {
		keyParams := key.ECCParameters
		tmplParams := template.Public.ECCParameters

		if keyParams.CurveID != tmplParams.CurveID {
			return fmt.Errorf("curve ID")
		}

		if !reflect.DeepEqual(keyParams.Sign, tmplParams.Sign) {
			return fmt.Errorf("signing scheme")
		}

		if !reflect.DeepEqual(keyParams.Symmetric, tmplParams.Symmetric) {
			return fmt.Errorf("symmetric scheme")
		}

		if !reflect.DeepEqual(keyParams.KDF, tmplParams.KDF) {
			return fmt.Errorf("KDF")
		}

		// XXX: check if point is on the curve
	} else if key.RSAParameters != nil && template.Public.RSAParameters != nil {
		keyParams := key.RSAParameters
		tmplParams := template.Public.RSAParameters

		if !reflect.DeepEqual(keyParams.Sign, tmplParams.Sign) {
			return fmt.Errorf("signing scheme")
		}

		if !reflect.DeepEqual(keyParams.Symmetric, tmplParams.Symmetric) {
			return fmt.Errorf("symmetric scheme")
		}

		if keyParams.KeyBits != tmplParams.KeyBits {
			return fmt.Errorf("key length")
		}
	}

	return nil
}

func certifyKey(ca *ecdsa.PrivateKey, caKid string, ek *rsa.PublicKey, rootQN *api.Name, key *api.Key, keyName string, serviceName string, now time.Time) (*api.Name, string, *api.EncryptedCredential, error) {
	qname, err := api.ComputeName(rootQN, key.Public)
	if err != nil {
		return nil, "", nil, err
	}

	plainCert, err := cred.IssueDeviceCredential(serviceName, qname, now, caKid, ca)
	if err != nil {
		return nil, "", nil, err
	}

	encKek, encSec, encCert, nonce, err :=
		makeCredential(&key.Public, ek, 0 /* ununsed downstream. bug? */, []byte(plainCert))
	if err != nil {
		return nil, "", nil, err
	}

	enccred := api.EncryptedCredential{
		Name:       keyName,
		KeyID:      api.Buffer(encKek),
		Credential: api.Buffer(encCert),
		Secret:     api.Buffer(encSec),
		Nonce:      api.Buffer(nonce),
	}

	return &qname, plainCert, &enccred, nil
}

func creationDataHash(rootName *api.Name, rootQN *api.Name, rootNameAlg tpm2.Algorithm, keyNameAlg tpm2.Algorithm, outsideInfo string) ([]byte, error) {
	// encode expected TPMS_CREATION_DATA
	nameAlgHash, err := keyNameAlg.Hash()
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
		ParentNameAlg:       rootNameAlg,
		ParentName:          tpm2.Name(*rootName),
		ParentQualifiedName: tpm2.Name(*rootQN),
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

func verifyAttestationData(keyQN *api.Name, keyName *api.Name, keyPub crypto.PublicKey, keyNameHash crypto.Hash, expectedCreationDataHash []byte, attestationData *api.Attest, attestationSignature *api.Signature) error {
	// check attestationData
	if attestationData.Magic != 0xFF544347 {
		return fmt.Errorf("device key creation data has the wrong magic number: %#v", attestationData.Magic)
	}

	if attestationData.Type != tpm2.TagAttestCreation {
		return fmt.Errorf("device key creation data is of the wrong type: %#v", attestationData.Type)
	}

	// attestation is self-signed
	isSha256 := attestationData.QualifiedSigner.Digest.Alg == tpm2.AlgSHA256
	namesMatch := api.EqualNames((*api.Name)(&attestationData.QualifiedSigner), keyQN)
	if !isSha256 || !namesMatch {
		return fmt.Errorf("device key creation data has wrong signer. Got %#v, expected %#v", attestationData.QualifiedSigner.Digest, keyQN.Digest)
	}

	if attestationData.ExtraData != nil {
		return fmt.Errorf("device key creation data has wrong extra data: %#v", attestationData.ExtraData)
	}

	// XXX: ClockInfo
	// XXX: FirmwareVersion

	// check creationInfo
	info := attestationData.AttestedCreationInfo
	if info == nil {
		return fmt.Errorf("device key creation data is malformed (no creationInfo)")
	}

	if !api.EqualNames((*api.Name)(&info.Name), keyName) {
		return fmt.Errorf("device key creation data has wrong subject: %#v", info.Name)
	}

	if !bytes.Equal([]byte(info.OpaqueDigest), expectedCreationDataHash) {
		return fmt.Errorf("wrong creation data hash: %#v != %#v", info.OpaqueDigest, expectedCreationDataHash)
	}

	attestationDataBlob, err := tpm2.AttestationData(*attestationData).Encode()
	if err != nil {
		return err
	}

	attestationHasher := keyNameHash.New()
	attestationHasher.Write(attestationDataBlob)
	attestationHash := attestationHasher.Sum([]byte{})

	var sigValid bool = false
	if attestationSignature.ECC != nil {
		ec, ok := (keyPub).(*ecdsa.PublicKey)
		if !ok || ec.Curve != elliptic.P256() {
			return errors.New("quote key is not a ECDSA key over NIST P-256")
		}
		sigValid = ecdsa.Verify(ec, attestationHash, attestationSignature.ECC.R, attestationSignature.ECC.S)
	} else if attestationSignature.RSA != nil {
		pssOpt := rsa.PSSOptions{Hash: crypto.SHA256, SaltLength: rsa.PSSSaltLengthAuto}
		sigValid = rsa.VerifyPSS(keyPub.(*rsa.PublicKey), crypto.SHA256, attestationHash, attestationSignature.RSA.Signature, &pssOpt) == nil
	} else {
		return errors.New("invalid sig alg")
	}

	if !sigValid {
		return fmt.Errorf("attestation signature invalid")
	} else {
		return nil
	}
}

func extractEK(ek *api.PublicKey, ekCert *api.Certificate) (*rsa.PublicKey, *api.Name, error) {
	// check EK
	ekPub, err := tpm2.Public(*ek).Key()
	if err != nil {
		return nil, nil, err
	}

	// compute qname
	ekQN, err := api.ComputeName(tpm2.HandleEndorsement, ek)
	if err != nil {
		return nil, nil, err
	}

	// convert to usable RSA key
	if ek, ok := ekPub.(*rsa.PublicKey); ok {
		return ek, &ekQN, nil
	} else {
		return nil, nil, fmt.Errorf("no a rsa ek")
	}
}

func Enroll(ctx context.Context, tx pgx.Tx, ca *ecdsa.PrivateKey, caKid string, req api.Enrollment, serviceName string, org string, actor string, now time.Time) (int64, []*api.EncryptedCredential, error) {
	// verify root against template
	err := verifyKeyAgainstTemplate(&req.Root, &configuration.DefaultConfiguration.Root)
	if err != nil {
		return 0, nil, err
	}

	// Root Name and QN
	rootQN, err := api.ComputeName(tpm2.HandleEndorsement, req.Root)
	if err != nil {
		return 0, nil, err
	}
	rootName, err := api.ComputeName(req.Root)
	if err != nil {
		return 0, nil, err
	}

	// verify the EK and it's certificate
	ek, ekQN, err := extractEK(&req.EndorsementKey, req.EndorsementCertificate)
	if err != nil {
		return 0, nil, err
	}

	// verify & certify device keys
	keyrows := make(map[string]KeysRow, len(req.Keys))
	var enccred []*api.EncryptedCredential
	for i, k := range req.Keys {
		if tmpl, ok := configuration.DefaultConfiguration.Keys[i]; ok {
			// check against template
			err := verifyKeyAgainstTemplate(&k.Public, &tmpl)
			if err != nil {
				return 0, nil, err
			}

			// public key in a usable format
			keyPub, err := tpm2.Public(k.Public).Key()
			if err != nil {
				return 0, nil, err
			}
			keyNameHash, err := k.Public.NameAlg.Hash()
			if err != nil {
				return 0, nil, err
			}

			// key Name and QN
			keyQN, err := api.ComputeName(rootQN, k.Public)
			if err != nil {
				return 0, nil, err
			}
			keyName, err := api.ComputeName(k.Public)
			if err != nil {
				return 0, nil, err
			}

			// verify creation proof
			cdh, err := creationDataHash(&rootName, &rootQN, req.Root.NameAlg, k.Public.NameAlg, tmpl.Label)
			if err != nil {
				return 0, nil, err
			}
			err = verifyAttestationData(&keyQN, &keyName, keyPub, keyNameHash, cdh, &k.CreationProof, &k.CreationProofSignature)
			if err != nil {
				return 0, nil, err
			}
		} else {
			return 0, nil, fmt.Errorf("unknown key %s", i)
		}

		// make key credential
		qname, cred, ec, err := certifyKey(ca, caKid, ek, &rootQN, &k, i, serviceName, now)
		if err != nil {
			return 0, nil, err
		}
		enccred = append(enccred, ec)
		keyrows[i] = KeysRow{
			Name:       i,
			Public:     k.Public,
			QName:      *qname,
			Credential: cred,
			DeviceId:   -1, // set later
		}
	}

	// create new baseline and initially populate some fields
	bline := baseline.New()
	bline.EndorsementCertificate = req.EndorsementCertificate

	// create a new policy
	pol := policy.New()

	// insert device row
	devrow, err := Insert(ctx, tx, ekQN, &rootQN, bline, pol, req.NameHint, org, now, actor)
	if err != nil {
		return 0, nil, database.Error(err)
	}

	// insert key rows
	var keys [][]interface{}
	for i := range keyrows {
		k := keyrows[i]
		keys = append(keys, []interface{}{
			k.Name, k.Public, k.QName, k.Credential, devrow.Id,
		})
		keyrows[i] = k
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"v2", "keys"},
		[]string{"name", "public", "fpr", "credential", "device_id"},
		pgx.CopyFromRows(keys))
	if err != nil {
		return 0, nil, database.Error(err)
	}

	// device changes
	var changes []int64 // this is just a bogus var to store results that are never used
	err = pgxscan.Get(ctx, tx, &changes, `
			--
			-- Enroll-Fetch-Changes
			--
			WITH existing_changes AS (
				SELECT * FROM v2.changes
				WHERE v2.changes.device_id = $1

			), dev_keys AS (
				SELECT id FROM v2.keys
				WHERE v2.keys.device_id = $1

			), new AS (
				INSERT INTO v2.changes (actor, type, timestamp, device_id, organization_id, key_id)
				SELECT $2, 'enroll'::v2.changes_type, $3, $1, $4, dev_keys.id
				FROM dev_keys
				RETURNING *

			)
			SELECT array_agg(id) FROM (
				SELECT * FROM new UNION SELECT * FROM existing_changes
			) AS t
			GROUP BY t.timestamp
			ORDER BY t.timestamp DESC
			LIMIT $5
		`, devrow.Id, actor, now, devrow.OrganizationId, 10)
	if err != nil {
		return 0, nil, database.Error(err)
	}

	return devrow.Id, enccred, err
}

func Insert(ctx context.Context, tx pgx.Tx, hwid *api.Name, fpr *api.Name, bline *baseline.Values, pol *policy.Values, nameHint string, org string, now time.Time, actor string) (*Row, error) {
	baselineJson, err := baseline.ToRow(bline)
	if err != nil {
		return nil, err
	}
	polcol, err := pol.Serialize()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		--
		-- Enroll-Device-Part-1
		--

		CREATE TEMPORARY TABLE org ON COMMIT DROP AS
		WITH
		-- create new org if needed
		existing AS (
			SELECT id AS org_id
			FROM v2.organizations
			WHERE external = $1
			LIMIT 1
		), created AS (
			INSERT INTO v2.organizations (external, devices, features, updated_at)
			SELECT $1, 10, ARRAY[]::v2.organizations_feature[], 'NOW'::timestamp
			WHERE NOT EXISTS (SELECT * FROM existing)
			RETURNING id AS org_id
		)
		SELECT org_id FROM existing UNION SELECT org_id FROM created
	`, org)
	if err != nil {
		return nil, database.Error(err)
	}

	_, err = tx.Exec(ctx, `
		--
		-- Enroll-Device-Part-2
		--
		
		CREATE TEMPORARY TABLE retired_devices ON COMMIT DROP AS
		WITH
		-- retire all devices with the same hwid. this must be a sepearate query
		-- otherwise the unique constraint for hwid is violated when we insert a new
		-- device. this is because all queries in a CTE work on the same snapshot and
		-- thus don't "see" changes by the previous queries.
		retired_devices AS (
		UPDATE v2.devices SET retired = TRUE
		WHERE hwid = $1 AND organization_id = any(SELECT * FROM org) AND retired = FALSE
			RETURNING id
		)
		SELECT id AS device_id
		FROM retired_devices
	`, hwid)
	if err != nil {
		return nil, database.Error(err)
	}

	var row Row
	err = pgxscan.Get(ctx, tx, &row, `
		--
		-- Enroll-Device-Part-3
		--

		WITH 
		-- convert the name hint to a free name by appending a counter.
		name_exact_matches AS (
			SELECT count(*) AS name_exact_matches
			FROM v2.devices
			WHERE name = $1 AND organization_id = any(SELECT * FROM org)

		), name_prefix_len AS (
			SELECT char_length($1) + 1 AS name_prefix_len

		), name_next_counter AS (
			SELECT COALESCE(MAX(substring(name, name_prefix_len)::integer) + 1,2) AS name_next_counter
			FROM v2.devices, name_prefix_len
			WHERE name LIKE ($1 || '%') AND name SIMILAR TO ($1 || ' [0-9]+')

		),
		-- insert the new row into devices. picks an unused name based on the expressions above
		new_devices AS (
			INSERT INTO v2.devices (hwid, fpr, retired, organization_id, replaced_by, baseline, policy, name)
			SELECT $2, $3, FALSE, (SELECT * FROM org), NULL, $4, $5,
				CASE WHEN name_exact_matches = 0 THEN $1
				ELSE $1 || ' ' || name_next_counter::text
				END
			FROM name_next_counter, name_exact_matches
			RETURNING *

		),
		-- set the already retired devices to be replaced by the new device
		t AS (
			UPDATE v2.devices SET replaced_by = (SELECT id from new_devices)
			WHERE v2.devices.id IN (SELECT device_id FROM retired_devices)

		)
		SELECT 
			id,
			hwid,
			fpr,
			name,
			cookie,
			retired,
			organization_id,
			replaced_by
		FROM new_devices
	`, nameHint, hwid, fpr, baselineJson, polcol)
	if err != nil {
		return nil, database.Error(err)
	}

	return &row, nil
}

// Testing only
func InsertKey(ctx context.Context, tx pgx.Tx, device string, public *api.PublicKey, name string, qname *api.Name, credential string) (*KeysRow, error) {
	var row KeysRow
	err := pgxscan.Get(ctx, tx, &row, `
		INSERT INTO v2.keys VALUES (
			DEFAULT,
			$1,
			$2,
			$3,
			$4,
			$5
		) RETURNING *
	`, public, name, qname, credential, device)
	if err != nil {
		return nil, err
	} else {
		return &row, nil
	}
}
