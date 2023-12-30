package eventlog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"

	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

var (
	ErrFormat = errors.New("format error")
	ErrHash   = errors.New("wrong pcr bank")
)

type ImaReplayErr struct {
	InvalidPCRs map[string]string
}

func (ImaReplayErr) Error() string {
	return "replay err"
}

type imaEvent struct {
	Sequence int
	PCR      uint32
	Digest   [20]byte
	Name     string
	Data     []byte
}

func (e imaEvent) String() string {
	return fmt.Sprintf("PCR %02d [%040x] %s (%x)", e.PCR, e.Digest, e.Name, e.Data)
}

func (e imaEvent) RawEvent() Event {
	return Event{
		Sequence: e.Sequence,
		Index:    int(e.PCR),
		Type:     0xffffffff,

		Data:   e.Data,
		Digest: e.Digest[:],
		Alg:    HashSHA1,
	}
}

type ImaNgEvent struct {
	imaEvent
	Algo       string
	FileDigest []byte
	Path       string
	Signature  []byte
}

func (e ImaNgEvent) String() string {
	return fmt.Sprintf("PCR %02d [%040x] [%s%x] %s (%x)", e.PCR, e.Digest, e.Algo, e.FileDigest, e.Path, e.Signature)
}

func (e ImaNgEvent) RawEvent() Event {
	ev := Event{
		Sequence: e.Sequence,
		Index:    int(e.PCR),
		Type:     0xffffffff,

		Data:   e.Data,
		Digest: e.Digest[:],
		Alg:    HashSHA1,
	}
	return ev
}

//static struct ima_template_desc builtin_templates[] = {
//	{.name = IMA_TEMPLATE_IMA_NAME, .fmt = IMA_TEMPLATE_IMA_FMT},
//	{.name = "ima-ng", .fmt = "d-ng|n-ng"},
//	{.name = "ima-sig", .fmt = "d-ng|n-ng|sig"},
//	{.name = "ima-buf", .fmt = "d-ng|n-ng|buf"},
//	{.name = "ima-modsig", .fmt = "d-ng|n-ng|sig|d-modsig|modsig"},
//	{.name = "", .fmt = ""},	/* placeholder for a custom format */
//};
//
//static LIST_HEAD(defined_templates);
//static DEFINE_SPINLOCK(template_list);
//
//static const struct ima_template_field supported_fields[] = {
//	{.field_id = "d", .field_init = ima_eventdigest_init,
//	 .field_show = ima_show_template_digest},
//	{.field_id = "n", .field_init = ima_eventname_init,
//	 .field_show = ima_show_template_string},
//	{.field_id = "d-ng", .field_init = ima_eventdigest_ng_init,
//	 .field_show = ima_show_template_digest_ng},
//	{.field_id = "n-ng", .field_init = ima_eventname_ng_init,
//	 .field_show = ima_show_template_string},
//	{.field_id = "sig", .field_init = ima_eventsig_init,
//	 .field_show = ima_show_template_sig},
//	{.field_id = "buf", .field_init = ima_eventbuf_init,
//	 .field_show = ima_show_template_buf},
//	{.field_id = "d-modsig", .field_init = ima_eventdigest_modsig_init,
//	 .field_show = ima_show_template_digest_ng},
//	{.field_id = "modsig", .field_init = ima_eventmodsig_init,
//	 .field_show = ima_show_template_sig},
//};

///* print format:
// *       32bit-le=pcr#
// *       char[20]=template digest
// *       32bit-le=template name size
// *       char[n]=template name
// *       [eventdata length]
// *       eventdata[n]=template specific data
// */

