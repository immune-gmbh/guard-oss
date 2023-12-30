package workflow

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cebind "github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/go-tpm/tpm2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/device"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/event"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

func TestEnroll(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	rng := rand.New(rand.NewSource(42))
	ek := api.GenerateEK(rng)
	root, _ := api.GenerateIdentityKey(rng)
	aik, attest, sig, _ := api.GenerateDeviceKey(root, rng)
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

	testCases := map[string]func(*testing.T, *pgxpool.Pool, *api.Enrollment){
		"Successful": testSuccessfulEnroll,
		"OverQuota":  testOverQuotaEnroll,
	}

	for name, fn := range testCases {
		pgsqlC.Reset(t, ctx, database.MigrationFiles)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn, enrollment)
		})
	}
}

func testSuccessfulEnroll(t *testing.T, conn *pgxpool.Pool, enrollment *api.Enrollment) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	_, creds, err := Enroll(ctx, conn, enrollment, caPriv, caKid, "testsrv", "ext-1", now)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	// billing update event
	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		ctx := req.Context()
		msg := cehttp.NewMessageFromHttpRequest(req)
		ev, err := cebind.ToEvent(ctx, msg)
		assert.NoError(t, err)
		assert.Equal(t, ev.Type(), api.BillingUpdateEventType)
		resp := httptest.NewRecorder()
		resp.WriteHeader(202)
		send = true
		return resp.Result()
	})
	ev, err := event.NewProcessor(ctx, "http://example.com", "testsrv",
		event.WithClient{
			Client: client,
		},
		event.WithCredentials{
			PrivateKey: caPriv,
			Kid:        caKid,
		})
	assert.NoError(t, err)
	queue.RunProcessor(t, conn, ev)
	assert.True(t, send)

	// db entry
	aikName, err := api.ComputeName(tpm2.HandleEndorsement, enrollment.Root, enrollment.Keys["aik"].Public)
	assert.NoError(t, err)
	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	defer tx.Rollback(ctx)
	dev, err := device.GetByFingerprint(ctx, tx, &aikName)
	assert.NoError(t, err)

	// GetByFingerprint is specialized and does not return device state, query state separately here
	dev2, err := device.Get(ctx, tx, dev.Id, dev.OrganizationId, now)
	assert.NoError(t, err)
	assert.Equal(t, dev2.State, api.StateUnseen)
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func testOverQuotaEnroll(t *testing.T, conn *pgxpool.Pool, enrollment *api.Enrollment) {
	ctx := context.Background()
	now := time.Now()

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	assert.NoError(t, err)
	caKid, err := key.ComputeKid(&caPriv.PublicKey)
	assert.NoError(t, err)

	err = organization.UpdateQuota(ctx, conn, "ext-1", 0, now)
	assert.NoError(t, err)

	_, _, err = Enroll(ctx, conn, enrollment, caPriv, caKid, "testsrv", "ext-1", now)
	assert.Equal(t, ErrQuotaExceeded, err)
}
