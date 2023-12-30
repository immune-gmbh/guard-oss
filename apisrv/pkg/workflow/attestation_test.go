package workflow

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/agent/v3/pkg/typevisit"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/appraisal"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/binarly"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/event"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/inteltsc"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

func parseEvidence(t *testing.T, f string) *api.Evidence {
	buf, err := os.ReadFile(f)
	assert.NoError(t, err)

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	assert.NoError(t, err)

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]api.Buffer{
			"11": ev.PCRs,
		}
	}

	return &ev
}

func TestTypeVisitor(t *testing.T) {
	firmware := api.FirmwareProperties{EPPInfo: &api.EPPInfo{EarlyLaunchDrivers: map[string]api.HashBlob{"brab": {ZData: api.Buffer{23}}}, AntimalwareProcesses: map[string]api.HashBlob{"brab": {ZData: api.Buffer{23}}}}}
	var up blob.UploadJob
	BlobStoreVisitor.Visit(&firmware, func(v reflect.Value, opts typevisit.FieldOpts) {
		if opts != blob.WindowsExecutable {
			return
		}

		assert.Equal(t, reflect.Map, v.Kind())
		keys := v.MapKeys()
		if len(keys) == 1 {
			assert.Equal(t, "brab", keys[0].Interface().(string))
		} else {
			return
		}

		hbuf := v.MapIndex(keys[0]).Interface().(api.HashBlob)
		if len(hbuf.ZData) > 0 {
			up = blob.UploadJob{
				CompressedContents: hbuf.ZData,
				Namespace:          string(opts),
			}
		}
	})
	assert.Equal(t, byte(23), up.CompressedContents[0])
}

func TestAttest(t *testing.T) {
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
		"NoJobs":           testAttestNoJob,
		"Concurrent":       testAttestConcurrent,
		"Binarly":          testAttestBinarly,
		"TSC":              testAttestTSC,
		"OverQuota":        testOverQuotaAttest,
		"FailedPrevAttest": testFailedPrevAttest,
		"Timeout":          testAttestTimeout,
		"NoSupplyChain":    testAttestNoSupplyChain,
		"Disabled":         testAttestAPIDisabled,
		"PolicyIntelTSC":   testAttestPolicyIntelTSC,
		"PolicyEPP":        testAttestPolicyEPP,
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

func testAttestNoJob(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	props := api.FirmwareProperties{}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := &api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, appr)

	dev, err := device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)
}

func testAttestBinarly(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		Flash: testProps.Firmware.Flash,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	// copy for l8r b/c evidence is changed due to hashblob processing
	ev2 := ev

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, &ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	dev, err := device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.WithinDuration(t, now, *dev.AttestationInProgress, time.Second)

	// run binarly jobs
	count := 0
	proc, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", now, 30*time.Second)
		assert.NoError(t, err)
		count += 1
	})
	// submit & poll
	assert.Equal(t, 2, count)

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	dev, err = device.Get(ctx, tx, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	assert.Nil(t, dev.AttestationInProgress)

	tx.Rollback(ctx)

	// attest again, expecting cached results
	now = time.Now()
	appr, err = Attest(ctx, conn, store, &ev2, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, appr)

	// send events
	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(202)
		send = true
		return resp.Result()
	})
	proc2, err := event.NewProcessor(ctx, "http://example.com", "testsrv",
		event.WithClient{Client: client})
	assert.NoError(t, err)
	queue.RunProcessor(t, conn, proc2)
	assert.True(t, send)

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	assert.Nil(t, dev.AttestationInProgress)
}

func testAttestTSC(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		SMBIOS: testProps.Firmware.SMBIOS,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	// copy for l8r b/c evidence is changed due to hashblob processing
	ev2 := ev

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, &ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	// run tsc jobs
	count := 0
	proc, err := inteltsc.NewProcessor(ctx, conn, []inteltsc.Site{inteltsc.KaisTestCredentials})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", now, 30*time.Second)
		assert.NoError(t, err)
		count += 1
	})
	assert.Less(t, 0, count)

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	dev, err := device.Get(ctx, tx, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	tx.Rollback(ctx)

	// attest again, expecting cached results
	now = time.Now()
	appr, err = Attest(ctx, conn, store, &ev2, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, appr)

	// send events
	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(202)
		send = true
		return resp.Result()
	})
	proc2, err := event.NewProcessor(ctx, "http://example.com", "testsrv",
		event.WithClient{Client: client})
	assert.NoError(t, err)
	queue.RunProcessor(t, conn, proc2)
	assert.True(t, send)
}

