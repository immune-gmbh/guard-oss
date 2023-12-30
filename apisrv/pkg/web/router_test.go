package web

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/jsonapi"
	"github.com/gowebpki/jcs"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/appraisal"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/authentication"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/blob"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/change"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/configuration"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"

	agapi "github.com/immune-gmbh/agent/v3/pkg/api"
	agattest "github.com/immune-gmbh/agent/v3/pkg/core"
	agstate "github.com/immune-gmbh/agent/v3/pkg/state"
)

//go:embed seed.sql
var seedSql string

//go:embed seed-tagged.sql
var seedTaggedSql string

//go:embed seed-issue.sql
var seedIssueSql string

func TestRouter(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool, *testServer){
		"GetSingleDevice":         testGetDevice,
		"PatchSingleDevice":       testPatchPolicyDevice,
		"GetSingleDeviceWithTags": testGetDeviceWithTags,
		"Configuration":           testConfiguration,
		"ListChanges":             testListChanges,
		"Override":                testAttestAndOverride,
		"Attest":                  testAttest,
		"Tags":                    testTags,
		"Issue1176":               testIssue1176,
		"Issue1576":               testIssue1576,
		"Issues":                  testIssues,
		"Dashboard":               testDashboard,
		"FilterTags":              testFilterTags,
		"FilterState":             testFilterState,
		"FilterIssue":             testFilterIssue,
	}
	pgsql := mock.PostgresContainer(t, ctx)
	defer pgsql.Terminate(ctx)
	minio := mock.MinioContainer(t, ctx)
	defer minio.Terminate(ctx)

	// please only increase level temporarily for specific debug session; it increases clutter significantly
	log.Logger = log.Logger.Level(zerolog.InfoLevel)

	for name, fn := range tests {
		if name == "Tags" || name == "FilterTags" {
			pgsql.ResetAndSeed(t, ctx, database.MigrationFiles, seedTaggedSql, 35)
		} else if name == "FilterIssue" {
			pgsql.ResetAndSeed(t, ctx, database.MigrationFiles, seedIssueSql, 50)
		} else {
			pgsql.ResetAndSeed(t, ctx, database.MigrationFiles, seedSql, 16)
		}

		t.Run(name, func(t *testing.T) {
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

			conn := pgsql.Connect(t, ctx)
			auth, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			assert.NoError(t, err)
			kid, err := key.ComputeKid(&auth.PublicKey)
			assert.NoError(t, err)
			authKey := key.Key{
				Kid:    kid,
				Issuer: "testsrv",
				Key:    auth.PublicKey,
			}
			pingCh := make(chan bool)
			ks := key.NewSet()
			ks.Replace(&[]key.Key{authKey})
			router, err := NewRouter(ctx, conn, store, &auth.PublicKey, auth, ks, pingCh, "testsrv", "http://api.example.com/", "http://app.example.com/", nil, time.Hour)
			assert.NoError(t, err)
			srv := testServer{
				Auth:     &authKey,
				AuthPriv: auth,
				Handler:  router,
			}

			defer conn.Close()
			fn(t, conn, &srv)
		})
	}
}

func testConfiguration(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	server := httptest.NewServer(srv.Handler)
	client := server.Client()

	// no header
	req, err := http.NewRequest("GET", server.URL+"/v2/configuration", nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	cfg := new(jsonapi.OnePayload)
	err = json.NewDecoder(resp.Body).Decode(cfg)
	assert.NoError(t, err)

	// modifed
	req, err = http.NewRequest("GET", server.URL+"/v2/configuration", nil)
	assert.NoError(t, err)
	req.Header.Add("If-Modified-Since", configuration.DefaultConfigurationModTime.UTC().Add(-1*time.Hour).Format(http.TimeFormat))
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(cfg)
	assert.NoError(t, err)

	// not modified
	req, err = http.NewRequest("GET", server.URL+"/v2/configuration", nil)
	assert.NoError(t, err)
	req.Header.Add("If-Modified-Since", time.Now().UTC().Format(http.TimeFormat))
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotModified, resp.StatusCode)
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Empty(t, buf)
}

