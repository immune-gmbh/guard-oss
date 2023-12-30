// Intel SPS HECI client

package intelme

import (
	"bytes"
	"encoding/binary"
)

const (
	SPSGetMEBIOSInterface = 0x1
	SPSGetVendorLabel     = 0x2
)

type SPSRequest struct {
	Heci    Hecihdr
	Command uint8
}

type SPSGetMEBIOSResponse struct {
	Header       Hecihdr
	MajorVersion uint8
	MinorVersion uint8
	Features1    uint32
	Features2    uint32
}

func DecodeSPSGetBios(b []byte) (*SPSGetMEBIOSResponse, error) {
	var resp SPSGetMEBIOSResponse
	reader := bytes.NewReader(b)
	if err := binary.Read(reader, binary.LittleEndian, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func EncodeSPSGetBios() []byte {
	var hdr Hecihdr
	hdr.SetMEAddr(HECISPS)
	hdr.SetHostAddr(0x0)
	hdr.SetLength(0x1)
	hdr.SetMsgComplete(true)
	msg := SPSRequest{
		Heci:    hdr,
		Command: SPSGetMEBIOSInterface,
	}
	writer := new(bytes.Buffer)
	if err := binary.Write(writer, binary.LittleEndian, msg); err != nil {
		panic(err)
	}
	return writer.Bytes()
}

type SPSGetMEVendorLabelResponse struct {
	Heci        Hecihdr
	Command     uint8
	Reserved    [3]byte
	VendorLabel [4]byte
	Signature   [32]byte
}

func DecodeSPSGetVendorLabel(b []byte) (*SPSGetMEVendorLabelResponse, error) {
	var resp SPSGetMEVendorLabelResponse
	reader := bytes.NewReader(b)
	if err := binary.Read(reader, binary.LittleEndian, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func EncodeSPSGetVendorLabel() []byte {
	var hdr Hecihdr
	hdr.SetMEAddr(HECISPS)
	hdr.SetHostAddr(0x0)
	hdr.SetLength(0x1)
	hdr.SetMsgComplete(true)
	msg := SPSRequest{
		Heci:    hdr,
		Command: SPSGetVendorLabel,
	}
	writer := new(bytes.Buffer)
	if err := binary.Write(writer, binary.LittleEndian, msg); err != nil {
		panic(err)
	}
	return writer.Bytes()
}
