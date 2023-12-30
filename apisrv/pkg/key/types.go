package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
)

type Key struct {
	Kid    string
	Issuer string
	Key    ecdsa.PublicKey
}

func ComputeKid(pub crypto.PublicKey) (string, error) {
	buf, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}

	cksum := sha256.Sum256(buf)
	return hex.EncodeToString(cksum[0:8]), nil
}

// SubjectPublicKeyInfo (RFC 5280, Section 4.1) ASN.1 DER
func NewKey(iss string, pkix []byte) (Key, error) {
	cksum := sha256.Sum256(pkix)
	kid := hex.EncodeToString(cksum[0:8])

	key, err := x509.ParsePKIXPublicKey(pkix)
	if err != nil {
		return Key{}, err
	}

	// must be a ECDSA key over NIST P-256
	if ec, ok := key.(*ecdsa.PublicKey); ok && ec.Curve == elliptic.P256() {
		return Key{Kid: kid, Key: *ec, Issuer: iss}, nil
	} else {
		return Key{}, fmt.Errorf("not a EC key over NIST P-256")
	}
}

func NewKeyBase64(iss string, b64 string) (Key, error) {
	buf, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return Key{}, err
	}
	return NewKey(iss, buf)
}

type Set struct {
	byIssuer *map[string][]Key
	byKid    *map[string]Key
	lock     *sync.RWMutex
}

func NewSet() *Set {
	iss := make(map[string][]Key)
	kid := make(map[string]Key)

	return &Set{
		byIssuer: &iss,
		byKid:    &kid,
		lock:     &sync.RWMutex{},
	}
}

func (k *Set) Key(kid string) (Key, bool) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	ks, ok := (*(k.byKid))[kid]
	return ks, ok
}

func (k *Set) KeysFor(iss string) []Key {
	k.lock.RLock()
	defer k.lock.RUnlock()

	if ks, ok := (*(k.byIssuer))[iss]; ok {
		return ks
	} else {
		return []Key{}
	}
}

func (k *Set) Equal(keys *[]Key) bool {
	k.lock.RLock()
	defer k.lock.RUnlock()

	if len(*keys) != len(*k.byKid) {
		return false
	}

	for _, kk := range *keys {
		if _, ok := (*k.byKid)[kk.Kid]; !ok {
			return false
		}
	}

	return true
}

func (k *Set) Replace(keys *[]Key) {
	kid := make(map[string]Key)
	iss := make(map[string][]Key)

	for _, key := range *keys {
		kid[key.Kid] = key

		// XXX: check for kid collision

		if car, ok := iss[key.Issuer]; ok {
			iss[key.Issuer] = append(car, key)
		} else {
			iss[key.Issuer] = []Key{key}
		}
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	k.byIssuer = &iss
	k.byKid = &kid
}

func (k *Set) Copy() []Key {
	k.lock.RLock()
	defer k.lock.RUnlock()

	ret := []Key{}
	for _, kk := range *k.byKid {
		ret = append(ret, kk)
	}

	return ret
}