func testGetDevice(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	body, status := srv.Get(t, "/v2/devices/100", testOrg)
	assert.Equal(t, http.StatusOK, status)

	dev2 := new(api.Device)
	err := jsonapi.UnmarshalPayload(bytes.NewReader(body), dev2)
	assert.NoError(t, err)
	assert.Equal(t, dev2.Name, "Test Device #1")
}

func testRenameDevice(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	// rename
	name := "Lalalallaa"
	patch := []*api.DevicePatch{
		{
			Id:   "100",
			Name: &name,
		},
	}
	body, status := srv.Patch(t, "/v2/devices/100", patch, testOrg)
	fmt.Println("1s PATCH", string(body))
	assert.Equal(t, http.StatusOK, status)

	// check name
	body, status = srv.Get(t, "/v2/devices/100", testOrg)
	assert.Equal(t, http.StatusOK, status)
	dev := new(api.Device)
	err := jsonapi.UnmarshalPayload(bytes.NewReader(body), dev)
	assert.NoError(t, err)
	assert.Equal(t, dev.Name, name)
}

func testPatchPolicyDevice(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	// change policy
	policypatch := api.DevicePolicy{
		EndpointProtection: "off",
		IntelTSC:           "off",
	}
	patch := []*api.DevicePatch{
		{
			Id:     "100",
			Policy: &policypatch,
		},
	}
	body, status := srv.Patch(t, "/v2/devices/100", patch, testOrg)
	fmt.Println("1s PATCH", string(body))
	assert.Equal(t, http.StatusOK, status)

	// check policy
	body, status = srv.Get(t, "/v2/devices/100", testOrg)
	assert.Equal(t, http.StatusOK, status)
	dev := new(api.Device)
	err := jsonapi.UnmarshalPayload(bytes.NewReader(body), dev)
	assert.NoError(t, err)
	assert.Equal(t, dev.Policy["endpoint_protection"], "off")
	assert.Equal(t, dev.Policy["intel_tsc"], "off")
}

func testGetDeviceWithTags(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	// fetch device 100, expect no tags
	body, status := srv.Get(t, "/v2/devices/100?fields[devices]=tags,name", testOrg)
	fmt.Println("1st GET", string(body))
	assert.Equal(t, http.StatusOK, status)

	dev2 := new(api.Device)
	err := jsonapi.UnmarshalPayload(bytes.NewReader(body), dev2)
	assert.NoError(t, err)
	assert.Equal(t, dev2.Name, "Test Device #1")
	assert.Empty(t, dev2.Tags)

	// add tags a,b,c to device 100
	patch := []*api.DevicePatch{
		{
			Id:   "100",
			Tags: []api.Tag{{Key: "a"}, {Key: "b"}, {Key: "c"}},
		},
	}
	body, status = srv.Patch(t, "/v2/devices/100", patch, testOrg)
	fmt.Println("1s PATCH", string(body))
	assert.Equal(t, http.StatusOK, status)

	// fetch device 100, expect tags a,b,c
	body, status = srv.Get(t, "/v2/devices/100?fields[devices]=tags,name", testOrg)
	fmt.Println("2nd GET", string(body))
	assert.Equal(t, http.StatusOK, status)

	err = jsonapi.UnmarshalPayload(bytes.NewReader(body), dev2)
	assert.NoError(t, err)
	assert.Equal(t, dev2.Name, "Test Device #1")
	var tmp []string
	for _, t := range dev2.Tags {
		tmp = append(tmp, t.Key)
	}
	assert.ElementsMatch(t, []string{"a", "b", "c"}, tmp)

	// delete tags a,b from device 100
	patch = []*api.DevicePatch{
		{
			Id:   "100",
			Tags: []api.Tag{{Key: "c"}},
		},
	}
	body, status = srv.Patch(t, "/v2/devices/100", patch, testOrg)
	fmt.Println("2nd PATCH", string(body))
	assert.Equal(t, http.StatusOK, status)

	// fetch device 100, expect tag c
	body, status = srv.Get(t, "/v2/devices/100?fields[devices]=tags,name", testOrg)
	fmt.Println("3rd GET", string(body))
	assert.Equal(t, http.StatusOK, status)

	err = jsonapi.UnmarshalPayload(bytes.NewReader(body), dev2)
	assert.NoError(t, err)
	assert.Equal(t, dev2.Name, "Test Device #1")
	tmp = nil
	for _, t := range dev2.Tags {
		tmp = append(tmp, t.Key)
	}
	assert.Equal(t, []string{"c"}, tmp)

	body, status = srv.Get(t, "/v2/tags", testOrg)
	fmt.Println("4th GET", string(body))
	assert.Equal(t, http.StatusOK, status)
	tags, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Tag)))
	assert.NoError(t, err)
	assert.Len(t, tags, 3)

	body, status = srv.Get(t, "/v2/tags?include=devices", testOrg)
	fmt.Println("5th GET", string(body))
	assert.Equal(t, http.StatusOK, status)

	body, status = srv.Get(t, "/v2/tags?search=a", testOrg)
	fmt.Println("6th GET", string(body))
	assert.Equal(t, http.StatusOK, status)
	tags, err = jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Tag)))
	assert.NoError(t, err)
	assert.Len(t, tags, 1)

	body, status = srv.Get(t, fmt.Sprint("/v2/tags/", tags[0].(*api.Tag).Id), testOrg)
	fmt.Println("7th GET", string(body))
	assert.Equal(t, http.StatusOK, status)
	tag := new(api.Tag)
	err = jsonapi.UnmarshalPayload(bytes.NewReader(body), tag)
	assert.NoError(t, err)

	body, status = srv.Get(t, "/v2/tags?include=devices", testOrg)
	fmt.Println("8th GET", string(body))
	assert.Equal(t, http.StatusOK, status)
}

