package event

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

func TestEventProcessor(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"SendAck":        testSendAck,
		"SendNack":       testSendNack,
		"EventTypes":     testEventTypes,
		"Config":         testConfig,
		"SuccessfulPing": testSuccessfulPing,
		"FailedPing":     testFailedPing,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.Reset(t, ctx, database.MigrationFiles)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testSendAck(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	devs := map[string]int{
		"d5d70fad-c11f-4609-8cab-17438305f0b0": 1337,
		"39f5930f-d9cf-49d4-9ad4-2054ae4dfea5": 1,
		"3fab3a3e-c286-47cf-a096-dfc899441c8e": 0,
		"f2561288-76e5-42cb-bea8-90c8d6265a63": 42,
		"ad89ec0a-d2a7-40a9-8297-d87fae349311": 23,
		"f8e47633-a3b5-4f7e-bcaa-3272fe51f7fa": 420,
		"1222fe6a-e762-41b9-8252-5af7a6547342": 999999,
	}

	row, err := BillingUpdate(ctx, db, "test", devs, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)

	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(202)
		send = true
		return resp.Result()
	})
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	proc, err := NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: priv,
			Kid:        "abc",
		},
		WithClient{Client: client})
	assert.NoError(t, err)

	job := queue.Job{Row: row}
	proc.Run(ctx, &job)
	assert.True(t, *job.Row.Successful)
	assert.True(t, send)
}

func testSendNack(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	devs := map[string]int{
		"d5d70fad-c11f-4609-8cab-17438305f0b0": 1337,
		"39f5930f-d9cf-49d4-9ad4-2054ae4dfea5": 1,
		"3fab3a3e-c286-47cf-a096-dfc899441c8e": 0,
		"f2561288-76e5-42cb-bea8-90c8d6265a63": 42,
		"ad89ec0a-d2a7-40a9-8297-d87fae349311": 23,
		"f8e47633-a3b5-4f7e-bcaa-3272fe51f7fa": 420,
		"1222fe6a-e762-41b9-8252-5af7a6547342": 999999,
	}

	row, err := BillingUpdate(ctx, db, "test", devs, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)

	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(400)
		send = true
		return resp.Result()
	})
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	proc, err := NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: priv,
			Kid:        "abc",
		},
		WithClient{Client: client})
	assert.NoError(t, err)

	job := queue.Job{Row: row}
	proc.Run(ctx, &job)
	assert.Nil(t, job.Row.Successful)
	assert.True(t, send)
}

func testEventTypes(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now()
	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(202)
		send = true
		return resp.Result()
	})
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	proc, err := NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: priv,
			Kid:        "abc",
		},
		WithClient{Client: client})
	assert.NoError(t, err)

	// billing
	devs := map[string]int{
		"d5d70fad-c11f-4609-8cab-17438305f0b0": 1337,
		"39f5930f-d9cf-49d4-9ad4-2054ae4dfea5": 1,
		"3fab3a3e-c286-47cf-a096-dfc899441c8e": 0,
		"f2561288-76e5-42cb-bea8-90c8d6265a63": 42,
		"ad89ec0a-d2a7-40a9-8297-d87fae349311": 23,
		"f8e47633-a3b5-4f7e-bcaa-3272fe51f7fa": 420,
		"1222fe6a-e762-41b9-8252-5af7a6547342": 999999,
	}
	row, err := BillingUpdate(ctx, db, "test", devs, now, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)
	job := queue.Job{Row: row}
	proc.Run(ctx, &job)
	assert.True(t, *job.Row.Successful)
	assert.True(t, send)
	send = false

	// expired appraisal
	dev := api.Device{
		Id:     "112233",
		Cookie: "asasasas",
		Name:   "Device #1",
		State:  "outdated",
	}
	appr1 := api.Appraisal{
		Id:       "testlalala",
		Received: time.Now().UTC().Add(-10 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(-1 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Trusted,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Trusted,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	err = json.Unmarshal([]byte(reportStr), &appr1.Report)
	assert.NoError(t, err)

	row, err = AppraisalExpired(ctx, db, "test", "org", &dev, &appr1, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)
	job = queue.Job{Row: row}
	proc.Run(ctx, &job)
	assert.True(t, *job.Row.Successful)
	assert.True(t, send)
	send = false

	// failed appraisal
	dev = api.Device{
		Id:     "112233",
		Cookie: "asasasas",
		Name:   "Device #1",
		State:  "vulnerable",
	}
	appr1 = api.Appraisal{
		Id:       "testlalala",
		Received: time.Now().UTC().Add(-10 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(10 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	err = json.Unmarshal([]byte(reportStr), &appr1.Report)
	assert.NoError(t, err)

	row, err = NewAppraisal(ctx, db, "test", "org", &dev, nil, &appr1, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)
	job = queue.Job{Row: row}
	proc.Run(ctx, &job)
	assert.True(t, *job.Row.Successful)
	assert.True(t, send)
	send = false

	// continued failing appraisal
	appr2 := api.Appraisal{
		Id:       "testlololl",
		Received: time.Now().UTC().Add(-4 * time.Hour * 24),
		Expires:  time.Now().UTC().Add(10 * time.Hour * 24),
		Verdict: api.Verdict{
			Type:               api.VerdictType,
			Result:             api.Vulnerable,
			SupplyChain:        api.Trusted,
			Configuration:      api.Trusted,
			Firmware:           api.Vulnerable,
			Bootloader:         api.Trusted,
			OperatingSystem:    api.Trusted,
			EndpointProtection: api.Trusted,
		},
		Report: api.Report{},
	}
	err = json.Unmarshal([]byte(reportStr), &appr1.Report)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(reportStr), &appr2.Report)
	assert.NoError(t, err)

	row, err = NewAppraisal(ctx, db, "test", "org", &dev, &appr1, &appr2, now)
	assert.NoError(t, err)
	assert.NotNil(t, row)
	job = queue.Job{Row: row}
	proc.Run(ctx, &job)
	assert.True(t, *job.Row.Successful)
	assert.True(t, send)
}

func testConfig(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	proc, err := NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: nil,
			Kid:        "abc",
		})
	assert.Error(t, err)
	assert.Nil(t, proc)

	proc, err = NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: priv,
			Kid:        "",
		})
	assert.Error(t, err)
	assert.Nil(t, proc)

	proc, err = NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithClient{Client: nil})
	assert.Error(t, err)
	assert.Nil(t, proc)

	proc, err = NewProcessor(ctx, "http://example.com/events", "testsrv", 42)
	assert.Error(t, err)
	assert.Nil(t, proc)
}

func testSuccessfulPing(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(202)
		send = true
		return resp.Result()
	})
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	proc, err := NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: priv,
			Kid:        "abc",
		},
		WithClient{Client: client})
	assert.NoError(t, err)

	err = proc.PingReceiver()
	assert.NoError(t, err)
	assert.True(t, send)
}

func testFailedPing(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	send := false
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(400)
		send = true
		return resp.Result()
	})
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	proc, err := NewProcessor(ctx, "http://example.com/events", "testsrv",
		WithCredentials{
			PrivateKey: priv,
			Kid:        "abc",
		},
		WithClient{Client: client})
	assert.NoError(t, err)

	err = proc.PingReceiver()
	assert.Error(t, err)
	assert.True(t, send)
}
