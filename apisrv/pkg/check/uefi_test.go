package check

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

func TestDedupBootAppSet(t *testing.T) {
	buf, err := ioutil.ReadFile("../../test/h12ssl.values.json")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var ev evidence.Values
	err = json.Unmarshal(buf, &ev)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	buf, err = ioutil.ReadFile("../../test/h12ssl.baseline.json")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	var baseV2 baseline.ValuesV2
	err = json.Unmarshal(buf, &baseV2)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	base := baseline.MigrateV2ToV3(&baseV2)

	subj, err := NewSubject(context.Background(), &ev, base, policy.New())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	res, err := Run(context.Background(), subj)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	found := false
	for _, issue := range res.Issues {
		if issue.Id() == issuesv1.UefiBootAppSetId {
			found = true

			bootAppSet := make(map[string]bool)
			ubas := issue.(*issuesv1.UefiBootAppSet)
			for _, app := range ubas.Args.Apps {
				_, ok := bootAppSet[app.Path]
				assert.False(t, ok)
				bootAppSet[app.Path] = true
			}
		}
	}

	assert.True(t, found)
}

func TestUEFIBootAppChanged(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-bootorder3.json",
		uefiBootApp{},
		issuesv1.UefiBootAppSetId,
		func(subj *Subject) {})
}

func TestBootAppPinCert(t *testing.T) {
	subj := parseEvidence(t, "../../test/ludmilla-bootapps.evidence.json")

	// test pin certs for newly added boot apps
	c := uefiBootApp{}
	c.Update(context.Background(), nil, subj)

	expected := map[string][32]uint8{"PciRoot(0x0)/Pci(0x2,0x1)/Pci(0x0,0x0)/NVMe(0x1,00-1b-44-8b-49-fe-04-ce)/HD(1,GPT,2f0013e2-edc0-4fa3-bf14-ad67c400e883,0x800,0x100000)/\\EFI\\ubuntu\\shimx64.efi,": {0xdd, 0xb5, 0x0, 0x76, 0x9e, 0xf7, 0xe, 0xe1, 0x66, 0x94, 0xd7, 0x16, 0xbc, 0x79, 0x55, 0xbc, 0x25, 0x67, 0x5c, 0x9f, 0xd3, 0xe2, 0xb9, 0x8d, 0x89, 0xe2, 0x83, 0x2b, 0x92, 0x16, 0x65, 0xbd}, "\\EFI\\ubuntu\\grubx64.efi,": [32]uint8{0x99, 0x6b, 0xcc, 0xf3, 0xfc, 0xf8, 0xc2, 0x89, 0xdc, 0x53, 0x4d, 0x41, 0x2d, 0x3f, 0x3d, 0xe7, 0xc7, 0x9a, 0xf2, 0xf1, 0xa8, 0xfa, 0xb8, 0xc2, 0xfb, 0xd, 0x94, 0xab, 0x4a, 0x23, 0x79, 0x77}}
	assert.Len(t, subj.Baseline.BootApplications, len(expected))
	assert.True(t, subj.BaselineModified)
	for k, v := range expected {
		assert.Equal(t, v, *subj.Baseline.BootApplications[k].PinnedCertificateFingerprint)

		// remove the pinned hash for a boot app that has not changed and later check if it is pinned again
		ba := subj.Baseline.BootApplications[k]
		ba.PinnedCertificateFingerprint = nil
		subj.Baseline.BootApplications[k] = ba
	}

	// test pin certs for boot apps that have not changed
	subj.BaselineModified = false
	c.Update(context.Background(), nil, subj)
	assert.True(t, subj.BaselineModified)
	for k, v := range expected {
		assert.Equal(t, v, *subj.Baseline.BootApplications[k].PinnedCertificateFingerprint)
	}
}

func TestGPT(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiPartitionTable{},
		issuesv1.UefiGptChangedId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if pev, ok := (*ev).(eventlog.UEFIGPTEvent); ok {
				fmt.Println("data before", pev.Data[0])
				c := make([]byte, len(pev.Data))
				copy(c, pev.Data)
				pev.Data = c
				pev.Data[0] ^= 0xff
				fmt.Println("data after", pev.Data[0])
				alg := pev.Alg.CryptoHash().New()
				alg.Write(pev.Data)
				fmt.Printf("before %x\n", pev.Digest)
				pev.Digest = alg.Sum(nil)
				fmt.Printf("after %x\n", pev.Digest)
				*ev = pev
			}
		}))
}

