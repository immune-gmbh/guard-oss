package eventlog

import (
	"bytes"
	"encoding/binary"

	"github.com/google/go-tpm/tpm2"
)

func Marshal(alg uint16, events []TPMEvent) ([]byte, error) {
	w := new(bytes.Buffer)

	if err := binary.Write(w, binary.LittleEndian, uint32(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(NoAction)); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, [20]byte{}); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(0x25)); err != nil {
		return nil, err
	}
	if _, err := w.Write([]byte("Spec ID Event03\000")); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(0)); err != nil {
		return nil, err
	}
	if _, err := w.Write([]byte{0, 2, 2, 2}); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(2)); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, tpm2.AlgSHA1); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(20)); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, tpm2.AlgSHA256); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(32)); err != nil {
		return nil, err
	}
	if _, err := w.Write([]byte{0}); err != nil {
		return nil, err
	}

	for _, ev := range events {
		raw := ev.RawEvent()
		if err := binary.Write(w, binary.LittleEndian, uint32(raw.Index)); err != nil {
			return nil, err
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(raw.Type)); err != nil {
			return nil, err
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(1)); err != nil {
			return nil, err
		}
		if err := binary.Write(w, binary.LittleEndian, alg); err != nil {
			return nil, err
		}
		if _, err := w.Write(raw.Digest); err != nil {
			return nil, err
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(len(raw.Data))); err != nil {
			return nil, err
		}
		if _, err := w.Write(raw.Data); err != nil {
			return nil, err
		}
	}

	return w.Bytes(), nil
}