func testListChanges(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	ctx := context.Background()

	tx, err := conn.Begin(ctx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer tx.Rollback(ctx)

	act := "Test Actor"
	devId := "120"
	_, err = change.New(ctx, tx, "rename", nil, testOrg, &devId, &act, time.Now())
	assert.NoError(t, err)
	tx.Commit(ctx)

	body, status := srv.Get(t, "/v2/changes", testOrg)
	assert.Equal(t, http.StatusOK, status)

	chs, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Change)))
	assert.NoError(t, err)
	assert.NotEmpty(t, chs)
}

func testAttestAsync(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	// enroll
	state := agstate.NewState()
	ctx := context.Background()
	server := httptest.NewServer(srv.Handler)
	base, err := url.Parse(server.URL + "/v2/")
	assert.NoError(t, err)
	apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	client := agapi.Client{
		HTTP:               server.Client(),
		Base:               base,
		Auth:               apiToken,
		HTTPRequestTimeout: time.Second * 300,
		PostRequestTimeout: time.Second * 300,
		AgentVersion:       "dummy",
	}
	_, err = state.EnsureFresh(&client)
	assert.NoError(t, err)
	enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)

	attestationClient := agattest.NewCore()
	attestationClient.State = state
	attestationClient.Client = client
	attestationClient.Log = &log.Logger
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	attestationClient.StatePath = f.Name()
	err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
	assert.NoError(t, err)

	// attest
	evidence := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	appr, _, err := sendEvidence(t, &client, state, evidence)
	assert.NoError(t, err)
	assert.Nil(t, appr)
}

var attestationFixtures = map[string]string{
	"latitude-7520.evidence.json":                 api.Trusted,
	"IMN-SUPERMICRO-EPP-MSDefender.evidence.json": api.Trusted,
	"sr630.evidence.json":                         api.Trusted,
	"r340.evidence.json":                          api.Trusted,
	"DESKTOP-MIMU51J-EPP-ESET.evidence.json":      api.Trusted,
	"h12ssl-2.evidence.json":                      api.Unsupported,
	"good-1.evidence.json":                        api.Unsupported,
	"sr630-no-ima-log.evidence.json":              api.Unsupported,
	"IMN-SUPERMICRO-no-epp.evidence.json":         api.Unsupported,
}