func TestChangePK(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiSecureBootKeys{},
		issuesv1.UefiSecureBootKeysId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if pev, ok := (*ev).(eventlog.UEFIVariableDriverConfigEvent); ok && pev.VariableName == "PK" {
				c := make([]byte, len(pev.Data))
				copy(c, pev.Data)
				pev.Data = c
				pev.Data[0] ^= 0xff
				alg := pev.Alg.CryptoHash().New()
				alg.Write(pev.Data)
				pev.Digest = alg.Sum(nil)
				*ev = pev
			}
		}))
}

func TestChangeKEK(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiSecureBootKeys{},
		issuesv1.UefiSecureBootKeysId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if pev, ok := (*ev).(eventlog.UEFIVariableDriverConfigEvent); ok && pev.VariableName == "KEK" {
				c := make([]byte, len(pev.Data))
				copy(c, pev.Data)
				pev.Data = c
				pev.Data[0] ^= 0xff
				alg := pev.Alg.CryptoHash().New()
				alg.Write(pev.Data)
				pev.Digest = alg.Sum(nil)
				*ev = pev
			}
		}))
}

func TestSecureBootOff(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-no-secureboot.json",
		uefiSecureBootDisabled{},
		issuesv1.UefiSecureBootVariablesId,
		func(subj *Subject) {})
}

func TestSecureBootOn(t *testing.T) {
	testPure(t,
		"../../test/DESKTOP-MIMU51J-no-secureboot.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiSecureBootDisabled{},
		"",
		func(subj *Subject) {})
}

func TestReenableSecureBoot(t *testing.T) {
	testPure(t,
		"../../test/DESKTOP-MIMU51J-no-secureboot.json",
		"../../test/DESKTOP-MIMU51J-reenable-secureboot.json",
		uefiSecureBootDisabled{},
		"",
		func(subj *Subject) {})
}

func TestEntriesRemovedFromDbx(t *testing.T) {
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-cleardbx.json",
		uefiDbx{},
		issuesv1.UefiSecureBootDbxId,
		func(subj *Subject) {})
}

func TestNoExitBootServices(t *testing.T) {
	testPure(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiExitBootServices{},
		issuesv1.UefiNoExitBootSrvId,
		func(subj *Subject) {
			ctx := context.Background()
			boot := evidence.EmptyBoot()
			events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
			assert.NoError(t, err)
			events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
			assert.NoError(t, err)
			events := append(events1, events256...)
			filteredAfter := []eventlog.TPMEvent{}

			for _, ev := range events {
				if ac, ok := ev.(eventlog.UEFIActionEvent); !ok || ac.Message != "Exit Boot Services Invocation" {
					filteredAfter = append(filteredAfter, ev)
				}
			}
			for _, ev := range filteredAfter {
				boot.Consume(ctx, ev)
			}
			subj.Boot = *boot
		})
}

func TestNoExitBootServicesReturn(t *testing.T) {
	testPure(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiExitBootServices{},
		issuesv1.UefiNoExitBootSrvId,
		func(subj *Subject) {
			ctx := context.Background()
			boot := evidence.EmptyBoot()
			events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
			assert.NoError(t, err)
			events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
			assert.NoError(t, err)
			events := append(events1, events256...)
			filteredAfter := []eventlog.TPMEvent{}

			for _, ev := range events {
				if ac, ok := ev.(eventlog.UEFIActionEvent); !ok || ac.Message != "Exit Boot Services Returned with Success" {
					filteredAfter = append(filteredAfter, ev)
				}
			}
			for _, ev := range filteredAfter {
				boot.Consume(ctx, ev)
			}
			subj.Boot = *boot
		})
}

func TestNoSeparator(t *testing.T) {
	for pcr := 0; pcr < 7; pcr += 1 {
		testPure(t,
			"../../test/DESKTOP-MIMU51J-before.json",
			"../../test/DESKTOP-MIMU51J-before.json",
			uefiSeparators{},
			issuesv1.UefiBootFailureId,
			func(subj *Subject) {
				ctx := context.Background()
				boot := evidence.EmptyBoot()
				events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
				assert.NoError(t, err)
				events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
				assert.NoError(t, err)
				events := append(events1, events256...)
				filteredAfter := []eventlog.TPMEvent{}

				for _, ev := range events {
					if _, ok := ev.(eventlog.SeparatorEvent); !ok || ev.RawEvent().Index != pcr {
						filteredAfter = append(filteredAfter, ev)
					}
				}
				for _, ev := range filteredAfter {
					boot.Consume(ctx, ev)
				}
				subj.Boot = *boot
			})
	}
}

