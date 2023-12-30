package check

import (
	"bytes"
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func TestPlatformCodeChangedButNotSMBIOSVersion(t *testing.T) {
	// no change
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		intelBootGuard{},
		issuesv1.UefiIbbNoUpdateId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if (*ev).RawEvent().Type == eventlog.SCRTMContents {
				pev := (*ev).(eventlog.CRTMContentEvent)
				if len(pev.Digest) > 0 {
					pev.Digest[0] ^= 0xff
					*ev = pev
				}
			}
		}))

	// bios version change
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		intelBootGuard{},
		"",
		func(subj *Subject) {
			modEventlog(t, func(ev *eventlog.TPMEvent) {
				if (*ev).RawEvent().Type == eventlog.SCRTMContents {
					pev := (*ev).(eventlog.CRTMContentEvent)
					if len(pev.Digest) > 0 {
						pev.Digest[0] ^= 0xff
						*ev = pev
					}
				}
			})(subj)

			subj.Values.SMBIOS.Data = bytes.ReplaceAll(subj.Values.SMBIOS.Data, []byte("1.14.1"), []byte("2.14.1"))
		})

	// bios release date change
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		intelBootGuard{},
		"",
		func(subj *Subject) {
			modEventlog(t, func(ev *eventlog.TPMEvent) {
				if (*ev).RawEvent().Type == eventlog.SCRTMContents {
					pev := (*ev).(eventlog.CRTMContentEvent)
					if len(pev.Digest) > 0 {
						pev.Digest[0] ^= 0xff
						*ev = pev
					}
				}
			})

			subj.Values.SMBIOS.Data = bytes.ReplaceAll(subj.Values.SMBIOS.Data, []byte("12/18/2021"), []byte("12/18/2022"))
		})

	// XXX: SCRTM event
}
