package eventlog

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/x509"
)

// SecurebootState describes the secure boot status of a machine, as determined
// by processing its event log.
type SecurebootState struct {
	Enabled bool

	// PlatformKeys enumerates keys which can sign a key exchange key.
	PlatformKeys []x509.Certificate
	// PlatformKeys enumerates key hashes which can sign a key exchange key.
	PlatformKeyHashes [][]byte

	// ExchangeKeys enumerates keys which can sign a database of permitted or
	// forbidden keys.
	ExchangeKeys []x509.Certificate
	// ExchangeKeyHashes enumerates key hashes which can sign a database or
	// permitted or forbidden keys.
	ExchangeKeyHashes [][]byte

	// PermittedKeys enumerates keys which may sign binaries to run.
	PermittedKeys []x509.Certificate
	// PermittedHashes enumerates hashes which permit binaries to run.
	PermittedHashes [][]byte

	// ForbiddenKeys enumerates keys which must not permit a binary to run.
	ForbiddenKeys []x509.Certificate
	// ForbiddenKeys enumerates hashes which must not permit a binary to run.
	ForbiddenHashes [][]byte

	// PreSeparatorAuthority describes the use of a secure-boot key to authorize
	// the execution of a binary before the separator.
	PreSeparatorAuthority []x509.Certificate
	// PostSeparatorAuthority describes the use of a secure-boot key to authorize
	// the execution of a binary after the separator.
	PostSeparatorAuthority []x509.Certificate
}