func testAttest(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	for fix, epp := range attestationFixtures {
		t.Run(fix, func(t *testing.T) {
			// tpm setup
			ctx := context.Background()
			state := agstate.NewState()

			// client setup
			server := httptest.NewServer(srv.Handler)
			base, err := url.Parse(server.URL + "/v2/")
			assert.NoError(t, err)
			apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
			assert.NoError(t, err)
			client := agapi.Client{
				HTTP:               server.Client(),
				Base:               base,
				Auth:               apiToken,
				HTTPRequestTimeout: time.Second * 300,
				PostRequestTimeout: time.Second * 300,
				AgentVersion:       "dummy",
			}

			// enroll
			_, err = state.EnsureFresh(&client)
			assert.NoError(t, err)
			enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
			assert.NoError(t, err)
			attestationClient := agattest.NewCore()
			attestationClient.State = state
			attestationClient.Client = client
			attestationClient.Log = &log.Logger
			f, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			attestationClient.StatePath = f.Name()
			err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
			assert.NoError(t, err)

			// attest
			evidence := parseEvidence(t, fmt.Sprint("../../test/", fix))
			// hack to prevent Binarly & TSC jobs
			evidence.Firmware.Flash.Data = nil
			evidence.Firmware.Flash.ZData = nil
			evidence.Firmware.Flash.Sha256 = nil
			evidence.Firmware.SMBIOS.Data = nil
			evidence.Firmware.SMBIOS.ZData = nil
			evidence.Firmware.SMBIOS.Sha256 = nil
			appr, link, err := sendEvidence(t, &client, state, evidence)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// check verdict
			assert.Equal(t, epp, appr.Verdict.EndpointProtection)

			re := regexp.MustCompile(`^http:/app.example.com/devices/(\d+)\?.+$`)
			devid := re.FindStringSubmatch(link)[1]

			_, status := srv.Post(t, fmt.Sprintf("/v2/devices/%s/override", devid), []string{
				issuesv1.TpmDummyId,
				issuesv1.TpmEndorsementCertUnverifiedId,
			}, apiToken)
			assert.Equal(t, http.StatusOK, status)

			evidence = parseEvidence(t, fmt.Sprint("../../test/", fix))
			evidence.Firmware.Flash.Data = nil
			evidence.Firmware.Flash.ZData = nil
			evidence.Firmware.Flash.Sha256 = nil
			evidence.Firmware.SMBIOS.Data = nil
			evidence.Firmware.SMBIOS.ZData = nil
			evidence.Firmware.SMBIOS.Sha256 = nil
			appr, _, err = sendEvidence(t, &client, state, evidence)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// check verdict
			assert.Equal(t, epp, appr.Verdict.EndpointProtection)

			// get device
			body, status := srv.Get(t, fmt.Sprint("/v2/devices/", devid), testOrg)
			assert.Equal(t, http.StatusOK, status)
			dev := new(api.Device)
			err = jsonapi.UnmarshalPayload(bytes.NewBuffer(body), dev)
			if assert.NoError(t, err) {
				assert.Equal(t, api.StateTrusted, dev.State)
			}
		})
	}
}

