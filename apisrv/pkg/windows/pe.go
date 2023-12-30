package windows

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/asn1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	saferwall "github.com/saferwall/pe"
)

type ELAMCertificateInfo struct {
	Hash        []byte
	Algorithm   hash.Hash
	ExtKeyUsage []asn1.ObjectIdentifier
}

type versionStruct struct {
	Length      uint16
	ValueLength uint16
	Type        uint16
	Key         string
	Value       []byte
	Children    []versionStruct
}

type vsFixedFileInfo struct {
	Signature        uint32
	StrucVersion     uint32
	FileVersionMS    uint32
	FileVersionLS    uint32
	ProductVersionMS uint32
	ProductVersionLS uint32
	FileFlagsMask    uint32
	FileFlags        uint32
	FileOS           uint32
	FileType         uint32
	FileSubtype      uint32
	FileDateMS       uint32
	FileDateLS       uint32
}

type Version struct {
	Vendor    string
	Product   string
	Version   string
	Component string
}

var dwordHexRe = regexp.MustCompile("^[0-9a-fA-F]{8}$")

type AlignableReader struct {
	address int
	inner   *bytes.Reader
}

func NewAlignedReader(address int, b []byte) *AlignableReader {
	return &AlignableReader{
		address: address,
		inner:   bytes.NewReader(b),
	}
}

func (rd *AlignableReader) Len() int {
	return rd.inner.Len()
}

func (rd *AlignableReader) Read(b []byte) (n int, err error) {
	n, err = rd.inner.Read(b)
	rd.address += n
	return
}

func (rd *AlignableReader) AlignTo(n int) error {
	if n <= 0 {
		return errors.New("alignment <= 0")
	}
	pad := rd.address % n
	if pad > 0 && pad < n {
		w, err := io.CopyN(io.Discard, rd.inner, int64(pad))
		if err != nil {
			return err
		}
		rd.address += int(w)
	}
	return nil
}

func parseVersion(rd *AlignableReader, size int, isString bool) (*versionStruct, error) {
	var ver versionStruct
	var words [3]uint16
	var l = rd.Len()
	err := binary.Read(rd, binary.LittleEndian, &words)
	if err != nil {
		return nil, err
	}
	ver.Key, err = readWideString(rd)
	if err != nil {
		return nil, err
	}
	childIsString := dwordHexRe.MatchString(ver.Key)
	if isString {
		words[1] = words[1] * 2
	}
	err = rd.AlignTo(4)
	if err != nil {
		return nil, err
	}
	ver.Value = make([]byte, int(words[1]))
	_, err = rd.Read(ver.Value)
	if err != nil {
		return nil, err
	}

	for !isString && size-(l-rd.Len()) > 0 {
		var ch *versionStruct
		ch, err = parseVersion(rd, size-(l-rd.Len()), childIsString)
		if err != nil {
			return nil, err
		}
		ver.Children = append(ver.Children, *ch)
		err = rd.AlignTo(4)
		if err != nil {
			return nil, err
		}
	}

	return &ver, nil
}

func readWideString(rd io.Reader) (string, error) {
	var wchars []uint16
	for len(wchars) == 0 || wchars[len(wchars)-1] != 0 {
		var wchar uint16
		err := binary.Read(rd, binary.LittleEndian, &wchar)
		if err != nil {
			return "", err
		}
		wchars = append(wchars, wchar)
	}

	if len(wchars) < 2 {
		return "", nil
	}

	var chars []byte
	for _, r := range utf16.Decode(wchars[0 : len(wchars)-1]) {
		chars = utf8.AppendRune(chars, r)
	}

	return string(chars), nil
}

func Parse(buf []byte) (*saferwall.File, error) {
	f, err := saferwall.NewBytes(buf, &saferwall.Options{})
	if err != nil {
		return nil, err
	}
	err = f.Parse()
	if err != nil {
		return nil, err
	}

	return f, nil
}

