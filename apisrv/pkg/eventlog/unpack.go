package eventlog

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/klauspost/compress/zstd"
)

func UnpackTPM2EventLogZ(eventLogZ []byte) ([][]byte, error) {
	var logBuffers [][]byte

	d, err := zstd.NewReader(bytes.NewReader(eventLogZ))
	if err != nil {
		return nil, err
	}
	defer d.Close()

	r := bytes.NewBuffer([]byte{})
	_, err = io.Copy(r, d)
	if err != nil {
		return nil, err
	}

	// each event log in a concatenated event log field is prefixed with its length
	for r.Len() > 0 {
		var logLen uint32
		if err := binary.Read(r, binary.LittleEndian, &logLen); err != nil {
			return nil, err
		}
		if logLen > uint32(r.Len()) {
			return nil, io.ErrUnexpectedEOF
		}
		singleLog := make([]byte, logLen)
		if l, err := r.Read(singleLog); err != nil {
			return nil, err
		} else if l != int(logLen) {
			return nil, io.ErrUnexpectedEOF
		}
		logBuffers = append(logBuffers, singleLog)
	}

	return logBuffers, nil
}