func testAttestAndOverride(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	ctx := context.Background()
	// enroll
	state := agstate.NewState()
	server := httptest.NewServer(srv.Handler)
	base, err := url.Parse(server.URL + "/v2/")
	assert.NoError(t, err)
	apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	client := agapi.Client{
		HTTP:               server.Client(),
		Base:               base,
		Auth:               apiToken,
		HTTPRequestTimeout: time.Second * 300,
		PostRequestTimeout: time.Second * 300,
		AgentVersion:       "dummy",
	}
	_, err = state.EnsureFresh(&client)
	assert.NoError(t, err)
	enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	attestationClient := agattest.NewCore()
	attestationClient.State = state
	attestationClient.Client = client
	attestationClient.Log = &log.Logger
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	attestationClient.StatePath = f.Name()
	err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
	assert.NoError(t, err)

	// attest
	evidence := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	_, _, err = sendEvidence(t, &client, state, evidence)
	assert.NoError(t, err)

	_, status := srv.Post(t, "/v2/devices/1000/override", []string{
		issuesv1.TpmDummyId,
		issuesv1.TpmEndorsementCertUnverifiedId,
	}, apiToken)
	assert.Equal(t, http.StatusOK, status)

	evidence = parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	appr, _, err := sendEvidence(t, &client, state, evidence)
	if assert.NoError(t, err) {
		assert.Equal(t, api.Trusted, appr.Verdict.Result)
	}
	// attest
	evidence = parseEvidence(t, "../../test/DESKTOP-MIMU51J-no-secureboot.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	appr, _, err = sendEvidence(t, &client, state, evidence)
	if assert.NoError(t, err) {
		assert.Equal(t, api.Vulnerable, appr.Verdict.Result)
	}

	_, status = srv.Post(t, "/v2/devices/1000/override", []string{
		issuesv1.TpmDummyId,
		issuesv1.UefiSecureBootVariablesId,
	}, apiToken)
	assert.Equal(t, http.StatusOK, status)

	// get device
	body, status := srv.Get(t, "/v2/devices/1000", testOrg)
	assert.Equal(t, http.StatusOK, status)
	dev := new(api.Device)
	err = jsonapi.UnmarshalPayload(bytes.NewBuffer(body), dev)
	if assert.NoError(t, err) {
		assert.Equal(t, api.StateTrusted, dev.State)
	}

	// attest
	evidence = parseEvidence(t, "../../test/DESKTOP-MIMU51J-no-secureboot.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	appr, _, err = sendEvidence(t, &client, state, evidence)
	if assert.NoError(t, err) {
		assert.Equal(t, api.Trusted, appr.Verdict.Result)
	}
}

func testIssue1176(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	ctx := context.Background()
	// enroll
	state := agstate.NewState()
	server := httptest.NewServer(srv.Handler)
	base, err := url.Parse(server.URL + "/v2/")
	assert.NoError(t, err)
	apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	client := agapi.Client{
		HTTP:               server.Client(),
		Base:               base,
		Auth:               apiToken,
		HTTPRequestTimeout: time.Second * 300,
		PostRequestTimeout: time.Second * 300,
		AgentVersion:       "dummy",
	}
	_, err = state.EnsureFresh(&client)
	assert.NoError(t, err)
	enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	attestationClient := agattest.NewCore()
	attestationClient.State = state
	attestationClient.Client = client
	attestationClient.Log = &log.Logger
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	attestationClient.StatePath = f.Name()
	err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
	assert.NoError(t, err)

	// attest
	evidence := parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	_, _, err = sendEvidence(t, &client, state, evidence)
	assert.NoError(t, err)

	_, status := srv.Post(t, "/v2/devices/1000/override", []string{
		issuesv1.TpmDummyId,
		issuesv1.TpmEndorsementCertUnverifiedId,
	}, apiToken)
	assert.Equal(t, http.StatusOK, status)

	evidence = parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	appr, _, err := sendEvidence(t, &client, state, evidence)
	if assert.NoError(t, err) {
		assert.Equal(t, api.Trusted, appr.Verdict.Result)
	}

	// attest w/o event log -> ok
	evidence = parseEvidence(t, "../../test/DESKTOP-MIMU51J-before.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	evidence.Firmware.TPM2EventLog.Data = nil
	appr, _, err = sendEvidence(t, &client, state, evidence)
	if assert.NoError(t, err) {
		assert.Equal(t, api.Trusted, appr.Verdict.Result)
	}
}

func parseEvidence(t *testing.T, f string) *agapi.Evidence {
	buf, err := os.ReadFile(f)
	assert.NoError(t, err)

	var ev agapi.Evidence
	err = json.Unmarshal(buf, &ev)
	assert.NoError(t, err)

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]agapi.Buffer{
			"11": ev.PCRs,
		}
	}

	return &ev
}

func sendEvidence(t *testing.T, client *agapi.Client, st *agstate.State, evidence *agapi.Evidence) (*agapi.Appraisal, string, error) {
	// transform firmware info into json and crypto-safe canonical json representations
	imaLog := evidence.Firmware.IMALog
	if imaLog != nil && mrand.Intn(1) > 0 {
		evidence.Firmware.IMALog = nil
	}

	// strip blobs for out-of-band transfer
	blobs := agapi.ProcessFirmwarePropertiesHashBlobs(&evidence.Firmware)

	fwPropsJSON, err := json.Marshal(evidence.Firmware)
	assert.NoError(t, err)
	evidence.Firmware.IMALog = imaLog
	fwPropsJCS, err := jcs.Transform(fwPropsJSON)
	assert.NoError(t, err)
	fwPropsHash := sha256.Sum256(fwPropsJCS)

	// load AIK
	aik, ok := st.Keys["aik"]
	assert.True(t, ok)
	assert.NotEmpty(t, aik.Private)
	qname, err := api.ComputeName(tpm2.HandleEndorsement, tpm2.Name(st.Root.Name), tpm2.Public(aik.Public))
	assert.NoError(t, err)

	somepriv, err := x509.ParsePKCS8PrivateKey([]byte(aik.Private))
	assert.NoError(t, err)
	priv, ok := somepriv.(*ecdsa.PrivateKey)
	assert.True(t, ok)

	signingHasher := sha256.New()

	var pcrSel []tpm2.PCRSelection
	if evidence.Quote == nil {
		pcrs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}
		for _, pcr := range pcrs {
			signingHasher.Write(evidence.PCRs[strconv.Itoa(pcr)])
		}
		pcrSel = []tpm2.PCRSelection{{
			Hash: tpm2.AlgSHA256,
			PCRs: pcrs,
		}}
	} else {
		assert.NotNil(t, evidence.Quote.AttestedQuoteInfo)
		for _, sel := range evidence.Quote.AttestedQuoteInfo.PCRSelection {
			bank := evidence.AllPCRs[strconv.Itoa(int(sel.Hash))]
			for _, pcr := range sel.PCRs {
				signingHasher.Write(bank[strconv.Itoa(pcr)])
			}
		}
		pcrSel = evidence.Quote.AttestedQuoteInfo.PCRSelection
	}
	attest := tpm2.AttestationData{
		Magic:           0xFF544347,
		Type:            tpm2.TagAttestQuote,
		ExtraData:       fwPropsHash[:],
		QualifiedSigner: tpm2.Name(qname),
		AttestedQuoteInfo: &tpm2.QuoteInfo{
			PCRSelection: pcrSel,
			PCRDigest:    signingHasher.Sum([]byte{}),
		},
	}

	attestBlob, err := attest.Encode()
	assert.NoError(t, err)
	attestHasher := sha256.New()
	attestHasher.Write(attestBlob)
	r, s, err := ecdsa.Sign(rand.Reader, priv, attestHasher.Sum([]byte{}))
	assert.NoError(t, err)
	sig := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{
			HashAlg: tpm2.AlgSHA256,
			R:       r,
			S:       s,
		},
	}

	point := (tpm2.Public(aik.Public)).ECCParameters.Point
	assert.Equal(t, 0, point.X().Cmp(priv.X))
	assert.Equal(t, 0, point.Y().Cmp(priv.Y))
	evidence.Quote = (*agapi.Attest)(&attest)
	evidence.Signature = (*agapi.Signature)(&sig)

	return client.Attest(context.Background(), aik.Credential, *evidence, blobs)
}

