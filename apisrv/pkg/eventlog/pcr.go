package eventlog

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/google/go-tpm/tpm2"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
)

const (
	PCRMax = 24
)

type PCR struct {
	Index     int
	Digest    []byte
	DigestAlg crypto.Hash
}

func ImplementedPCRs(conn io.ReadWriteCloser) (tpm2.PCRSelection, error) {
	// GetCap PCR allocations
	prop, _, err := tpm2.GetCapability(conn, tpm2.CapabilityPCRs, 10, 0)
	if err != nil {
		return tpm2.PCRSelection{}, err
	}
	sels := []tpm2.PCRSelection{}

	for _, iface := range prop {
		sel := iface.(tpm2.PCRSelection)
		if len(sel.PCRs) > 0 {
			sels = append(sels, sel)
		}
	}

	sort.Slice(sels, func(i, j int) bool {
		return sels[i].Hash > sels[j].Hash
	})

	return sels[0], nil
}

func ReadPCRs(conn io.ReadWriteCloser, sel tpm2.PCRSelection) ([]PCR, error) {
	var pcrs []PCR
	data, err := tpm2.ReadPCRs(conn, sel)
	if err != nil {
		return nil, err
	}
	for k, v := range data {
		var p PCR
		p.Index = k
		p.DigestAlg, err = sel.Hash.Hash()
		if err != nil {
			return nil, err
		}
		p.Digest = make([]byte, len(v))
		copy(p.Digest, v)
		pcrs = append(pcrs, p)
	}
	return pcrs, nil
}

func ConvertPCRAPIBufferMapToEventLogPCRs(pcrBank map[string]api.Buffer, digestAlg crypto.Hash) ([]PCR, error) {
	var pcrs []PCR

	for k, v := range pcrBank {
		idx, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("parse pcr %w", err)
		}
		pcrs = append(pcrs, PCR{
			Index:     idx,
			DigestAlg: digestAlg,
			Digest:    []byte(v),
		})
	}

	return pcrs, nil
}

func ConvertPCRHexStringMapToEventLogPCRs(pcrBank map[string]string, digestAlg crypto.Hash) ([]PCR, error) {
	var pcrs []PCR

	for k, v := range pcrBank {
		idx, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("parse pcr idx %w", err)
		}
		vv, err := hex.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("parse pcr val %w", err)
		}
		pcrs = append(pcrs, PCR{
			Index:     idx,
			DigestAlg: digestAlg,
			Digest:    vv,
		})
	}

	return pcrs, nil
}
