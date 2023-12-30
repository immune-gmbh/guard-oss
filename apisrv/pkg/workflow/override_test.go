package workflow

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/appraisal"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/binarly"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/inteltsc"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestOverride(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)
	minio := mock.MinioContainer(t, ctx)
	defer minio.Terminate(ctx)

	rng := rand.New(rand.NewSource(42))
	ek := api.GenerateEK(rng)
	root, _ := api.GenerateIdentityKey(rng)
	aik, attest, sig, aikPriv := api.GenerateDeviceKey(root, rng)
	enrollment := &api.Enrollment{
		NameHint:               "localhost",
		EndorsementKey:         api.PublicKey(ek),
		EndorsementCertificate: nil,
		Root:                   api.PublicKey(root),
		Keys: map[string]api.Key{
			"aik": {
				Public:                 api.PublicKey(aik),
				CreationProof:          api.Attest(attest),
				CreationProofSignature: api.Signature(sig),
			},
		},
		Cookie: "hello",
	}

	testCases := map[string]func(*testing.T, *pgxpool.Pool, *blob.Storage, *api.Enrollment, *ecdsa.PrivateKey){
		"Unseen":                  testOverrideUnseen,
		"Trusted":                 testOverrideTrusted,
		"Vulnerable":              testOverrideVuln,
		"VulnerableNoChange":      testOverrideVulnNoChange,
		"VulnerableAcceptNothing": testOverrideVulnAcceptNothing,
		"VulnerableClearedCache":  testOverrideVulnClearedCache,
		"NoSuchDevice":            testOverrideNoSuchDevice,
	}

	for name, fn := range testCases {
		pgsqlC.Reset(t, ctx, database.MigrationFiles)
		store, err := blob.NewStorage(ctx,
			blob.WithBucket{Bucket: minio.Bucket},
			blob.WithEndpoint{
				Endpoint: minio.Endpoint,
				Region:   "us-east-1",
			},
			blob.WithCredentials{
				Key:    minio.Key,
				Secret: minio.Secret,
			},
			blob.WithTestMode{Enable: true})
		assert.NoError(t, err)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn, store, enrollment, aikPriv)
		})
	}
}

func enroll(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	// enroll
	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)
}

func makeVulnerable(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	enroll(t, conn, store, en, aikPriv)
	testProps := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	props := api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		SMBIOS:        testProps.Firmware.SMBIOS,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, testProps.Firmware, time.Now())
	finishAttest(t, conn, store, en, aikPriv)
	testProps = parseEvidence(t, "../../test/DESKTOP-MIMU51J-no-secureboot.json")
	props = api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		SMBIOS:        testProps.Firmware.SMBIOS,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, props, time.Now())
	finishAttest(t, conn, store, en, aikPriv)
}

func startAttest(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		SMBIOS: testProps.Firmware.SMBIOS,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, props, time.Now())
}

func startAttestWithProps(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey, props api.FirmwareProperties, now time.Time) {
	ctx := context.Background()

	sha1bank := map[string]api.Buffer{}
	sha256bank := map[string]api.Buffer{}
	signingHasher := sha256.New()
	pcrsel := []tpm2.PCRSelection{}
	evlog, err := eventlog.ParseEventLog(props.TPM2EventLog.Data)
	if err == nil {
		var /*sha1sel,*/ sha256sel []int

		// sha1
		for pcr := 0; pcr < 24; pcr += 1 {
			pcrs := []eventlog.PCR{
				{Index: pcr, DigestAlg: crypto.SHA1, Digest: make([]byte, 20)},
			}
			digest, _ := evlog.Replay(pcrs)
			if len(digest) < 1 || len(digest[0]) == 0 {
				continue
			}
			sha1bank[strconv.Itoa(pcr)] = api.Buffer(digest[0])
			//signingHasher.Write(digest[0])
			//sha1sel = append(sha1sel, pcr)
		}
		//pcrsel = append(pcrsel, tpm2.PCRSelection{Hash: tpm2.AlgSHA1, PCRs: sha1sel})

		// sha256
		for pcr := 0; pcr < 24; pcr += 1 {
			pcrs := []eventlog.PCR{
				{Index: pcr, DigestAlg: crypto.SHA256, Digest: make([]byte, 32)},
			}
			digest, _ := evlog.Replay(pcrs)
			if len(digest) < 1 || len(digest[0]) == 0 {
				continue
			}
			sha256bank[strconv.Itoa(pcr)] = api.Buffer(digest[0])
			signingHasher.Write(digest[0])
			sha256sel = append(sha256sel, pcr)
		}
		pcrsel = append(pcrsel, tpm2.PCRSelection{Hash: tpm2.AlgSHA256, PCRs: sha256sel})
	}

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	attest := tpm2.AttestationData{
		Magic:           0xFF544347,
		Type:            tpm2.TagAttestQuote,
		ExtraData:       propsHash[0],
		QualifiedSigner: tpm2.Name(aikName),
		AttestedQuoteInfo: &tpm2.QuoteInfo{
			PCRSelection: pcrsel,
			PCRDigest:    signingHasher.Sum([]byte{}),
		},
	}

	attestBlob, err := attest.Encode()
	assert.NoError(t, err)
	attestHasher := sha256.New()
	attestHasher.Write(attestBlob)
	r, s, err := ecdsa.Sign(crand.Reader, aikPriv, attestHasher.Sum([]byte{}))
	assert.NoError(t, err)
	sig := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{
			HashAlg: tpm2.AlgSHA256,
			R:       r,
			S:       s,
		},
	}

	point := (tpm2.Public(en.Keys["aik"].Public)).ECCParameters.Point
	assert.Equal(t, 0, point.X().Cmp(aikPriv.X))
	assert.Equal(t, 0, point.Y().Cmp(aikPriv.Y))

	ev := &api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(attest),
		Signature: api.Signature(sig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      sha256bank,
		AllPCRs:   map[string]map[string]api.Buffer{"11": sha256bank}, //, "4": sha1bank},
		Firmware:  props,
		Cookie:    "hello",
	}
	_, err = Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
}

