package eventlog

import (
	"bytes"
	"crypto"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseEvents(t *testing.T) {
	var emptyPCRs [PCRMax]PCR

	for i := range emptyPCRs {
		emptyPCRs[i].Index = i
		emptyPCRs[i].Digest = make([]byte, 20)
		emptyPCRs[i].DigestAlg = crypto.SHA1
	}
	testParseEvent(t, emptyPCRs[:], "testdata/binary_bios_measurements_15")
	testParseEvent(t, emptyPCRs[:], "testdata/binary_bios_measurements_27")
	testParseEvent(t, emptyPCRs[:], "testdata/linux_event_log")
	testParseEvent(t, emptyPCRs[:], "testdata/tpm12_windows_lenovo_x1carbonv3")
}

func TestParseCryptoAgileEvents(t *testing.T) {
	var emptyPCRs [PCRMax]PCR
	for i := range emptyPCRs {
		emptyPCRs[i].Index = i
		emptyPCRs[i].Digest = make([]byte, 32)
		emptyPCRs[i].DigestAlg = crypto.SHA256
	}

	testParseEvent(t, emptyPCRs[:], "testdata/crypto_agile_eventlog")
	testParseEvent(t, emptyPCRs[:], "testdata/tpm2_windows_lenovo_yogax1v2")
	testParseEvent(t, emptyPCRs[:], "testdata/windows_event_log")
}

func testParseEvent(t *testing.T, PCRs []PCR, filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("reading test data %s: %v", filename, err)
	}
	el, err := ParseEventLog(data)
	if err != nil {
		t.Fatalf("parsing event log %s: %v", filename, err)
	}
	outputEvents, err := el.Verify(PCRs[:])
	if err != nil {
		if replayErr, isReplayErr := err.(ReplayError); isReplayErr {
			outputEvents = replayErr.Events
		} else {
			t.Fatalf("failed to verify from event log %s: %v", filename, err)
		}
	}
	if len(outputEvents) == 0 {
		t.Fatalf("failed to extract any events from %s", filename)
	}

	parsedEvents, err := ParseEvents(outputEvents)

	if err != nil {
		t.Fatalf("parsing events %s: %v", filename, err)
	}

	if len(parsedEvents) == 0 {
		t.Fatalf("failed to parse any events from %s", filename)
	}

	reference := filename + ".json"
	referenceData, err := ioutil.ReadFile(reference)
	if err != nil {
		t.Fatalf("failed to read json reference %s: %v", reference, err)
	}

	parsedEventsJSON, err := json.MarshalIndent(parsedEvents, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal json for %s: %v", reference, err)
	}

	if string(parsedEventsJSON) != string(referenceData) {
		t.Fatalf("parsed events for %s don't match reference JSON", filename)
	}
}

func TestParseUEFIVariableData(t *testing.T) {
	data := []byte{0x61, 0xdf, 0xe4, 0x8b, 0xca, 0x93, 0xd2, 0x11, 0xaa, 0xd, 0x0, 0xe0, 0x98,
		0x3, 0x2b, 0x8c, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x53, 0x0, 0x65, 0x0, 0x63, 0x0, 0x75, 0x0, 0x72, 0x0,
		0x65, 0x0, 0x42, 0x0, 0x6f, 0x0, 0x6f, 0x0, 0x74, 0x0, 0x1}
	want := UEFIVariableData{
		Header: UEFIVariableDataHeader{
			VariableName:       EFIGUID{Data1: 0x8be4df61, Data2: 0x93ca, Data3: 0x11d2, Data4: [8]uint8{0xaa, 0xd, 0x0, 0xe0, 0x98, 0x3, 0x2b, 0x8c}},
			UnicodeNameLength:  0xa,
			VariableDataLength: 0x1,
		},
		UnicodeName:  []uint16{0x53, 0x65, 0x63, 0x75, 0x72, 0x65, 0x42, 0x6f, 0x6f, 0x74},
		VariableData: []uint8{0x1},
	}

	got, err := ParseUEFIVariableData(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseEFIVariableData() failed: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseUEFIVariableData() mismatch (-want +got):\n%s", diff)
	}
}
