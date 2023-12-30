package check

import (
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func TestGRUBGoodUpdate(t *testing.T) {
	testCheck(t,
		"../../test/test-before.json",
		"../../test/test-update.json",
		grub{},
		"",
		func(subj *Subject) {})
}

func TestGRUBKernelUpdate(t *testing.T) {
	testCheck(t,
		"../../test/test-before.json",
		"../../test/test-before.json",
		grub{},
		issuesv1.GrubBootChangedId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if pev, ok := (*ev).(eventlog.IPLEvent); ok && pev.Message == "/vmlinuz-5.4.0-96-generic\000" {
				pev.Digest[0] ^= 0xff
				*ev = pev
			}
		}))
}

func TestGRUBInitrdUpdate(t *testing.T) {
	testCheck(t,
		"../../test/test-before.json",
		"../../test/test-before.json",
		grub{},
		issuesv1.GrubBootChangedId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if pev, ok := (*ev).(eventlog.IPLEvent); ok && pev.Message == "/initrd.img-5.4.0-96-generic\000" {
				pev.Digest[0] ^= 0xff
				*ev = pev
			}
		}))
}
