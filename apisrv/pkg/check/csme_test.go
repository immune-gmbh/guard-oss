package check

import (
	"bytes"
	"context"
	"path"
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/intelme"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	"github.com/stretchr/testify/assert"
)

func TestCSMECodeChangeButNotReportedVersion(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		csmeNoUpdate{},
		issuesv1.CsmeNoUpdateId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if (*ev).RawEvent().Type == eventlog.NonhostInfo {
				if pev, ok := (*ev).(eventlog.NonHostInfoEvent); ok {
					if fm, err := intelme.ParseMeasurmentEvent(pev); err == nil {
						for j := range fm.Events {
							ev := fm.Events[j]
							if cev, ok := ev.(intelme.ExtendManifestEvent); ok {
								idx := bytes.Index(pev.Data, cev.Data)
								pev.Data[idx] ^= 0xff
								cev.Data[0] ^= 0xff
								fm.Events[j] = cev
							}
						}
						alg := pev.Alg.CryptoHash().New()
						alg.Write(intelme.ReplayER(eventlog.HashSHA384, false, fm.Events))
						pev.Digest = alg.Sum(nil)
					}
					*ev = pev
				}
			}
		}))
}

func TestCSMECodeChangeAndVersionDowngrade(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		csmeDowngrade{},
		issuesv1.CsmeDowngradeId,
		func(subj *Subject) {
			modEventlog(t, func(ev *eventlog.TPMEvent) {
				if (*ev).RawEvent().Type == eventlog.NonhostInfo {
					if pev, ok := (*ev).(eventlog.NonHostInfoEvent); ok {
						if fm, err := intelme.ParseMeasurmentEvent(pev); err == nil {
							for j := range fm.Events {
								ev := fm.Events[j]
								if cev, ok := ev.(intelme.ExtendManifestEvent); ok {
									idx := bytes.Index(pev.Data, cev.Data)
									pev.Data[idx] ^= 0xff
									cev.Data[0] ^= 0xff
									fm.Events[j] = cev
								} else if cev, ok := ev.(intelme.ManifestVersionEvent); ok {
									idx := bytes.Index(pev.Data, cev.Data)
									pev.Data[idx] = 0
									cev.Data[0] = 0
									fm.Events[j] = cev
								}
							}
							alg := pev.Alg.CryptoHash().New()
							alg.Write(intelme.ReplayER(eventlog.HashSHA384, false, fm.Events))
							pev.Digest = alg.Sum(nil)
						}
						*ev = pev
					}
				}
			})(subj)

			getFwVerCmd := intelme.EncodeGetFirmwareVersion()
			for i := range subj.Values.ME {
				cmds := &subj.Values.ME[i]
				if cmds.GUID.String() == intelme.CSME_MKHIGuid.String() {
					for j := range cmds.Commands {
						c := &cmds.Commands[j]
						if bytes.Equal(c.Command, getFwVerCmd) {
							c.Response[6] -= 1
						}
					}
				}
			}
		})
}

func TestCSMEVersion(t *testing.T) {
	files := []string{
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/ludmilla.evidence.json",
		"../../test/Haruhi.json",
		"../../test/x1carbon.json",
		"../../test/sr630.evidence.json",
		"../../test/elitebook.json",
	}

	for _, f := range files {
		t.Run(path.Base(f), func(t *testing.T) {
			subj := parseEvidence(t, f)

			ann := csmeVulnerableVersion{}.Verify(context.Background(), subj)
			assert.Nil(t, ann)
		})
	}
}

func TestVulnCSMEVersion(t *testing.T) {
	subj := parseEvidence(t, "../../test/test-HP-EliteBook-830-G6-baseline.json")
	ann := csmeVulnerableVersion{}.Verify(context.Background(), subj)
	assert.NotNil(t, ann)
	assert.False(t, ann.Fatal)
}
