package check

import (
	"context"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

const (
	FWUPD_RELEASE_FLAG_TRUSTED_PAYLOAD     uint64 = (1 << iota) // The payload binary is trusted
	FWUPD_RELEASE_FLAG_TRUSTED_METADATA                         // The payload metadata is trusted
	FWUPD_RELEASE_FLAG_IS_UPGRADE                               // The release is newer than the device version.
	FWUPD_RELEASE_FLAG_IS_DOWNGRADE                             // The release is older than the device version
	FWUPD_RELEASE_FLAG_BLOCKED_VERSION                          // The installation of the release is blocked as below device version-lowest
	FWUPD_RELEASE_FLAG_BLOCKED_APPROVAL                         // The installation of the release is blocked as release not approved by an administrator
	FWUPD_RELEASE_FLAG_IS_ALTERNATE_BRANCH                      // The release is an alternate branch of firmware
	FWUPD_RELEASE_FLAG_IS_COMMUNITY                             //The release is supported by the community and not the hardware vendor
)

type lvfsFirmwareUpdateCheck struct{}

func (lvfsFirmwareUpdateCheck) String() string {
	return "LVFS Update Check"
}

func (lvfsFirmwareUpdateCheck) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if subj.Baseline.AllowOutdatedFirmware {
		return nil
	}
	if len(subj.FwupdDevices) == 0 {
		return nil
	}

	var iss issuesv1.FirmwareUpdate
	iss.Common.Id = issuesv1.FirmwareUpdateId
	iss.Common.Aspect = issuesv1.FirmwareUpdateAspect
	iss.Common.Incident = false

	for _, device := range subj.FwupdDevices {
		if len(device.Releases) == 0 {
			continue
		}

		// the releases array is sorted by version
		upd := device.Releases[0]

		// no upgrades found
		if (upd.Flags & FWUPD_RELEASE_FLAG_IS_UPGRADE) != FWUPD_RELEASE_FLAG_IS_UPGRADE {
			continue
		}

		up := issuesv1.FirmwareUpdateUpdates{
			Name:    device.Name,
			Current: device.Version,
			Next:    upd.Version,
		}
		iss.Args.Updates = append(iss.Args.Updates, up)
	}

	if len(iss.Args.Updates) > 0 {
		return &iss
	} else {
		return nil
	}
}

func (lvfsFirmwareUpdateCheck) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.FirmwareUpdateId)

	if allow && !subj.Baseline.AllowOutdatedFirmware {
		subj.Baseline.AllowOutdatedFirmware = true
		subj.BaselineModified = true
	}
}
