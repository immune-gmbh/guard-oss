package windows

import (
	"fmt"
	"testing"

	saferwall "github.com/saferwall/pe"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	sys, err := saferwall.New("../../test/eelam.sys", &saferwall.Options{})
	assert.NoError(t, err)
	err = sys.Parse()
	assert.NoError(t, err)

	ver, err := GetVersion(sys)
	assert.NoError(t, err)
	fmt.Println(ver)

	exe, err := saferwall.New("../../test/ekrn.exe", &saferwall.Options{})
	assert.NoError(t, err)
	err = exe.Parse()
	assert.NoError(t, err)

	ver, err = GetVersion(exe)
	assert.NoError(t, err)
	fmt.Println(ver)
}

func TestGetCertInfo(t *testing.T) {
	sys, err := saferwall.New("../../test/eelam.sys", &saferwall.Options{})
	assert.NoError(t, err)
	err = sys.Parse()
	assert.NoError(t, err)

	certs, err := GetELAMCertificateInfo(sys)
	assert.NoError(t, err)
	assert.Len(t, certs, 10)
	fmt.Println(certs)

	exe, err := saferwall.New("../../test/ekrn.exe", &saferwall.Options{})
	assert.NoError(t, err)
	err = exe.Parse()
	assert.NoError(t, err)

	_, err = GetELAMCertificateInfo(exe)
	assert.Error(t, err)
}
