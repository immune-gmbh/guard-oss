package inteltsc

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"

	acert "github.com/google/go-attestation/attributecert"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrUnknownContentType = errors.New("unknown content type")
	ErrIncompleteZip      = errors.New("incomplete zip file")

	// testing only
	KaisTestCredentials = Site{
		Vendor:       "Test",
		ClientId:     "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		ClientSecret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Username:     "xxxx.xxxx@xxx.xxx",
		Password:     "xxxxxxxxxxxxxxxxx",
		Endpoint:     "https://xxx.xxx.xxx/",
	}
)

type Site struct {
	Vendor       string
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Username     string
	Password     string
	Endpoint     string
}

type openSessionResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func openSession(ctx context.Context, site *Site, client *http.Client) (string, error) {
	form := make(url.Values)
	form.Add("grant_type", "password")
	form.Add("client_id", site.ClientId)
	form.Add("client_secret", site.ClientSecret)
	form.Add("username", site.Username)
	form.Add("password", site.Password)
	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	uri := fmt.Sprintf("%soauth/token", site.Endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, strings.NewReader(form.Encode()))
	if err != nil {
		tel.Log(ctx).WithError(err).Error("encode request")
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("request token")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tel.Log(ctx).WithField("http.status_code", resp.Status).Error("open session")
		return "", ErrNotFound
	}
	contentType := resp.Header.Get("content-type")
	if contentType != "application/json" {
		tel.Log(ctx).WithField("http.content_type", contentType).Error("open session")
		return "", ErrUnknownContentType
	}

	var res openSessionResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("decode open session")
		return "", err
	}

	if res.AccessToken == "" {
		return "", ErrUnknownContentType
	} else {
		return res.AccessToken, nil
	}
}

type getFilesResult struct {
	StartedAt string         `json:"started_at,omitempty"`
	Certs     []getFilesCert `json:"certs"`
}

type getFilesCert struct {
	Serial     string                 `json:"serial"`
	EkSerial   string                 `json:"ek_serial,omitempty"`
	ZipFile    string                 `json:"zip_file,omitempty"`
	LinkZip    string                 `json:"link_zip,omitempty"`
	LinkDPD    string                 `json:"link_dpd,omitempty"`
	LinkABD    string                 `json:"link_abd,omitempty"`
	LinkPAC    string                 `json:"link_pac,omitempty"`
	LinkSOC    string                 `json:"link_soc,omitempty"`
	ZipContent []string               `json:"zip_content,omitempty"`
	History    map[string]interface{} `json:"history,omitempty"`
	Date       string                 `json:"date,omitempty"`
	Updated    string                 `json:"updated,omitempty"`
}

func getFiles(ctx context.Context, site *Site, client *http.Client, session string, serial string) (*getFilesCert, error) {
	vals := make(url.Values)
	vals.Add("include", serial)
	vals.Add("get_files", "true")
	vals.Add("access_token", session)
	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	uri := fmt.Sprintf("%sapi/tsc/v1/serials?%s", site.Endpoint, vals.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("encode request")
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("get files")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tel.Log(ctx).WithField("http.status_code", resp.Status).Error("get files")
		return nil, ErrNotFound
	}
	contentType, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if err != nil || contentType != "application/json" {
		tel.Log(ctx).WithError(err).WithField("http.content_type", contentType).Error("get files")
		return nil, ErrUnknownContentType
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("reading body")
		return nil, err
	}

	var resv1 []getFilesCert
	var resv2 getFilesResult
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&resv1)
	if err != nil {
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&resv2)
		if err != nil {
			tel.Log(ctx).WithError(err).Error("decode get files")
			return nil, err
		}
		resv1 = resv2.Certs
	}
	if len(resv1) == 0 {
		return nil, ErrNotFound
	}

	return &resv1[0], nil
}

func downloadFile(ctx context.Context, client *http.Client, links *getFilesCert) (string, []*acert.AttributeCertificate, error) {
	if links.LinkZip == "" {
		return "", nil, ErrIncompleteZip
	}

	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, links.LinkZip, nil)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("encode request")
		return "", nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("download zip")
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		tel.Log(ctx).WithField("http.status_code", resp.StatusCode).Error("download zip")
		return "", nil, ErrIncompleteZip
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read body")
		return "", nil, err
	}

	return UnpackZip(ctx, body)
}

// only public for use in test
func UnpackZip(ctx context.Context, file []byte) (string, []*acert.AttributeCertificate, error) {
	archive, err := zip.NewReader(bytes.NewReader(file), int64(len(file)))
	if err != nil {
		tel.Log(ctx).WithError(err).Error("open zip")
		return "", nil, err
	}

	xmldoc := ""
	certs := make([]*acert.AttributeCertificate, 0)
	for _, f := range archive.File {
		if strings.HasSuffix(f.Name, ".xml") {
			if xmldoc == "" {
				fd, err := f.Open()
				if err != nil {
					tel.Log(ctx).WithField("file", f.Name).WithError(err).Error("scan zip")
					return "", nil, err
				}
				buf, err := ioutil.ReadAll(fd)
				if err != nil {
					tel.Log(ctx).WithField("file", f.Name).WithError(err).Error("read xml")
					return "", nil, err
				}
				xmldoc = string(buf)
			} else {
				tel.Log(ctx).WithField("file", f.Name).Error("dup xml file")
				continue
			}
		} else if strings.HasSuffix(f.Name, ".cer") {
			fd, err := f.Open()
			if err != nil {
				tel.Log(ctx).WithField("file", f.Name).WithError(err).Error("scan zip")
				return "", nil, err
			}
			buf, err := ioutil.ReadAll(fd)
			if err != nil {
				tel.Log(ctx).WithField("file", f.Name).WithError(err).Error("read cert")
				return "", nil, err
			}
			cert, err := acert.ParseAttributeCertificate(buf)
			if err != nil {
				tel.Log(ctx).WithField("file", f.Name).WithError(err).Error("parse cert")
				return "", nil, err
			}

			certs = append(certs, cert)
		}
	}

	if xmldoc == "" {
		return "", nil, ErrIncompleteZip
	}
	return xmldoc, certs, nil
}
immu.ne