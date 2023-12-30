package evidence

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_wrapInsecure(t *testing.T) {
	ev := parseEvidence(t, "../../test/ludmilla.evidence.json")
	ev2 := parseEvidence(t, "../../test/ludmilla.evidence.json")

	// test if WrapInsecure modifies evidence and read in data twice above
	// to be absolutely sure that no inner pointers are pointing to the same data
	WrapInsecure(ev)
	assert.True(t, reflect.DeepEqual(ev, ev2), "WrapInsecure modifies evidence")
}
