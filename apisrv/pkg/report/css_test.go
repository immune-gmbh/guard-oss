package report

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/9elements/converged-security-suite/v2/pkg/test"
	"github.com/9elements/go-linux-lowlevel-hw/pkg/hwapi"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
)

func TestCSS(t *testing.T) {
	buf, err := ioutil.ReadFile("../../test/test.evidence.json")
	if err != nil {
		t.Fatal(err)
	}

	var ev api.Evidence
	err = json.Unmarshal(buf, &ev)
	if err != nil {
		t.Fatal(err)
	}

	val, err := evidence.WrapInsecure(&ev)
	if assert.NoError(t, err) {
		var cssApi hwapi.LowLevelHardwareInterfaces = NewCSSAPI(val)
		for _, t := range test.TestsCPU {
			fmt.Printf("Test: %s\n", t.Name)
			t.Run(cssApi, nil)
			fmt.Println(t.ErrorText)
		}
	}
}
