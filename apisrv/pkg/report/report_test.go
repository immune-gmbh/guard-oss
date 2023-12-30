package report

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
)

var (
	evidenceFiles []string = []string{
		"../../test/cave02.9e.network.evidence.json",
		"../../test/cave02.nopcr.evidence.json",
		"../../test/vision.9elements.com.evidence.json",
		"../../test/test.evidence.json",
		"../../test/IMN-DELL.evidence.json",
		"../../test/IMN-SUPERMICRO.evidence.json",
		"../../test/sr630.evidence.json",
		"../../test/ludmilla.evidence.json",
	}

	reportDirname string = "../../test"
)

func parseEvidence(t *testing.T, f string) *evidence.Values {
	buf, err := ioutil.ReadFile(f)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]api.Buffer{
			"11": ev.PCRs,
		}
	}

	val, err := evidence.WrapInsecure(&ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return val
}

type scenario struct {
	Evidence *evidence.Values
}

type withOnDiskEvidence struct {
	Path string
}

// UEFIVariables
type withMissingEfivars struct{}
type withoutEfivarfs struct{}
type withInaccassibleEfivars struct{}

func setupScenario(t *testing.T, opts ...interface{}) scenario {
	ev := parseEvidence(t, "../../test/IMN-DELL.evidence.json")

	for _, opt := range opts {
		switch opt.(type) {
		case withOnDiskEvidence:
			ev = parseEvidence(t, opt.(withOnDiskEvidence).Path)
		case withMissingEfivars:
			ev.UEFIVariables = []api.UEFIVariable{
				{
					Vendor: api.EFIGlobalVariable.String(),
					Name:   "PK",
					Error:  api.NoResponse,
				}, {
					Vendor: api.EFIGlobalVariable.String(),
					Name:   "KEK",
					Error:  api.NoResponse,
				},
			}
		case withoutEfivarfs:
			ev.UEFIVariables = []api.UEFIVariable{
				{
					Vendor: api.EFIGlobalVariable.String(),
					Name:   "PK",
					Error:  api.NotImplemented,
				}, {
					Vendor: api.EFIGlobalVariable.String(),
					Name:   "KEK",
					Error:  api.NotImplemented,
				},
			}
		case withInaccassibleEfivars:
			ev.UEFIVariables = []api.UEFIVariable{
				{
					Vendor: api.EFIGlobalVariable.String(),
					Name:   "PK",
					Error:  api.NoPermission,
				}, {
					Vendor: api.EFIGlobalVariable.String(),
					Name:   "KEK",
					Error:  api.NoPermission,
				},
			}
		default:
			t.Fatalf("unknown opt %v", opt)
		}
	}

	return scenario{
		Evidence: ev,
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestAnalysis(t *testing.T) {
	evidence := make(map[string]*evidence.Values)

	for _, s := range evidenceFiles {
		evidence[path.Base(s)] = parseEvidence(t, s)
	}

	for s, ev := range evidence {
		t.Run(s, func(t *testing.T) {
			report, err := Compile(context.Background(), ev)
			if err != nil {
				t.Fatal(err)
			}

			buf, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			switch s {
			case "cave02.nopcr.evidence.json":
				if report.Values.TPM != nil && len(report.Values.TPM.PCR) > 0 {
					t.Error("PCR no empty")
				}
			case "test.evidence.json":
				if len(report.Values.NICs) != 0 {
					t.Error("expected zero NICs")
				}
			case "Haruhi.json":
				fmt.Println(string(buf))
				//pk := &(*report.Values.UEFI.PlatformKeys)[0]
				//if pk.Type != api.EFICertificate || *pk.Subject != "CN=Platform Key,C=Platform Key" {
				//	t.Error("wrong pk")
				//}
			case "yv3-pvt-slot4.json":
				fmt.Println(string(buf))
				if report.Values.ME.State != "insecure" {
					t.Error("expected manufacturing mode")
				}
			case "x1carbon.json":
				fmt.Println(string(buf))
				if report.Values.ME.Version[0] != 15 {
					t.Error("expected empty version 15 ME")
				}
			case "bootstrap.json":
				if report.Values.UEFI != nil {
					t.Error("expected empty UEFI")
				}
				if report.Values.TXT != nil {
					t.Error("expected empty TXT")
				}
				if report.Values.TPM != nil {
					t.Error("expected empty TPM")
				}
				if report.Values.SEV != nil {
					t.Error("expected empty SEV")
				}
				if report.Values.SGX != nil {
					t.Error("expected empty SGX")
				}
				if report.Values.ME.Version[0] != 8 {
					t.Error("expected empty version 8 ME")
				}
			case "cave07.json":
				if report.Values.UEFI != nil {
					t.Error("expected empty UEFI")
				}
				if report.Values.TXT != nil {
					t.Error("expected empty TXT")
				}
				if report.Values.TPM != nil {
					t.Error("expected empty TPM")
				}
				if report.Values.SEV != nil {
					t.Error("expected empty SEV")
				}
				if report.Values.SGX != nil {
					t.Error("expected empty SGX")
				}
				if report.Values.ME.Version[0] != 11 {
					t.Error("expected empty version 11 ME")
				}
			case "ludmilla.evidence.json":
				if len(report.Values.NICs) == 0 {
					t.Error("expected some NICs")
				}
				if report.Values.SMBIOS.Serial != "PC1SC0VK" {
					t.Error("expected SMBIOS serial PC1SC0VK")
				}
				if report.Values.SMBIOS.UUID != "b112304c-28f2-11b2-a85c-fb492b570fd7" {
					t.Error("expected SMBIOS UUID b112304c-28f2-11b2-a85c-fb492b570fd7")
				}
			default:
			}

			// can be .json or .evidence.json
			basename := strings.TrimSuffix(strings.TrimSuffix(s, ".evidence.json"), ".json")
			fmt.Println(basename)
			reportPath := path.Join(reportDirname, basename+".report.json")
			err = os.WriteFile(reportPath, buf, 0664)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func BenchmarkAnalysis(b *testing.B) {
	ev := make(map[string]*evidence.Values)

	for _, s := range evidenceFiles {
		buf, err := ioutil.ReadFile(s)
		if err != nil {
			panic(err)
		}

		var props api.FirmwareProperties
		err = json.Unmarshal(buf, &props)
		if err != nil {
			panic(err)
		}

		ev[path.Base(s)], err = evidence.WrapInsecure(&api.Evidence{
			Firmware: props,
		})
		if err != nil {
			panic(err)
		}

	}

	b.ResetTimer()
	for s, ev := range ev {
		b.Run(s, func(b *testing.B) {
			for i := b.N; i > 0; i -= 1 {
				Compile(context.Background(), ev)
			}
		})
	}
}
