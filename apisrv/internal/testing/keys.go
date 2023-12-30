package testing

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/key"
)

func NewTestKeyset(iss string, t *testing.T) (*key.Set, *ecdsa.PrivateKey) {
	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key1pub := key1.Public().(*ecdsa.PublicKey)
	key1der, err := x509.MarshalPKIXPublicKey(key1pub)
	if err != nil {
		t.Fatal(err)
	}
	key1key, err := key.NewKey(iss, key1der)
	if err != nil {
		t.Fatal(err)
	}
	ks := key.NewSet()
	ks.Replace(&[]key.Key{key1key})

	return ks, key1
}

func NewTestKey(iss string, t *testing.T) (key.Key, *ecdsa.PrivateKey) {
	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key1pub := key1.Public().(*ecdsa.PublicKey)
	key1der, err := x509.MarshalPKIXPublicKey(key1pub)
	if err != nil {
		t.Fatal(err)
	}
	key1key, err := key.NewKey(iss, key1der)
	if err != nil {
		t.Fatal(err)
	}

	return key1key, key1
}
