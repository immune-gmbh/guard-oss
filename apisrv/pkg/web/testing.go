package web

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/jsonapi"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/organization"
)

var testOrg = "ext-id-1"

func setupOrg(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	err := organization.UpdateQuota(ctx, pool, testOrg, 1000, time.Now())
	assert.NoError(t, err)
}

type testServer struct {
	Auth     *key.Key
	AuthPriv *ecdsa.PrivateKey
	Handler  http.Handler
}

func (srv *testServer) Get(t *testing.T, r string, org string) ([]byte, int) {
	req, err := http.NewRequest("GET", r, bytes.NewBufferString(""))
	assert.NoError(t, err)

	token, err := authentication.IssueUserCredential("testsrv", org, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	rr := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	assert.NoError(t, err)

	return body, rr.Code
}

func (srv *testServer) Patch(t *testing.T, r string, payload interface{}, org string) ([]byte, int) {
	buf := new(bytes.Buffer)
	err := jsonapi.MarshalPayload(buf, payload)
	//err := json.NewEncoder(buf).Encode(payload)
	assert.NoError(t, err)

	req, err := http.NewRequest("PATCH", r, buf)
	assert.NoError(t, err)

	token, err := authentication.IssueUserCredential("testsrv", org, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	rr := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	assert.NoError(t, err)

	return body, rr.Code
}

func (srv *testServer) Post(t *testing.T, r string, payload interface{}, auth string) ([]byte, int) {
	buf := new(bytes.Buffer)
	//err := jsonapi.MarshalPayload(buf, payload)
	err := json.NewEncoder(buf).Encode(payload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", r, buf)
	assert.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	rr := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	assert.NoError(t, err)

	return body, rr.Code
}
