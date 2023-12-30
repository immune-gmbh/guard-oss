package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"testing"
)

func TestKeys(t *testing.T) {
	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key1pub := key1.Public().(*ecdsa.PublicKey)
	key1der, err := x509.MarshalPKIXPublicKey(key1pub)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewKey("aaa", key1der)
	if err != nil {
		t.Fatal(err)
	}
}
