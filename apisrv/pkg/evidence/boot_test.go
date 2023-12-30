package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/intelme"
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
)

func parseEvidence(t *testing.T, f string) *api.Evidence {
	buf, err := os.ReadFile(f)
	assert.NoError(t, err)

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	assert.NoError(t, err)

	if len(ev.AllPCRs) == 0 {
		ev.AllPCRs = map[string]map[string]api.Buffer{
			"11": ev.PCRs,
		}
	}

	return &ev
}

func BenchmarkConsume(b *testing.B) {
	evidence := make(map[string]*api.Evidence)

	for _, s := range evidenceFiles {
		buf, err := os.ReadFile(s)
		if err != nil {
			panic(err)
		}

		var ev api.Evidence
		err = json.Unmarshal(buf, &ev)
		if err != nil {
			panic(err)
		}

		evidence[path.Base(s)] = &ev
	}

	b.ResetTimer()
	for s, ev := range evidence {
		b.Run(s, func(b *testing.B) {
			for i := b.N; i > 0; i -= 1 {
				BootFromEvidence(context.Background(), ev)
			}
		})
	}
}

func TestCSMEEvents(t *testing.T) {
	evidence := parseEvidence(t, "../../test/issue906.1-good.evidence.json")
	boot, err := BootFromEvidence(context.Background(), evidence)
	assert.NoError(t, err)

	if boot.CSMEOperationMode != nil {
		fmt.Println(*boot.CSMEOperationMode)
	}
	keys := make(map[uint8]bool)
	for k := range boot.CSMEComponentVersions {
		keys[k] = true
	}
	for k := range boot.CSMEComponentHash {
		keys[k] = true
	}
	for k := range keys {
		hash, hashok := boot.CSMEComponentHash[k]
		ver, verok := boot.CSMEComponentVersions[k]
		fmt.Printf("component %s", intelme.MeasuredEntityToString(0, k))
		if verok {
			fmt.Printf(" has version %d.%d.%d.%d", ver.Version[0], ver.Version[1], ver.Version[2], ver.Version[3])
		} else {
			fmt.Printf(" has no version")
		}
		if hashok {
			fmt.Printf(" fw image %x\n", hash)
		} else {
			fmt.Printf(" no fw image checksum\n")
		}
	}
}

func TestIMAEvents(t *testing.T) {
	evidence := parseEvidence(t, "../../test/sr630.evidence.json")
	boot, err := BootFromEvidence(context.Background(), evidence)
	assert.NoError(t, err)

	fmt.Println(boot.Files["/etc/motd"])
}
