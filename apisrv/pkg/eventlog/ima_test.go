package eventlog

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSR630(t *testing.T) {
	buf, err := ioutil.ReadFile("../../test/sr630.binary_measurements")
	assert.NoError(t, err)

	ctx := context.Background()
	log, err := ParseIMA(ctx, buf)
	assert.NoError(t, err)

	for _, ev := range log {
		fmt.Println(ev)
	}
}

func TestParseH12SSL(t *testing.T) {
	buf, err := ioutil.ReadFile("../../test/h12ssl.binary_measurements")
	assert.NoError(t, err)

	ctx := context.Background()
	log, err := ParseIMA(ctx, buf)
	assert.NoError(t, err)

	for _, ev := range log {
		fmt.Println(ev)
	}
}
