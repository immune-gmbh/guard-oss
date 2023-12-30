package baseline

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
)

var (
	ErrUnknownBank     = errors.New("unknown pcr bank")
	ErrDeserialization = errors.New("cannot deserialize hash")
)

type Hash struct {
	Sha1   *[20]byte
	Sha256 *[32]byte
}

func dupSha1(hash []byte) *[20]byte {
	var ret [20]byte
	copy(ret[:], hash)
	return &ret
}

func dupSha256(hash []byte) *[32]byte {
	var ret [32]byte
	copy(ret[:], hash)
	return &ret
}

func cmpSha1(a *[20]byte, b *[20]byte) bool {
	if a == nil || b == nil {
		return true
	}
	aa := *a
	bb := *b
	return bytes.Equal(aa[:], bb[:])
}

func cmpSha256(a *[32]byte, b *[32]byte) bool {
	if a == nil || b == nil {
		return true
	}
	aa := *a
	bb := *b
	return bytes.Equal(aa[:], bb[:])
}

func NewHash(sum []byte) (Hash, error) {
	switch len(sum) {
	case 0:
		return Hash{}, nil
	case 20:
		return Hash{Sha1: dupSha1(sum)}, nil
	case 32:
		return Hash{Sha256: dupSha256(sum)}, nil
	default:
		return Hash{}, ErrUnknownBank
	}
}

func (hash *Hash) String() string {
	if hash.IsUnset() {
		return "{}"
	} else {
		var sha1, sha256 []byte
		if hash.Sha1 != nil {
			sha1 = (*hash.Sha1)[:]
		}
		if hash.Sha256 != nil {
			sha256 = (*hash.Sha256)[:]
		}
		return fmt.Sprintf("{%x %x}", sha1, sha256)
	}
}

func (hash *Hash) CompareDigest(buf []byte) bool {
	ret := true
	if hash.Sha1 != nil {
		alg := sha1.New()
		alg.Write(buf)
		ret = bytes.Equal(alg.Sum(nil), (*hash.Sha1)[:]) && ret
	}
	if hash.Sha256 != nil {
		alg := sha256.New()
		alg.Write(buf)
		ret = bytes.Equal(alg.Sum(nil), (*hash.Sha256)[:]) && ret
	}
	return ret
}

func BeforeAfter(before, after *Hash) (string, string) {
	if before != nil && before.Sha256 != nil && after != nil && after.Sha256 != nil && !bytes.Equal(before.Sha256[:], after.Sha256[:]) {
		return fmt.Sprintf("%x", before.Sha256[:]), fmt.Sprintf("%x", after.Sha256[:])
	} else if before != nil && before.Sha1 != nil && after != nil && after.Sha1 != nil && !bytes.Equal(before.Sha1[:], after.Sha1[:]) {
		return fmt.Sprintf("%x", before.Sha1[:]), fmt.Sprintf("%x", after.Sha1[:])
	} else {
		var bstr, astr string
		if before != nil && before.Sha256 != nil {
			bstr = fmt.Sprintf("%x", before.Sha256[:])
		} else if before != nil && before.Sha1 != nil {
			bstr = fmt.Sprintf("%x", before.Sha1[:])
		}
		if after != nil && after.Sha256 != nil {
			astr = fmt.Sprintf("%x", after.Sha256[:])
		} else if after != nil && after.Sha1 != nil {
			astr = fmt.Sprintf("%x", after.Sha1[:])
		}

		return bstr, astr
	}
}

func (hash *Hash) IsUnset() bool {
	return hash.Sha1 == nil && hash.Sha256 == nil
}

func (hash *Hash) IntersectsWith(other *Hash) bool {
	return cmpSha1(hash.Sha1, other.Sha1) && cmpSha256(hash.Sha256, other.Sha256)
}

func (hash *Hash) UnionWith(other *Hash) bool {
	changed := false

	if hash.Sha1 == nil && other.Sha1 != nil {
		sha1 := [20]byte{}
		copy(sha1[:], (*other.Sha1)[:])
		hash.Sha1 = &sha1
		changed = true
	}
	if hash.Sha256 == nil && other.Sha256 != nil {
		sha256 := [32]byte{}
		copy(sha256[:], (*other.Sha256)[:])
		hash.Sha256 = &sha256
		changed = true
	}
	return changed
}

func (hash *Hash) ReplaceWith(other *Hash) bool {
	changed := false

	if other.Sha1 != nil {
		sha1 := [20]byte{}
		copy(sha1[:], (*other.Sha1)[:])
		hash.Sha1 = &sha1
		changed = true
	}
	if other.Sha256 != nil {
		sha256 := [32]byte{}
		copy(sha256[:], (*other.Sha256)[:])
		hash.Sha256 = &sha256
		changed = true
	}
	return changed
}

type serializedHash struct {
	SHA1   string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

func (a Hash) MarshalJSON() ([]byte, error) {
	ser := serializedHash{}
	if a.Sha1 != nil {
		ser.SHA1 = base64.StdEncoding.EncodeToString((*a.Sha1)[:])
	}
	if a.Sha256 != nil {
		ser.SHA256 = base64.StdEncoding.EncodeToString((*a.Sha256)[:])
	}
	return json.Marshal(ser)
}

func (a *Hash) UnmarshalJSON(data []byte) error {
	ser := serializedHash{}
	err := json.Unmarshal(data, &ser)
	if err != nil {
		return ErrDeserialization
	}

	hash := Hash{}
	if ser.SHA1 != "" {
		buf, err := base64.StdEncoding.DecodeString(ser.SHA1)
		if err != nil || len(buf) != 20 {
			return ErrDeserialization
		}
		hash.Sha1 = dupSha1(buf)
	}
	if ser.SHA256 != "" {
		buf, err := base64.StdEncoding.DecodeString(ser.SHA256)
		if err != nil || len(buf) != 32 {
			return ErrDeserialization
		}
		hash.Sha256 = dupSha256(buf)
	}

	*a = hash
	return nil
}

func EventDigest(e eventlog.TPMEvent) *Hash {
	hash, err := NewHash(e.RawEvent().Digest)
	if err == nil {
		return &hash
	} else {
		return &Hash{}
	}
}
