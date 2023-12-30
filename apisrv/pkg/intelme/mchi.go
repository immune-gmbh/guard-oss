// Manufacturing and Configuration Host Inteface
// Extends MKHI group MCA

package intelme

import (
	"bytes"
	"context"
	"encoding/binary"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

const (
	MCHIGroupID = 0x0a

	MCHIReadFile       = 0x02
	MCHISetFile        = 0x03
	MCHICommitFile     = 0x04
	MCHICoreBIOSDone   = 0x05
	MCHIGetRPMCStatus  = 0x08
	MCHIReadFileEx     = 0x0a
	MCHISetFileEx      = 0x0b
	MCHIARBHSVNCommit  = 0x1b
	MCHIARBHSVNGetInfo = 0x1c
)

type MCHIHeader struct {
	GroupID  uint8
	Command  uint8
	Reserved uint8
	Result   uint8
}

func EncodeMCHIARBHSVNGetInfo() []byte {
	msg := MCHIHeader{
		GroupID:  MCHIGroupID,
		Command:  MCHIARBHSVNGetInfo,
		Reserved: 0,
		Result:   0,
	}

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, msg)
	return buf.Bytes()
}

func DecodeMCHIARBHSVNGetInfo(ctx context.Context, b []byte) ([]ARBHSVN, error) {
	var hdr MCHIHeader
	hdrlen := 4
	if len(b) < hdrlen {
		tel.Log(ctx).Errorf("reponse header too short")
		return nil, ErrFormat
	}
	err := binary.Read(bytes.NewReader(b[:hdrlen]), binary.LittleEndian, &hdr)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read MCHIHeader")
		return nil, ErrFormat
	}
	if hdr.GroupID != MCHIGroupID {
		tel.Log(ctx).Errorf("wrong group ID, want %d, got %d", MCHIGroupID, hdr.GroupID)
		return nil, ErrHeader
	}
	if hdr.Command&0x7f != ARBHSVNGetInfo {
		tel.Log(ctx).Errorf("wrong command ID, want %d, got %d", ARBHSVNGetInfo, hdr.Command&0x7f)
		return nil, ErrHeader
	}
	if hdr.Command&0x80 == 0 {
		tel.Log(ctx).Errorf("not a response message")
		return nil, ErrHeader
	}
	msglen := 4 + 4
	if len(b) < msglen {
		tel.Log(ctx).Errorf("size mismatch, want at least %d, got %d", msglen, len(b))
		return nil, ErrHeader
	}
	var num uint32
	if num > 1024 {
		tel.Log(ctx).Errorf("too many ARBHSVN entries")
		return nil, ErrHeader
	}
	entries := make([]ARBHSVN, num)
	err = binary.Read(bytes.NewReader(b[hdrlen:]), binary.LittleEndian, &entries)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read MCHIARBHSVNGetInfo entries")
		return nil, ErrFormat
	}

	return entries, err
}

func EncodeReadFileEx(fileId uint32, offset uint32, size uint32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, fileId)
	_ = binary.Write(buf, binary.LittleEndian, offset)
	_ = binary.Write(buf, binary.LittleEndian, size)
	_ = binary.Write(buf, binary.LittleEndian, uint8(0))

	return encodeMKHI(MCHIGroupID, ReadFileEx, buf.Bytes())
}

func DecodeReadFileEx(ctx context.Context, buf []byte) ([]byte, error) {
	data, err := decodeMKHI(ctx, MCHIGroupID, ReadFileEx, buf)
	if err != nil || len(data) < 4 {
		return nil, ErrFormat
	}

	var dataSize uint32
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &dataSize)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("read ReadFileEx dataSize")
		return nil, ErrFormat
	}

	if int(dataSize)+4 > len(data) {
		return nil, ErrFormat
	}

	return data[4 : int(dataSize)+4], nil
}