func parseSizedBuffer(ctx context.Context, rd io.Reader, maxLen int) ([]byte, error) {
	var sz uint32

	err := binary.Read(rd, binary.LittleEndian, &sz)
	if err != nil {
		return nil, err
	}
	if int(sz) >= maxLen {
		tel.Log(ctx).WithFields(log.Fields{"want": sz, "got": maxLen}).Error("size past buffer")
		return nil, ErrFormat
	}

	buf := make([]byte, sz)
	_, err = io.ReadFull(rd, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func ParseIMA(ctx context.Context, binaryMeasurements []byte) ([]TPMEvent, error) {
	var ret []TPMEvent

	rd := bytes.NewReader(binaryMeasurements)
	for {
		var e imaEvent

		e.Sequence = len(ret)

		/*
		 * 1st: PCRIndex
		 * PCR used defaults to the same (config option) in
		 * little-endian format, unless set in policy
		 */
		err := binary.Read(rd, binary.LittleEndian, &e.PCR)
		if err == io.EOF {
			return ret, nil
		}
		if err != nil {
			return nil, err
		}

		/* 2nd: template digest */
		_, err = io.ReadFull(rd, e.Digest[:])
		if err != nil {
			return nil, err
		}

		/* 3rd: template name size */
		/* 4th:  template name */
		name, err := parseSizedBuffer(ctx, rd, len(binaryMeasurements))
		if err != nil {
			return nil, err
		}
		e.Name = string(name)
		if e.Name == "ima" {
			continue
		}

		/* 5th:  template length (except for 'ima' template) */
		/* 6th:  template specific data */
		e.Data, err = parseSizedBuffer(ctx, rd, len(binaryMeasurements))
		if err != nil {
			return nil, err
		}

		if e.Name == "ima-sig" || e.Name == "ima-ng" || e.Name == "ima-buf" {
			var e2 ImaNgEvent

			e2.Sequence = e.Sequence
			e2.PCR = e.PCR
			e2.Digest = e.Digest
			e2.Name = e.Name
			e2.Data = e.Data

			rd2 := bufio.NewReader(bytes.NewReader(e.Data))

			// d-ng
			var sz uint32
			err = binary.Read(rd2, binary.LittleEndian, &sz)
			if err != nil {
				return nil, err
			}
			if int(sz) >= len(e.Data) {
				return nil, ErrFormat
			}

			e2.Algo, err = rd2.ReadString(0)
			if err != nil {
				return nil, err
			}
			e2.Algo = strings.TrimSuffix(e2.Algo, "\000")

			e2.FileDigest = make([]byte, int(sz)-len(e2.Algo)-1)
			_, err = io.ReadFull(rd2, e2.FileDigest)
			if err != nil {
				return nil, err
			}

			// n-ng
			path, err := parseSizedBuffer(ctx, rd2, len(e.Data))
			if err != nil {
				return nil, err
			}
			e2.Path = strings.TrimSuffix(string(path), "\000")

			if e.Name == "ima-sig" {
				// sig
				e2.Signature, err = parseSizedBuffer(ctx, rd2, len(e.Data))
				if err != nil {
					return nil, err
				}
			}

			ret = append(ret, e2)
		} else {
			ret = append(ret, e)
		}
	}
}

// there is a race condition between reading the runtime measurement log and
// the PCRs: When new events are measured after reading the PCRs but before
// reading the log, the latter won't replay against the former because the new
// events are recored in the log but not in the PCRs. Our solution is to read
// (quote) the PCRs first and read the log after that. When verifying we replay
// the events until they match with the quoted PCR and return the log prefix up
// to that point. replays successfully.
func VerifyIMA(events []TPMEvent, bank map[string]string, algo hash.Hash) ([]TPMEvent, error) {
	verified := []TPMEvent{}
	computed := make(map[int][]byte)
	zero := make([]byte, 20)
	fox := append([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, make([]byte, algo.Size()-20)...)

outer:
	for _, ev := range events {
		algo.Reset()

		raw := ev.RawEvent()
		pcr, ok := computed[raw.Index]
		if !ok {
			pcr = make([]byte, algo.Size())
		}
		_, err := algo.Write(pcr)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(raw.Digest, zero) {
			_, err = algo.Write(fox)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = algo.Write(raw.Digest)
			if err != nil {
				return nil, err
			}
			_, err = algo.Write(make([]byte, algo.Size()-20))
			if err != nil {
				return nil, err
			}
		}
		computed[raw.Index] = algo.Sum(nil)

		// matching event log prefix found?
		verified = append(verified, ev)
		for pcr, val := range computed {
			p := bank[fmt.Sprint(pcr)]
			if p != hex.EncodeToString(val) {
				continue outer
			}
		}

		return verified, nil
	}

	err := ImaReplayErr{InvalidPCRs: make(map[string]string)}
	for p, v := range computed {
		err.InvalidPCRs[fmt.Sprint(p)] = fmt.Sprintf("%x", v)
	}
	return nil, err
}
