package inteltsc

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

func TestRealInstance(t *testing.T) {
	t.Skipf("Intel TSC test instance is down")
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	client := http.DefaultClient

	// get a oauth access token
	sess, err := openSession(ctx, &KaisTestCredentials, client)
	assert.NoError(t, err)

	fmt.Println("access token", sess)

	// get a oauth access token
	links, err := getFiles(ctx, &KaisTestCredentials, client, sess, "PF2B5BEE")
	assert.NoError(t, err)

	data, certs, err := downloadFile(ctx, client, links)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.NotEmpty(t, certs)

	fmt.Printf("%#v\n", data)
	fmt.Printf("%#v\n", certs[0])
}

func TestOpenSessionFailed(t *testing.T) {
	ctx := context.Background()
	site := Site{
		ClientId:     "blah",
		ClientSecret: "blub",
		Username:     "kai",
		Password:     "pass",
		Endpoint:     "http://localhost",
	}

	// refuse credentials
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(401)
		return resp.Result()
	})
	sess, err := openSession(ctx, &site, client)
	assert.Error(t, err)
	assert.Empty(t, sess)

	// send garbage
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.WriteString("Garbage")
		return resp.Result()
	})
	sess, err = openSession(ctx, &site, client)
	assert.Error(t, err)
	assert.Empty(t, sess)

	// send garbage, correct content type
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.WriteString("Garbage")
		return resp.Result()
	})
	sess, err = openSession(ctx, &site, client)
	assert.Error(t, err)
	assert.Empty(t, sess)

	// send wrong JSON
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.WriteString(`{"foo":"bar"}`)
		return resp.Result()
	})
	sess, err = openSession(ctx, &site, client)
	assert.Error(t, err)
	assert.Empty(t, sess)
}

func TestGetFiles(t *testing.T) {
	ctx := context.Background()
	site := Site{
		Vendor:       "test",
		ClientId:     "blah",
		ClientSecret: "blub",
		Username:     "kai",
		Password:     "pass",
		Endpoint:     "http://localhost",
	}

	// refuse
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(401)
		return resp.Result()
	})
	links, err := getFiles(ctx, &site, client, "session", "serial")
	assert.Error(t, err)
	assert.Empty(t, links)

	// send garbage
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.WriteString("Garbage")
		return resp.Result()
	})
	links, err = getFiles(ctx, &site, client, "session", "serial")
	assert.Error(t, err)
	assert.Empty(t, links)

	// send garbage, correct content type
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.WriteString("Garbage")
		return resp.Result()
	})
	links, err = getFiles(ctx, &site, client, "session", "serial")
	assert.Error(t, err)
	assert.Empty(t, links)

	// send empty list
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.WriteString(`[]`)
		return resp.Result()
	})
	links, err = getFiles(ctx, &site, client, "session", "serial")
	assert.Error(t, err)
	assert.Empty(t, links)

	// send empty list v2 api
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.WriteString(`{"started_at":"2022-08-03 05:30:44","certs":[]}`)
		return resp.Result()
	})
	links, err = getFiles(ctx, &site, client, "session", "serial")
	assert.Error(t, err)
	assert.Empty(t, links)

	// send real v2 answer
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.WriteString(`
      {
  "started_at": "2022-08-03 05:30:44",
  "certs": [
    {
      "serial": "J101WYR1",
      "zip_file": "J101WYR1.zip",
      "zip_content": [
        "LENOVO_DCG_126062_J101WYR1_DPD__J101WYR1_DPD_INTC_Platform_Data.xml",
        "LENOVO_DCG_126061_J101WYR1_PAC__J101WYR1_PCD_INTC_Platform_Cert_RSA.cer"
      ],
      "date": "2022-03-30 17:56:27",
      "updated": "2022-06-21 15:43:06",
      "link_zip": "https://tsc.intel.com/lenovo-dcg/data/owners/2022-08-03/68/J101WYR1.zip",
      "link_dpd": "https://tsc.intel.com/lenovo-dcg/data/owners/2022-08-03/68/LENOVO_DCG_126062_J101WYR1_DPD__J101WYR1_DPD_INTC_Platform_Data.xml",
      "link_pcd": "https://tsc.intel.com/lenovo-dcg/data/owners/2022-08-03/68/LENOVO_DCG_126061_J101WYR1_PAC__J101WYR1_PCD_INTC_Platform_Cert_RSA.cer",
      "history": {
        "LENOVO_DCG_126061_J101WYR1_PAC__J101WYR1_PCD_INTC_Platform_Cert_RSA.cer": "2022-03-30 17:56:26",
        "LENOVO_DCG_126062_J101WYR1_DPD__J101WYR1_DPD_INTC_Platform_Data.xml": "2022-03-30 17:56:26"
      }
    }
  ]
} `)
		return resp.Result()
	})
	links, err = getFiles(ctx, &site, client, "session", "serial")
	assert.NoError(t, err)
	assert.NotNil(t, links)
}

func TestDownloadFile(t *testing.T) {
	ctx := context.Background()

	// no link
	data, certs, err := downloadFile(ctx, nil, &getFilesCert{})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, certs)

	// download fails
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(401)
		return resp.Result()
	})
	data, certs, err = downloadFile(ctx, client, &getFilesCert{LinkZip: "http://example.com/blah.zip"})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, certs)

	// send garbage
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.WriteString("Garbage")
		return resp.Result()
	})
	data, certs, err = downloadFile(ctx, client, &getFilesCert{LinkZip: "http://example.com/blah.zip"})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, certs)

	// send empty zip
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		w.Close()
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.Write(buf.Bytes())
		return resp.Result()
	})
	data, certs, err = downloadFile(ctx, client, &getFilesCert{LinkZip: "http://example.com/blah.zip"})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, certs)

	// send xml only
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		fw, err := w.Create("Blub.xml")
		assert.NoError(t, err)
		fw.Write([]byte("Hello, World"))
		w.Close()
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.Write(buf.Bytes())
		return resp.Result()
	})
	data, certs, err = downloadFile(ctx, client, &getFilesCert{LinkZip: "http://example.com/blah.zip"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World", data)
	assert.Empty(t, certs)

	// send 2 xml files
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		fw, err := w.Create("Blub.xml")
		assert.NoError(t, err)
		fw.Write([]byte("Hello, World"))
		fw, err = w.Create("Blah.xml")
		assert.NoError(t, err)
		fw.Write([]byte("Hello, World"))
		w.Close()
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.Write(buf.Bytes())
		return resp.Result()
	})
	data, certs, err = downloadFile(ctx, client, &getFilesCert{LinkZip: "http://example.com/blah.zip"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World", data)
	assert.Empty(t, certs)

	// send invalid cert
	client = test.NewTestClient(func(req *http.Request) *http.Response {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		fw, err := w.Create("Blub.xml")
		assert.NoError(t, err)
		fw.Write([]byte("Hello, World"))
		fw, err = w.Create("Blah.cer")
		assert.NoError(t, err)
		fw.Write([]byte("Hello, World"))
		w.Close()
		resp := httptest.NewRecorder()
		resp.WriteHeader(200)
		resp.Write(buf.Bytes())
		return resp.Result()
	})
	data, certs, err = downloadFile(ctx, client, &getFilesCert{LinkZip: "http://example.com/blah.zip"})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, certs)
}