func finishAttest(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()

	// run binarly and tsc jobs
	proc1, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc1, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", time.Now(), 30*time.Second)
		assert.NoError(t, err)
	})
	proc2, err := inteltsc.NewProcessor(ctx, conn, []inteltsc.Site{inteltsc.KaisTestCredentials})
	assert.NoError(t, err)

	queue.RunProcessorWithObserver(t, conn, proc2, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", time.Now(), 30*time.Second)
		assert.NoError(t, err)
	})
}

func testOverrideUnseen(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()

	enroll(t, conn, store, en, aikPriv)
	startAttest(t, conn, store, en, aikPriv)

	// override unseen
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := Override(ctx, tx, store, 1000, 1000, []string{"alalal"}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateUnseen, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	// start attestation
	// override still unseen (evidence in progress)
	tx, err = conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err = Override(ctx, tx, store, 1000, 1000, []string{"alalal"}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateUnseen, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func testOverrideTrusted(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	orgId := int64(1000)

	enroll(t, conn, store, en, aikPriv)
	startAttest(t, conn, store, en, aikPriv)
	finishAttest(t, conn, store, en, aikPriv)

	now := time.Now()
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	_, err = device.Get(ctx, tx, 1000, orgId, now)
	assert.NoError(t, err)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	// override
	tx, err = conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := Override(ctx, tx, store, 1000, orgId, []string{"alalal"}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func testOverrideVuln(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	orgId := int64(1000)

	makeVulnerable(t, conn, store, en, aikPriv)

	now := time.Now()
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := device.Get(ctx, tx, 1000, orgId, now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateVuln, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	// override
	tx, err = conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err = Override(ctx, tx, store, 1000, orgId, []string{
		issuesv1.UefiSecureBootVariablesId,
	}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	appr, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, conn, dev.Id, false, 1)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Len(t, appr, 1)
	if assert.NoError(t, err) {
		assert.WithinDuration(t, time.Now(), appr[0].Received, time.Second)
	}
}

func testOverrideVulnNoChange(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	orgId := int64(1000)

	makeVulnerable(t, conn, store, en, aikPriv)

	// override
	now := time.Now()
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := Override(ctx, tx, store, 1000, orgId, []string{
		issuesv1.TpmNoEventlogId,
	}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateVuln, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func testOverrideVulnAcceptNothing(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	orgId := int64(1000)

	makeVulnerable(t, conn, store, en, aikPriv)

	// override
	now := time.Now()
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := Override(ctx, tx, store, 1000, orgId, []string{}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateVuln, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func testOverrideVulnClearedCache(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	orgId := int64(1000)
	binRef := ""

	enroll(t, conn, store, en, aikPriv)
	testProps := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	props := api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		Flash:         testProps.Firmware.Flash,
		SMBIOS:        testProps.Firmware.SMBIOS,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, testProps.Firmware, time.Now())
	proc1, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc1, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", time.Now(), 30*time.Second)
		assert.NoError(t, err)
	})
	finishAttest(t, conn, store, en, aikPriv)
	testProps = parseEvidence(t, "../../test/DESKTOP-MIMU51J-no-secureboot.json")
	props = api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		SMBIOS:        testProps.Firmware.SMBIOS,
		Flash:         testProps.Firmware.Flash,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, props, time.Now())
	queue.RunProcessorWithObserver(t, conn, proc1, func(ctx context.Context, ty string, ref string) {
		binRef = ref
		err := Appraise(context.Background(), conn, store, ref, "testsrv", time.Now(), 30*time.Second)
		assert.NoError(t, err)
	})
	finishAttest(t, conn, store, en, aikPriv)

	// remove cached results
	assert.True(t, strings.HasPrefix(binRef, "binarly"))
	conn.Exec(ctx, "delete from v2.jobs where reference = $1", binRef)

	// override
	now := time.Now()
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := Override(ctx, tx, store, 1000, orgId, []string{
		issuesv1.UefiSecureBootVariablesId,
	}, "testsrv", now, "test")
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func testOverrideNoSuchDevice(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	orgId := int64(1000)
	now := time.Now()
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := Override(ctx, tx, store, 1000, orgId, []string{}, "testsrv", now, "test")
	assert.Equal(t, err, database.ErrNotFound)
	assert.Nil(t, dev)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}
