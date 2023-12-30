package eventlog

import (
	"crypto"
	"fmt"

	"github.com/google/go-tpm/tpm2"
)

type HashAlg uint8

var (
	HashSHA1   = HashAlg(tpm2.AlgSHA1)
	HashSHA256 = HashAlg(tpm2.AlgSHA256)
	HashSHA384 = HashAlg(tpm2.AlgSHA384)
)

func (a HashAlg) CryptoHash() crypto.Hash {
	switch a {
	case HashSHA1:
		return crypto.SHA1
	case HashSHA256:
		return crypto.SHA256
	case HashSHA384:
		return crypto.SHA384
	}
	return 0
}

func (a HashAlg) GoTPMAlg() tpm2.Algorithm {
	switch a {
	case HashSHA1:
		return tpm2.AlgSHA1
	case HashSHA256:
		return tpm2.AlgSHA256
	case HashSHA384:
		return tpm2.AlgSHA384
	}
	return 0
}

// String returns a human-friendly representation of the hash algorithm.
func (a HashAlg) String() string {
	switch a {
	case HashSHA1:
		return "SHA1"
	case HashSHA256:
		return "SHA256"
	case HashSHA384:
		return "SHA384"
	}
	return fmt.Sprintf("HashAlg<%d>", int(a))
}