func GetELAMCertificateInfo(file *saferwall.File) ([]ELAMCertificateInfo, error) {
	for _, typ := range file.Resources.Entries {
		for _, name := range typ.Directory.Entries {
			for _, lang := range name.Directory.Entries {
				if typ.Name == "MSELAMCERTINFOID" && name.Name == "MICROSOFTELAMCERTIFICATEINFO" {
					var numEntries uint16

					buf, err := file.GetData(lang.Data.Struct.OffsetToData, lang.Data.Struct.Size)
					if err != nil {
						return nil, err
					}
					rd := bytes.NewBuffer(buf)
					err = binary.Read(rd, binary.LittleEndian, &numEntries)
					if err != nil {
						return nil, err
					}

					elamCerts := make([]ELAMCertificateInfo, numEntries)

					for entryIdx := 0; entryIdx < int(numEntries); entryIdx += 1 {
						var ent ELAMCertificateInfo

						fpr, err := readWideString(rd)
						if err != nil {
							return nil, err
						}
						ent.Hash, err = hex.DecodeString(fpr)
						if err != nil {
							return nil, err
						}

						var algo uint16
						err = binary.Read(rd, binary.LittleEndian, &algo)
						if err != nil {
							return nil, err
						}
						switch algo {
						case 32771:
							ent.Algorithm = md5.New()
						case 32772:
							ent.Algorithm = sha1.New()
						case 32780:
							ent.Algorithm = sha256.New()
						case 32781:
							ent.Algorithm = sha512.New384()
						case 32782:
							ent.Algorithm = sha512.New()
						default:
							return nil, errors.New("unknown hash algo")
						}

						ekus, err := readWideString(rd)
						if err != nil {
							return nil, err
						}
						if ekus != "" {
							for _, s := range strings.Split(ekus, ";") {
								si := strings.Split(s, ".")
								ii := make([]int, len(si))
								for i := range si {
									ii[i], err = strconv.Atoi(si[i])
									if err != nil {
										return nil, err
									}
								}
								ent.ExtKeyUsage = append(ent.ExtKeyUsage, asn1.ObjectIdentifier(ii))
							}
						}

						elamCerts[entryIdx] = ent
					}

					return elamCerts, nil
				}
			}
		}
	}

	return nil, errors.New("no embedded certificate info")
}

func GetVersion(file *saferwall.File) (*Version, error) {
	for _, typ := range file.Resources.Entries {
		for _, name := range typ.Directory.Entries {
			for _, lang := range name.Directory.Entries {
				if typ.ID == uint32(saferwall.RTVersion) {
					rva := lang.Data.Struct.OffsetToData
					buf, err := file.GetData(rva, lang.Data.Struct.Size)
					if err != nil {
						return nil, err
					}
					ver, err := parseVersion(NewAlignedReader(int(rva), buf), len(buf), false)
					if err != nil {
						return nil, err
					}

					if ver.Key == "VS_VERSION_INFO" {
						var fixed vsFixedFileInfo
						err = binary.Read(bytes.NewReader(ver.Value), binary.LittleEndian, &fixed)
						if err != nil {
							return nil, err
						}

						for _, ver := range ver.Children {
							if ver.Key == "StringFileInfo" {
								var ret Version

								for _, ver := range ver.Children {
									for _, ver := range ver.Children {
										if len(ver.Value) > 2 {
											val, err := readWideString(bytes.NewReader(ver.Value))
											if err != nil {
												return nil, err
											}

											switch ver.Key {
											case "FileDescription":
												ret.Component = val
											case "CompanyName":
												ret.Vendor = val
											case "ProductName":
												ret.Product = val
											case "ProductVersion":
												ret.Version = val
											}
										}
									}
								}

								if ret.Version != "" && ret.Product != "" && ret.Vendor != "" && ret.Component != "" {
									return &ret, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, errors.New("no version info")
}
