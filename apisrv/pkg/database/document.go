package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/jackc/pgtype"
)

var ErrIsNull = errors.New("document is null")

// A JSONB wrapper type that handles \u0000 and has shorthands for retrieving
// the 'type' field.
type Document struct {
	sanitized pgtype.JSONB
}

type typed struct {
	Type string `json:"type"`
}

var encodeTokens = regexp.MustCompile(`%|\\\\|(([^\\]?)\\u0000)`)

// \u0000 -> %00
// % -> %25
func encode(src []byte) ([]byte, error) {
	// 110%
	dst := make([]byte, 0, int(float32(len(src))*1.1))
	i := 0
	j := 0

	for i < len(src) {
		m := encodeTokens.FindSubmatchIndex(src[i:])
		if len(m) == 0 {
			return append(dst, src[i:]...), nil
		}

		// copy everything up to the match
		mi := m[0] + i
		dst = append(dst, src[i:mi]...)
		j += mi - i
		i = mi

		// one char lookahead
		var mmi int
		mmiok := len(src) > mi+1
		if mmiok {
			mmi = mi + 1
		}

		// decide the type of token matched
		percent := src[mi] == '%'
		escaped := src[mi] == '\\' && mmiok && src[mmi] == '\\'
		null := src[mi] == '\\' && mmiok && src[mmi] == 'u'
		null2 := src[mi] != '\\' && mmiok && src[mmi] == '\\'

		if escaped {
			dst = append(dst, '\\', '\\')
			i += 2
			j += 2
		} else if percent {
			dst = append(dst, '%', '2', '5')
			i += 1
			j += 3
		} else if null || null2 {
			if null2 {
				dst = append(dst, src[mi])
				j += 1
				i += 1
			}
			dst = append(dst, '%', '0', '0')
			i += 6
			j += 3
		} else {
			_, l := utf8.DecodeRune(src[mi:])
			if l == 0 {
				l = 1
			}
			dst = append(dst, src[mi:mi+l]...)
			i += l
			j += l
		}
	}

	return dst, nil
}

var decodeTokens = regexp.MustCompile(`%(00|25)`)

// \u0000 -> %00
// % -> %25
func decode(src []byte) ([]byte, error) {
	dst := make([]byte, 0, len(src))
	i := 0
	j := 0

	for i < len(src) {
		m := decodeTokens.FindSubmatchIndex(src[i:])
		if len(m) == 0 {
			return append(dst, src[i:]...), nil
		}

		// copy everything up to the match
		mi := m[0] + i
		dst = append(dst, src[i:mi]...)
		j += mi - i
		i = mi

		// one char lookahead
		var mmi int
		mmiok := len(src) > mi+1
		if mmiok {
			mmi = mi + 1
		}

		// decide the type of token matched
		percent := src[mi] == '%' && mmiok && src[mmi] == '2'
		null := src[mi] == '%' && mmiok && src[mmi] == '0'

		if percent {
			dst = append(dst, '%')
			i += 3
			j += 1
		} else if null {
			dst = append(dst, '\\', 'u', '0', '0', '0', '0')
			i += 3
			j += 6
		} else {
			_, l := utf8.DecodeRune(src[mi:])
			if l == 0 {
				l = 1
			}
			dst = append(dst, src[mi:mi+l]...)
			i += l
			j += l
		}
	}

	return dst, nil
}

func NewDocument(doc interface{}) (Document, error) {
	if doc == nil {
		return Document{sanitized: pgtype.JSONB{Status: pgtype.Null}}, nil
	}
	plain, err := json.Marshal(doc)
	if err != nil {
		return Document{}, fmt.Errorf("new doc marshal: %w", err)
	}
	sanitized, err := encode(plain)
	if err != nil {
		return Document{}, fmt.Errorf("new doc encode: %w", err)
	}
	return Document{sanitized: pgtype.JSONB{Status: pgtype.Present, Bytes: sanitized}}, nil
}

func NewDocumentRaw(doc []byte) (Document, error) {
	if doc == nil {
		return Document{sanitized: pgtype.JSONB{Status: pgtype.Null}}, nil
	}
	sanitized, err := encode(doc)
	if err != nil {
		return Document{}, err
	}
	return Document{sanitized: pgtype.JSONB{Status: pgtype.Present, Bytes: sanitized}}, nil
}

func (doc Document) Type() string {
	if doc.IsNull() {
		return ""
	}
	plain, err := doc.Bytes()
	if err != nil {
		return ""
	}
	var ty typed
	json.Unmarshal(plain, &ty)
	return ty.Type
}

func (doc Document) Bytes() ([]byte, error) {
	return decode(doc.sanitized.Bytes)
}

func (doc Document) SizeSanitized() int {
	return len(doc.sanitized.Bytes)
}

func (doc Document) Decode(s interface{}) error {
	if doc.IsNull() {
		return ErrIsNull
	}
	plain, err := decode(doc.sanitized.Bytes)
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, s)
}

func (doc Document) IsNull() bool {
	return doc.sanitized.Status != pgtype.Present
}

func (doc *Document) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return doc.sanitized.DecodeBinary(ci, src)
}

func (doc Document) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	return doc.sanitized.EncodeBinary(ci, buf)
}

func (Document) PreferredResultFormat() int16 {
	return pgtype.TextFormatCode
}

func (dst *Document) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	return dst.sanitized.DecodeText(ci, src)
}

func (Document) PreferredParamFormat() int16 {
	return pgtype.TextFormatCode
}

func (src Document) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.sanitized.EncodeText(ci, buf)
}

func (dst *Document) Scan(src interface{}) error {
	return dst.sanitized.Scan(src)
}

func (src Document) Value() (driver.Value, error) {
	return src.sanitized.Value()
}
