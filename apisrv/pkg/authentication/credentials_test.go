package authentication

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	api "github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
	"github.com/stretchr/testify/assert"

	test "github.com/immune-gmbh/guard/apisrv/v2/internal/testing"
)

func TestServiceCredentials(t *testing.T) {
	keyset, priv := test.NewTestKeyset("apisrv", t)
	kid := keyset.KeysFor("apisrv")[0].Kid

	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("private key: %s\n", base64.StdEncoding.EncodeToString(pkcs8))
	fmt.Println("issued", time.Now().UTC())

	for i := 0; i < 3; i += 1 {
		sub := "CDAE9771-4761-46CD-883E-E0F30D9B40A8"
		tok, err := IssueServiceCredential("apisrv", &sub, time.Now().UTC(), kid, priv)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("token", tok)

		svc, org, err := VerifyServiceCredential(context.Background(), tok, keyset)
		if err != nil {
			t.Fatal(err)
		}

		if svc != "apisrv" {
			t.Fatal("wrong service")
		}

		if org == nil || *org != "CDAE9771-4761-46CD-883E-E0F30D9B40A8" {
			t.Fatal("wrong org")
		}
	}
}

func TestDeviceCredentials(t *testing.T) {
	keyset, priv := test.NewTestKeyset("apisrv", t)
	kid := keyset.KeysFor("apisrv")[0].Kid
	name1 := api.Name(api.GenerateName(rand.New(rand.NewSource(int64(os.Getpid())))))

	tok, err := IssueDeviceCredential("apisrv", name1, time.Now().UTC(), kid, priv)
	if err != nil {
		t.Fatal(err)
	}

	name2, err := VerifyDeviceCredential(context.Background(), tok, keyset)
	if err != nil {
		t.Fatal(err)
	}

	if !api.EqualNames(&name1, name2) {
		t.Fatal("wrong name")
	}
}

func TestValidateOrgCredential(t *testing.T) {
	serviceName := "authsrv"
	subject := "a3e85ce4-7687-5078-8716-c6c3dfaf554f"
	publicKey, err := ioutil.ReadFile("testdata/token-public-key")
	assert.NoError(t, err)
	token, err := ioutil.ReadFile("testdata/user-token")
	assert.NoError(t, err)

	k1buf, err := base64.StdEncoding.DecodeString(string(publicKey))
	assert.NoError(t, err)
	k1, err := key.NewKey(serviceName, k1buf)
	assert.NoError(t, err)
	ks := key.NewSet()
	ks.Replace(&[]key.Key{k1})

	authenticatedUser, err := VerifyUserCredential(context.Background(), string(token), ks)
	assert.NoError(t, err)

	assert.Equal(t, authenticatedUser.OrganizationExternal, subject)
	if strings.HasPrefix("tag:immu.ne,2021:user/", authenticatedUser.Actor) {
		t.Errorf("wrong name %s", authenticatedUser.Actor)
	}
}

func TestValidateServiceOrgCredential(t *testing.T) {
	serviceName := "authsrv"
	subject := "a3e85ce4-7687-5078-8716-c6c3dfaf554f"
	publicKey, err := ioutil.ReadFile("testdata/token-public-key")
	assert.NoError(t, err)
	token, err := ioutil.ReadFile("testdata/namespaced-service-token")
	assert.NoError(t, err)

	k1buf, err := base64.StdEncoding.DecodeString(string(publicKey))
	assert.NoError(t, err)

	k1, err := key.NewKey(serviceName, k1buf)
	assert.NoError(t, err)
	ks := key.NewSet()
	ks.Replace(&[]key.Key{k1})

	svc, org, err := VerifyServiceCredential(context.Background(), string(token), ks)
	assert.NoError(t, err)

	assert.NotNil(t, org)
	assert.Equal(t, *org, subject)
	assert.Equal(t, svc, serviceName)
}

func TestValidateServiceCredential(t *testing.T) {
	serviceName := "authsrv"
	publicKey, err := ioutil.ReadFile("testdata/token-public-key")
	assert.NoError(t, err)
	token, err := ioutil.ReadFile("testdata/service-token")
	assert.NoError(t, err)

	k1buf, err := base64.StdEncoding.DecodeString(string(publicKey))
	assert.NoError(t, err)
	k1, err := key.NewKey(serviceName, k1buf)
	assert.NoError(t, err)
	ks := key.NewSet()
	ks.Replace(&[]key.Key{k1})

	svc, org, err := VerifyServiceCredential(context.Background(), string(token), ks)
	assert.NoError(t, err)

	assert.Nil(t, org)
	assert.Equal(t, svc, serviceName)
}