func testTags(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	body, status := srv.Get(t, "/v2/devices", testOrg)
	assert.Equal(t, http.StatusOK, status)

	devs, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Device)))
	assert.NoError(t, err)
	assert.Len(t, devs, 3)

	for _, iface := range devs {
		dev, ok := iface.(*api.Device)
		assert.True(t, ok)
		if dev.Id != "102" {
			assert.NotEmpty(t, dev.Tags)
		} else {
			assert.Empty(t, dev.Tags)
		}
	}
}

func testIssues(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	// tpm setup
	ctx := context.Background()
	state := agstate.NewState()

	// client setup
	server := httptest.NewServer(srv.Handler)
	base, err := url.Parse(server.URL + "/v2/")
	assert.NoError(t, err)
	apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	client := agapi.Client{
		HTTP:               server.Client(),
		Base:               base,
		Auth:               apiToken,
		HTTPRequestTimeout: time.Second * 300,
		PostRequestTimeout: time.Second * 300,
		AgentVersion:       "dummy",
	}

	// enroll
	_, err = state.EnsureFresh(&client)
	assert.NoError(t, err)
	enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	attestationClient := agattest.NewCore()
	attestationClient.State = state
	attestationClient.Client = client
	attestationClient.Log = &log.Logger
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	attestationClient.StatePath = f.Name()
	err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
	assert.NoError(t, err)

	// attest
	evidence := parseEvidence(t, "../../test/r340.evidence.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	_, link, err := sendEvidence(t, &client, state, evidence)
	assert.NoError(t, err)

	re := regexp.MustCompile(`^http:/app.example.com/devices/(\d+)\?.+$`)
	devid := re.FindStringSubmatch(link)[1]

	_, status := srv.Post(t, fmt.Sprintf("/v2/devices/%s/override", devid), []string{
		issuesv1.TpmDummyId,
		issuesv1.TpmEndorsementCertUnverifiedId,
	}, apiToken)
	assert.Equal(t, http.StatusOK, status)

	evidence = parseEvidence(t, "../../test/r340.evidence.json")
	evidence.Firmware.Flash.Data = nil
	evidence.Firmware.SMBIOS.Data = nil
	_, _, err = sendEvidence(t, &client, state, evidence)
	assert.NoError(t, err)

	// get device
	body, status := srv.Get(t, fmt.Sprint("/v2/devices/", devid), testOrg)
	assert.Equal(t, http.StatusOK, status)
	dev := new(api.Device)
	err = jsonapi.UnmarshalPayload(bytes.NewBuffer(body), dev)
	if assert.NoError(t, err) {
		assert.Equal(t, api.StateTrusted, dev.State)
	}
	fmt.Println(string(body))
}

var dashboardFixtures = []string{
	"latitude-7520.evidence.json",
	"IMN-SUPERMICRO-EPP-MSDefender.evidence.json",
	"sr630.evidence.json",
	"r340.evidence.json",
	"DESKTOP-MIMU51J-EPP-ESET.evidence.json",
	"h12ssl-2.evidence.json",
	"good-1.evidence.json",
	"sr630-no-ima-log.evidence.json",
	"IMN-SUPERMICRO-no-epp.evidence.json",
}

func testDashboard(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	riskStats := make(map[string]appraisal.RowRisksByType)
	var incidentCount, incidentDevs int
	for _, fix := range dashboardFixtures {
		t.Run(fix, func(t *testing.T) {
			// tpm setup
			ctx := context.Background()
			state := agstate.NewState()

			// client setup
			server := httptest.NewServer(srv.Handler)
			base, err := url.Parse(server.URL + "/v2/")
			assert.NoError(t, err)
			apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
			assert.NoError(t, err)
			client := agapi.Client{
				HTTP:               server.Client(),
				Base:               base,
				Auth:               apiToken,
				HTTPRequestTimeout: time.Second * 300,
				PostRequestTimeout: time.Second * 300,
				AgentVersion:       "dummy",
			}

			// enroll
			_, err = state.EnsureFresh(&client)
			assert.NoError(t, err)
			enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
			assert.NoError(t, err)
			attestationClient := agattest.NewCore()
			attestationClient.State = state
			attestationClient.Client = client
			attestationClient.Log = &log.Logger
			f, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			attestationClient.StatePath = f.Name()
			err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
			assert.NoError(t, err)

			// attest
			evidence := parseEvidence(t, fmt.Sprint("../../test/", fix))
			evidence.Firmware.Flash.Data = nil
			evidence.Firmware.SMBIOS.Data = nil
			_, link, err := sendEvidence(t, &client, state, evidence)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			re := regexp.MustCompile(`^http:/app.example.com/devices/(\d+)\?.+$`)
			devid := re.FindStringSubmatch(link)[1]

			// get issues via appraisals route; don't use device appraisal, it doesn't (and mustn't) have full issues definitions
			body, status := srv.Get(t, fmt.Sprint("/v2/devices/", devid, "/appraisals"), testOrg)
			assert.Equal(t, http.StatusOK, status)

			var many jsonapi.ManyPayload
			err = json.NewDecoder(bytes.NewReader(body)).Decode(&many)
			assert.NoError(t, err)

			buf, err := json.Marshal(many.Data[0].Attributes)
			assert.NoError(t, err)

			appr := api.Appraisal{}
			err = json.Unmarshal(buf, &appr)
			if !assert.NoError(t, err) || !assert.NotEmpty(t, appr) || !assert.NotNil(t, appr.Issues) {
				t.FailNow()
			}

			// collect stats
			incDevs := false
			for _, rawIssue := range appr.Issues.Issues {
				gi, err := issuesv1.GenericIssueFromRawMessage(rawIssue)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				if gi.Incident() {
					incDevs = true
					incidentCount++
				} else {
					stats, ok := riskStats[gi.Id()]
					if !ok {
						stats = appraisal.RowRisksByType{NumOccurences: 1}
					} else {
						stats.NumOccurences++
					}
					riskStats[gi.Id()] = stats
				}
			}
			if incDevs {
				incidentDevs++
			}
		})
	}

	// get dashboard
	body, status := srv.Get(t, "/v2/dashboard", testOrg)
	if !assert.Equal(t, http.StatusOK, status, string(body)) {
		t.FailNow()
	}

	var one jsonapi.OnePayload
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&one)
	assert.NoError(t, err)

	buf, err := json.Marshal(one.Data.Attributes)
	assert.NoError(t, err)

	apiDashboard := new(api.Dashboard)
	err = json.Unmarshal(buf, &apiDashboard)
	if !assert.NoError(t, err) || !assert.NotEmpty(t, apiDashboard) || !assert.NotEmpty(t, apiDashboard.Risks) {
		t.FailNow()
	}

	// assert we have same number of entries so we can just try to match each of the one with the other and be sure it is consistent
	assert.Equal(t, len(riskStats), len(apiDashboard.Risks))
	for _, ris := range apiDashboard.Risks {
		assert.Equal(t, riskStats[ris.IssueTypeId].NumOccurences, ris.NumOccurences)
	}

	assert.Equal(t, incidentCount, apiDashboard.IncidentCount)
	assert.Equal(t, incidentDevs, apiDashboard.IncidentDevCount)
}