// ParseSecurebootState parses a series of events to determine the
// configuration of secure boot on a device. An error is returned if
// the state cannot be determined, or if the event log is structured
// in such a way that it may have been tampered post-execution of
// platform firmware.
func ParseSecurebootState(events []Event) (*SecurebootState, error) {
	// This algorithm verifies the following:
	// - All events in PCR 7 have event types which are expected in PCR 7.
	// - All events are parsable according to their event type.
	// - All events have digests values corresponding to their data/event type.
	// - No unverifiable events were present.
	// - All variables are specified before the separator and never duplicated.
	// - The SecureBoot variable has a value of 0 or 1.
	// - If SecureBoot was 1 (enabled), authority events were present indicating
	//   keys were used to perform verification.
	// - If SecureBoot was 1 (enabled), platform + exchange + database keys
	//   were specified.
	// - No UEFI debugger was attached.

	var (
		out           SecurebootState
		seenSeparator = map[HashAlg]bool{}
		seenAuthority = map[HashAlg]bool{}
		seenVars      = map[string]map[HashAlg]bool{}
	)

	for _, e := range events {
		if e.Index != 7 {
			continue
		}

		et, err := UntrustedParseEventType(uint32(e.Type))
		if err != nil {
			return nil, fmt.Errorf("unrecognised event type: %v", err)
		}

		digestVerify := e.DigestEquals(e.Data)
		switch et {
		case Separator:
			if _, ok := seenSeparator[e.Alg]; ok {
				return nil, fmt.Errorf("duplicate separator at event %d", e.Sequence)
			}
			seenSeparator[e.Alg] = true
			if !bytes.Equal(e.Data, []byte{0, 0, 0, 0}) {
				return nil, fmt.Errorf("invalid separator data at event %d: %v", e.Sequence, e.Data)
			}
			if digestVerify != nil {
				return nil, fmt.Errorf("invalid separator digest at event %d: %v", e.Sequence, digestVerify)
			}

		case EFIAction:
			if string(e.Data) == "UEFI Debug Mode" {
				return nil, errors.New("a UEFI debugger was present during boot")
			}
			return nil, fmt.Errorf("event %d: unexpected EFI action event", e.Sequence)

		case EFIVariableDriverConfig:
			v, err := ParseUEFIVariableData(bytes.NewReader(e.Data))
			if err != nil {
				return nil, fmt.Errorf("failed parsing EFI variable at event %d: %v", e.Sequence, err)
			}
			if vars, seenBefore := seenVars[v.VarName()]; seenBefore {
				if _, seenBefore = vars[e.Alg]; seenBefore {
					return nil, fmt.Errorf("duplicate EFI variable %q at event %d", v.VarName(), e.Sequence)
				} else {
					seenVars[v.VarName()][e.Alg] = true
				}
			} else {
				seenVars[v.VarName()] = map[HashAlg]bool{e.Alg: true}
			}
			if _, ok := seenSeparator[e.Alg]; ok {
				return nil, fmt.Errorf("event %d: variable %q specified after separator", e.Sequence, v.VarName())
			}

			if digestVerify != nil {
				return nil, fmt.Errorf("invalid digest for variable %q on event %d: %v", v.VarName(), e.Sequence, digestVerify)
			}

			switch v.VarName() {
			case "SecureBoot":
				if len(v.VariableData) != 1 {
					return nil, fmt.Errorf("event %d: SecureBoot data len is %d, expected 1", e.Sequence, len(v.VariableData))
				}
				out.Enabled = v.VariableData[0] == 1
			case "PK":
				if out.PlatformKeys, out.PlatformKeyHashes, err = v.SignatureData(); err != nil {
					return nil, fmt.Errorf("event %d: failed parsing platform keys: %v", e.Sequence, err)
				}
			case "KEK":
				if out.ExchangeKeys, out.ExchangeKeyHashes, err = v.SignatureData(); err != nil {
					return nil, fmt.Errorf("event %d: failed parsing key exchange keys: %v", e.Sequence, err)
				}
			case "db":
				if out.PermittedKeys, out.PermittedHashes, err = v.SignatureData(); err != nil {
					return nil, fmt.Errorf("event %d: failed parsing signature database: %v", e.Sequence, err)
				}
			case "dbx":
				if out.ForbiddenKeys, out.ForbiddenHashes, err = v.SignatureData(); err != nil {
					return nil, fmt.Errorf("event %d: failed parsing forbidden signature database: %v", e.Sequence, err)
				}
			}

		case EFIVariableAuthority:
			v, err := ParseUEFIVariableData(bytes.NewReader(e.Data))
			if err != nil {
				return nil, fmt.Errorf("failed parsing UEFI variable data: %v", err)
			}

			a, err := ParseUEFIVariableAuthority(v)
			if err != nil {
				// Workaround for: https://github.com/google/go-attestation/issues/157
				if err == ErrSigMissingGUID {
					// Versions of shim which do not carry
					// https://github.com/rhboot/shim/commit/8a27a4809a6a2b40fb6a4049071bf96d6ad71b50
					// have an erroneous additional byte in the event, which breaks digest
					// verification. If verification failed, we try removing the last byte.
					if digestVerify != nil && len(e.Data) > 0 {
						digestVerify = e.DigestEquals(e.Data[:len(e.Data)-1])
					}
				} else {
					return nil, fmt.Errorf("failed parsing EFI variable authority at event %d: %v", e.Sequence, err)
				}
			}
			seenAuthority[e.Alg] = true
			if digestVerify != nil {
				return nil, fmt.Errorf("invalid digest for authority on event %d: %v", e.Sequence, digestVerify)
			}
			if _, ok := seenSeparator[e.Alg]; !ok {
				out.PreSeparatorAuthority = append(out.PreSeparatorAuthority, a.Certs...)
			} else {
				out.PostSeparatorAuthority = append(out.PostSeparatorAuthority, a.Certs...)
			}

		default:
			return nil, fmt.Errorf("unexpected event type: %v", et)
		}
	}
	if !out.Enabled {
		return &out, nil
	}
	if len(seenAuthority) == 0 {
		return nil, errors.New("secure boot was enabled but no key was used")
	}
	if len(out.PlatformKeys) == 0 && len(out.PlatformKeyHashes) == 0 {
		return nil, errors.New("secure boot was enabled but no platform keys were known")
	}
	if len(out.ExchangeKeys) == 0 && len(out.ExchangeKeyHashes) == 0 {
		return nil, errors.New("secure boot was enabled but no key exchange keys were known")
	}
	if len(out.PermittedKeys) == 0 && len(out.PermittedHashes) == 0 {
		return nil, errors.New("secure boot was enabled but no keys or hashes were permitted")
	}
	return &out, nil
}
