package eventlog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func testParseEventLog(t *testing.T, testdata string) {
	data, err := ioutil.ReadFile(testdata)
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}
	var dump Dump
	if err := json.Unmarshal(data, &dump); err != nil {
		t.Fatalf("parsing test data: %v", err)
	}
	if _, err := ParseEventLog(dump.Log.Raw); err != nil {
		t.Fatalf("parsing event log: %v", err)
	}
}

func TestParseCryptoAgileEventLog(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/crypto_agile_eventlog")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}
	if _, err := ParseEventLog(data); err != nil {
		t.Fatalf("parsing event log: %v", err)
	}
}

func testEventLog(t *testing.T, testdata string) {
	data, err := ioutil.ReadFile(testdata)
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}
	var dump Dump
	if err := json.Unmarshal(data, &dump); err != nil {
		t.Fatalf("parsing test data: %v", err)
	}
	el, err := ParseEventLog(dump.Log.Raw)
	if err != nil {
		t.Fatalf("parsing event log: %v", err)
	}
	events, err := el.Verify(dump.Log.PCRs)
	if err != nil {
		t.Fatalf("validating event log: %v", err)
	}

	for i, e := range events {
		if e.Sequence != i {
			t.Errorf("event out of order: events[%d].sequence = %d, want %d", i, e.Sequence, i)
		}
	}
}

func TestParseEventLogEventSizeTooLarge(t *testing.T) {
	data := []byte{
		// PCR index
		0x30, 0x34, 0x39, 0x33,
		// type
		0x36, 0x30, 0x30, 0x32,

		// Digest
		0x31, 0x39, 0x36, 0x33, 0x39, 0x34, 0x34, 0x37, 0x39, 0x32,
		0x31, 0x32, 0x32, 0x37, 0x39, 0x30, 0x34, 0x30, 0x31, 0x6d,

		// Event size (3.183 GB)
		0xbd, 0xbf, 0xef, 0x47,

		// "event data"
		0x00, 0x00, 0x00, 0x00,
	}
	_, err := ParseEventLog(data)
	if err == nil {
		t.Fatalf("expected parsing invalid event log to fail")
	}
}

func TestParseSpecIDEvent(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []uint16
		wantErr bool
	}{
		{
			name: "sha1",
			data: append(
				[]byte("Spec ID Event03"), 0x0,
				0x0, 0x0, 0x0, 0x0, // platform class
				0x0,                // version minor
				0x2,                // version major
				0x0,                // errata
				0x8,                // uintn size
				0x1, 0x0, 0x0, 0x0, // num algs
				0x04, 0x0, // SHA1
				0x14, 0x0, // size
				0x2, // vendor info size
				0x0, 0x0,
			),
			want: []uint16{0x0004},
		},
		{
			name: "sha1_and_sha256",
			data: append(
				[]byte("Spec ID Event03"), 0x0,
				0x0, 0x0, 0x0, 0x0, // platform class
				0x0,                // version minor
				0x2,                // version major
				0x0,                // errata
				0x8,                // uintn size
				0x2, 0x0, 0x0, 0x0, // num algs
				0x04, 0x0, // SHA1
				0x14, 0x0, // size
				0x0B, 0x0, // SHA256
				0x20, 0x0, // size
				0x2, // vendor info size
				0x0, 0x0,
			),
			want: []uint16{0x0004, 0x000B},
		},
		{
			name: "invalid_version",
			data: append(
				[]byte("Spec ID Event03"), 0x0,
				0x0, 0x0, 0x0, 0x0, // platform class
				0x2,                // version minor
				0x1,                // version major
				0x0,                // errata
				0x8,                // uintn size
				0x2, 0x0, 0x0, 0x0, // num algs
				0x04, 0x0, // SHA1
				0x14, 0x0, // size
				0x0B, 0x0, // SHA256
				0x20, 0x0, // size
				0x2, // vendor info size
				0x0, 0x0,
			),
			wantErr: true,
		},
		{
			name: "malicious_number_of_algs",
			data: append(
				[]byte("Spec ID Event03"), 0x0,
				0x0, 0x0, 0x0, 0x0, // platform class
				0x0,                    // version minor
				0x2,                    // version major
				0x0,                    // errata
				0x8,                    // uintn size
				0xff, 0xff, 0xff, 0xff, // num algs
				0x04, 0x0, // SHA1
				0x14, 0x0, // size
				0x2, // vendor info size
				0x0, 0x0,
			),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := parseSpecIDEvent(test.data)
			var algs []uint16
			if (err != nil) != test.wantErr {
				t.Fatalf("parsing spec, wantErr=%t, got=%v", test.wantErr, err)
			}
			if err != nil {
				return
			}
			algsEq := func(want, got []uint16) bool {
				if len(got) != len(want) {
					return false
				}
				for i, alg := range got {
					if want[i] != alg {
						return false
					}
				}
				return true
			}

			for _, alg := range spec.algs {
				algs = append(algs, alg.ID)
			}

			if !algsEq(test.want, algs) {
				t.Errorf("algorithms, got=%x, want=%x", spec.algs, test.want)
			}
		})
	}
}

func TestParseEventLogEventSizeZero(t *testing.T) {
	data := []byte{
		// PCR index
		0x4, 0x0, 0x0, 0x0,

		// type
		0xd, 0x0, 0x0, 0x0,

		// Digest
		0x94, 0x2d, 0xb7, 0x4a, 0xa7, 0x37, 0x5b, 0x23, 0xea, 0x23,
		0x58, 0xeb, 0x3b, 0x31, 0x59, 0x88, 0x60, 0xf6, 0x90, 0x59,

		// Event size (0 B)
		0x0, 0x0, 0x0, 0x0,

		// no "event data"
	}

	if _, err := parseRawEvent(bytes.NewBuffer(data), nil); err != nil {
		t.Fatalf("parsing event log: %v", err)
	}
}

func TestParseShortNoAction(t *testing.T) {
	// https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClientSpecPlat_TPM_2p0_1p04_pub.pdf#page=110
	// says: "For EV_NO_ACTION events other than the EFI Specification ID event
	// (Section 9.4.5.1) the log will ...". Thus it is concluded other
	// than "EFI Specification ID" events are also valid as NO_ACTION events.
	//
	// Currently we just assume that such events will have Data shorter than
	// "EFI Specification ID" field.

	data, err := ioutil.ReadFile("testdata/short_no_action_eventlog")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}
	if _, err := ParseEventLog(data); err != nil {
		t.Fatalf("parsing event log: %v", err)
	}
}