func testIssue1576(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	// tpm setup
	ctx := context.Background()
	state := agstate.NewState()

	// client setup
	server := httptest.NewServer(srv.Handler)
	base, err := url.Parse(server.URL + "/v2/")
	assert.NoError(t, err)
	apiToken, err := authentication.IssueUserCredential("testsrv", testOrg, "Test User", time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	client := agapi.Client{
		HTTP:               server.Client(),
		Base:               base,
		Auth:               apiToken,
		HTTPRequestTimeout: time.Second * 300,
		PostRequestTimeout: time.Second * 300,
		AgentVersion:       "dummy",
	}

	// enroll
	_, err = state.EnsureFresh(&client)
	assert.NoError(t, err)
	enrollToken, err := authentication.IssueEnrollmentCredential("testsrv", testOrg, time.Now(), srv.Auth.Kid, srv.AuthPriv)
	assert.NoError(t, err)
	attestationClient := agattest.NewCore()
	attestationClient.State = state
	attestationClient.Client = client
	attestationClient.Log = &log.Logger
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	attestationClient.StatePath = f.Name()
	err = attestationClient.Enroll(ctx, enrollToken, true, "dummy")
	assert.NoError(t, err)

	// attest
	evidence := parseEvidence(t, "../../test/sr630.evidence.json")
	// hack to prevent Binarly & TSC jobs
	appr, _, err := sendEvidence(t, &client, state, evidence)
	assert.NoError(t, err)
	assert.Nil(t, appr)
}

func testFilterTags(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	body, status := srv.Get(t, "/v2/devices?filter[tags]=0ujtsYcgvSTl8PAuAdqWYSMnLOv", testOrg)
	assert.Equal(t, http.StatusOK, status)

	devs, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Device)))
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	dev, ok := devs[0].(*api.Device)
	assert.True(t, ok)
	assert.Equal(t, dev.Id, "100")
}

func testFilterState(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	body, status := srv.Get(t, "/v2/devices?filter[state]=vulnerable", testOrg)
	assert.Equal(t, http.StatusOK, status)

	devs, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Device)))
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	dev, ok := devs[0].(*api.Device)
	assert.True(t, ok)
	assert.Equal(t, dev.Id, "230")
}

func testFilterIssue(t *testing.T, conn *pgxpool.Pool, srv *testServer) {
	setupOrg(t, conn)

	body, status := srv.Get(t, "/v2/devices?filter[issue]=issue2", testOrg)
	assert.Equal(t, http.StatusOK, status)

	devs, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(body), reflect.TypeOf(new(api.Device)))
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	dev, ok := devs[0].(*api.Device)
	assert.True(t, ok)
	assert.Equal(t, dev.Id, "103")
}