func TestFirmwareDigestChanged(t *testing.T) {
	// change oprom digest
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiEmbeddedFirmware{},
		issuesv1.UefiOptionRomSetId,
		modEventlog(t, func(ev *eventlog.TPMEvent) {
			if (*ev).RawEvent().Type == eventlog.EFIPlatformFirmwareBlob {
				pev := (*ev).(eventlog.UEFIPlatformFirmwareBlobEvent)
				if len(pev.Digest) > 0 {
					pev.Digest[0] ^= 0xff
					*ev = pev
				}
			}
		}))
}

func TestFirmwareAdded(t *testing.T) {
	// add oprom
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiEmbeddedFirmware{},
		"",
		func(subj *Subject) {
			ctx := context.Background()
			boot := evidence.EmptyBoot()
			events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
			assert.NoError(t, err)
			events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
			assert.NoError(t, err)
			events := append(events1, events256...)
			filteredAfter := []eventlog.TPMEvent{}

			hit := false
			for _, ev := range events {
				if !hit && ev.RawEvent().Type == eventlog.EFIPlatformFirmwareBlob {
					ev2 := ev.(eventlog.UEFIPlatformFirmwareBlobEvent)
					ev2.BlobBase = 0xfff0_0000
					filteredAfter = append(filteredAfter, ev2)
					hit = true
				}
				filteredAfter = append(filteredAfter, ev)
			}
			for _, ev := range filteredAfter {
				boot.Consume(ctx, ev)
			}
			subj.Boot = *boot
		})
}

func TestFirmwareDeleted(t *testing.T) {
	// remove oprom
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiEmbeddedFirmware{},
		"",
		func(subj *Subject) {
			ctx := context.Background()
			boot := evidence.EmptyBoot()
			events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
			assert.NoError(t, err)
			events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
			assert.NoError(t, err)
			events := append(events1, events256...)
			filteredAfter := []eventlog.TPMEvent{}

			hit := false
			for _, ev := range events {
				if !hit && ev.RawEvent().Type == eventlog.EFIPlatformFirmwareBlob {
					hit = true
				} else {
					filteredAfter = append(filteredAfter, ev)
				}
			}
			for _, ev := range filteredAfter {
				boot.Consume(ctx, ev)
			}
			subj.Boot = *boot
		})
}

func TestFirmwareReordered(t *testing.T) {
	// reorder oproms
	testCheck(t,
		"../../test/DESKTOP-MIMU51J-before.json",
		"../../test/DESKTOP-MIMU51J-before.json",
		uefiEmbeddedFirmware{},
		"",
		func(subj *Subject) {
			ctx := context.Background()
			boot := evidence.EmptyBoot()
			events256, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA256))
			assert.NoError(t, err)
			events1, err := eventlog.ParseEvents(subj.EventLogs[subj.BootEventLogIdx].Events(eventlog.HashSHA1))
			assert.NoError(t, err)
			events := append(events1, events256...)

			for i, ev := range events {
				if ev.RawEvent().Type == eventlog.EFIPlatformFirmwareBlob && i+1 < len(events) {
					ev2 := events[i+1]
					if ev2.RawEvent().Type == eventlog.EFIPlatformFirmwareBlob {
						events[i+1] = ev
						events[i] = ev2
					}
				}
			}
			for _, ev := range events {
				boot.Consume(ctx, ev)
			}
			subj.Boot = *boot
		})
}

func TestIssue1429(t *testing.T) {
	subj := parseEvidence(t, "../../test/agent.log.evidence.json")
	blinefile, err := ioutil.ReadFile("../../test/agent.log.baseline.json")
	assert.NoError(t, err)
	var bline baseline.Values
	err = json.Unmarshal(blinefile, &bline)
	assert.NoError(t, err)
	subj.Baseline = &bline

	ctx := context.Background()
	res, err := Run(ctx, subj)
	assert.NoError(t, err)

	for _, iss := range res.Issues {
		_, err := database.NewDocument(iss)
		assert.NoError(t, err)
	}
}