func testAttestConcurrent(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		Flash:  testProps.Firmware.Flash,
		SMBIOS: testProps.Firmware.SMBIOS,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}
	// copy for l8r b/c evidence is changed due to hashblob processing
	ev2 := ev

	dev, err := device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, &ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.WithinDuration(t, now, *dev.AttestationInProgress, time.Second)

	now = now.Add(time.Minute)

	// attest again while other attest is stil WIP
	appr, err = Attest(ctx, conn, store, &ev2, nil, &aikName, "testsrv", now, 90*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.WithinDuration(t, now, *dev.AttestationInProgress, time.Second)

	// run binarly and tsc jobs
	proc1, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc1, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", now, 90*time.Second)
		assert.NoError(t, err)
	})
	proc2, err := inteltsc.NewProcessor(ctx, conn, []inteltsc.Site{inteltsc.KaisTestCredentials})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc2, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", now, 90*time.Second)
		assert.NoError(t, err)
	})

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	assert.Nil(t, dev.AttestationInProgress)
}

func testAttestJobTimeout(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		Flash:  testProps.Firmware.Flash,
		SMBIOS: testProps.Firmware.SMBIOS,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := &api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	dev, err := device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.WithinDuration(t, now, *dev.AttestationInProgress, time.Second)

	now = now.Add(time.Minute)

	// attest again while other attest is stil WIP
	appr, err = Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 90*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.WithinDuration(t, now, *dev.AttestationInProgress, time.Second)

	// run binarly and tsc jobs
	proc1, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc1, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", now, 90*time.Second)
		assert.NoError(t, err)
	})
	proc2, err := inteltsc.NewProcessor(ctx, conn, []inteltsc.Site{inteltsc.KaisTestCredentials})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc2, func(ctx context.Context, ty string, ref string) {
		err := Appraise(context.Background(), conn, store, ref, "testsrv", now, 90*time.Second)
		assert.NoError(t, err)
	})

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	assert.Nil(t, dev.AttestationInProgress)
}
func testOverQuotaAttest(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		Flash:  testProps.Firmware.Flash,
		SMBIOS: testProps.Firmware.SMBIOS,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := &api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	// set quota to zero
	err = organization.UpdateQuota(ctx, conn, "ext-1", 0, now)
	assert.NoError(t, err)

	// attest
	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	_, err = Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.Equal(t, ErrQuotaExceeded, err)

	dev, err := device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)
}

func testFailedPrevAttest(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		Flash: testProps.Firmware.Flash,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}
	// copy for l8r b/c evidence is changed by attest due to hashblob processing
	ev2 := ev

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, &ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	// run binarly jobs
	count := 0
	proc, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc, func(ctx context.Context, ty string, ref string) {
		// swallow attest
		count += 1
	})
	assert.Equal(t, 2, count)

	// attest again
	appr, err = Attest(ctx, conn, store, &ev2, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, appr)

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	dev, err := device.Get(ctx, tx, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	assert.Nil(t, dev.AttestationInProgress)

	tx.Rollback(ctx)
}

func testAttestTimeout(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		Flash: testProps.Firmware.Flash,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := &api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)
	appr, err := Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 30*time.Second)
	assert.NoError(t, err)
	assert.Nil(t, appr)

	dev, err := device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.NotNil(t, dev.AttestationInProgress)

	now = now.Add(20 * time.Minute)

	dev, err = device.Get(ctx, conn, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)
}

func testAttestNoSupplyChain(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	devid, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, props, now)
	assert.NoError(t, err)

	dev, err := device.Get(ctx, conn, devid, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)

	appr, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, conn, dev.Id, false, 1)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Len(t, appr, 1)
	assert.Equal(t, api.Unsupported, appr[0].Verdict.SupplyChain)
}

