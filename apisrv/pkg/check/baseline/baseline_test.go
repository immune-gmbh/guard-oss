package baseline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaselineRowRoundtrip(t *testing.T) {
	baseline := New()
	assert.Equal(t, Hash{}, baseline.BootGuardIBB)
	assert.Nil(t, baseline.EmbeddedFirmware)
	assert.Nil(t, baseline.BootVariables)
	assert.Nil(t, baseline.DBXContents)

	row, err := ToRow(baseline)
	assert.NoError(t, err)
	baseline2, err := FromRow(row)
	assert.NoError(t, err)

	// nil buffers don't survive serialization
	assert.Equal(t, Hash{}, baseline2.BootGuardIBB)
	assert.Nil(t, baseline2.EmbeddedFirmware)
	assert.Nil(t, baseline2.BootVariables)
	assert.Nil(t, baseline2.DBXContents)
}