func testAttestAPIDisabled(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		SMBIOS: testProps.Firmware.SMBIOS,
	}
	propsHash, err := evidence.EvidenceHashes(ctx, &api.Evidence{Firmware: props})
	assert.NoError(t, err)
	quote, quoteSig := api.GenerateQuote(tpm2.Public(en.Root), tpm2.Public(en.Keys["aik"].Public), aikPriv, propsHash[0], rng)
	ev := &api.Evidence{
		Type:      api.EvidenceType,
		Quote:     api.Attest(quote),
		Signature: api.Signature(quoteSig),
		Algorithm: fmt.Sprint(int(tpm2.AlgSHA256)),
		PCRs:      map[string]api.Buffer{"1": {}, "2": {}, "3": {}},
		AllPCRs:   map[string]map[string]api.Buffer{"11": {"1": {}, "2": {}, "3": {}}},
		Firmware:  props,
		Cookie:    "hello",
	}

	aikName, err := api.ComputeName(tpm2.HandleEndorsement, en.Root, en.Keys["aik"].Public)
	assert.NoError(t, err)

	DisableIntelTSC = true
	DisableBinarly = true

	appr, err := Attest(ctx, conn, store, ev, nil, &aikName, "testsrv", now, 10*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, appr)

	// run tsc jobs
	proc1, err := inteltsc.NewProcessor(ctx, conn, []inteltsc.Site{inteltsc.KaisTestCredentials})
	assert.NoError(t, err)
	proc2, err := binarly.NewProcessor(ctx, conn, store,
		binarly.WithTestmode{
			Enabled: true,
		})
	assert.NoError(t, err)
	assert.NoError(t, err)
	queue.RunProcessorWithObserver(t, conn, proc1, func(ctx context.Context, ty string, ref string) {
		assert.Fail(t, "tsc job")
	})
	queue.RunProcessorWithObserver(t, conn, proc2, func(ctx context.Context, ty string, ref string) {
		assert.Fail(t, "binarly job")
	})

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	dev, err := device.Get(ctx, tx, 1000, int64(1000), now)
	assert.NoError(t, err)
	assert.Equal(t, api.StateTrusted, dev.State)
	tx.Rollback(ctx)

	DisableIntelTSC = false
	DisableBinarly = false
}

func testAttestPolicyIntelTSC(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	devid, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	// must have tsc
	pol := policy.New()
	pol.IntelTSC = policy.True
	err = device.Patch(ctx, conn, devid, int64(1000), nil, nil, pol, now, "test")
	assert.NoError(t, err)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, props, now)
	assert.NoError(t, err)

	dev, err := device.Get(ctx, conn, devid, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)

	appr, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, conn, dev.Id, false, 1)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Len(t, appr, 1)
	assert.Equal(t, api.Vulnerable, appr[0].Verdict.SupplyChain)
}

func testAttestPolicyEPP(t *testing.T, conn *pgxpool.Pool, store *blob.Storage, en *api.Enrollment, aikPriv *ecdsa.PrivateKey) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	devid, creds, err := Enroll(ctx, conn, en, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	// must have epp
	pol := policy.New()
	pol.EndpointProtection = policy.True
	err = device.Patch(ctx, conn, devid, int64(1000), nil, nil, pol, now, "test")
	assert.NoError(t, err)

	testProps := parseEvidence(t, "../../test/issue926-x1carbon-after.evidence.json")
	props := api.FirmwareProperties{
		UEFIVariables: testProps.Firmware.UEFIVariables,
		TPM2EventLog:  testProps.Firmware.TPM2EventLog,
	}
	startAttestWithProps(t, conn, store, en, aikPriv, props, now)
	assert.NoError(t, err)

	dev, err := device.Get(ctx, conn, devid, int64(1000), now)
	assert.NoError(t, err)
	assert.Nil(t, dev.AttestationInProgress)

	appr, err := appraisal.GetLatestAppraisalsByDeviceId(ctx, conn, dev.Id, false, 1)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Len(t, appr, 1)
	assert.Equal(t, api.Vulnerable, appr[0].Verdict.EndpointProtection)
}
