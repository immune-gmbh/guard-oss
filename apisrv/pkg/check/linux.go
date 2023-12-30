package check

import (
	"context"
	"reflect"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

type grub struct{}

func (grub) String() string {
	return "GRUB2"
}

func (grub) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	kernelChanged := !subj.Baseline.LinuxDigest.IntersectsWith(&subj.Boot.LinuxDigest)
	initrdChanged := !subj.Baseline.InitrdDigest.IntersectsWith(&subj.Boot.InitrdDigest)
	cmdlineChanged := len(subj.Baseline.LinuxCommandLine) > 0 && !reflect.DeepEqual(subj.Baseline.LinuxCommandLine, subj.Boot.LinuxCommand)

	var iss issuesv1.GrubBootChanged
	iss.Common.Id = issuesv1.GrubBootChangedId
	iss.Common.Aspect = issuesv1.GrubBootChangedAspect
	iss.Common.Incident = true
	// kernel digest
	iss.Args.Before.Kernel, iss.Args.After.Kernel =
		baseline.BeforeAfter(&subj.Baseline.LinuxDigest, &subj.Boot.LinuxDigest)
		// kernel path
	iss.Args.Before.KernelPath = subj.Baseline.LinuxPath
	iss.Args.After.KernelPath = subj.Boot.LinuxFile
	// initrd digest
	iss.Args.Before.Initrd, iss.Args.After.Initrd =
		baseline.BeforeAfter(&subj.Baseline.InitrdDigest, &subj.Boot.InitrdDigest)
		// initrd path
	iss.Args.Before.InitrdPath = subj.Baseline.InitrdPath
	iss.Args.After.InitrdPath = subj.Boot.InitrdFile
	// kernel command line
	iss.Args.Before.CommandLine = subj.Baseline.LinuxCommandLine
	iss.Args.After.CommandLine = subj.Boot.LinuxCommand

	if kernelChanged && !cmdlineChanged && subj.Baseline.LinuxPath != "" && subj.Baseline.LinuxPath == subj.Boot.LinuxFile {
		return &iss
	}
	if initrdChanged && subj.Baseline.InitrdPath != "" && subj.Baseline.InitrdPath == subj.Boot.InitrdFile {
		return &iss
	}

	return nil
}

func (grub) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.GrubBootChangedId)
	change := false

	if allowChange {
		change = change || !reflect.DeepEqual(subj.Baseline.LinuxDigest, subj.Boot.LinuxDigest)
		subj.Baseline.LinuxDigest = subj.Boot.LinuxDigest
	} else {
		change = subj.Baseline.LinuxDigest.UnionWith(&subj.Boot.LinuxDigest) || change
	}
	if allowChange {
		change = change || !reflect.DeepEqual(subj.Baseline.InitrdDigest, subj.Boot.InitrdDigest)
		subj.Baseline.InitrdDigest = subj.Boot.InitrdDigest
	} else {
		change = subj.Baseline.InitrdDigest.UnionWith(&subj.Boot.InitrdDigest) || change
	}
	if len(subj.Baseline.LinuxCommandLine) == 0 || allowChange {
		change = change || !reflect.DeepEqual(subj.Baseline.LinuxCommandLine, subj.Boot.LinuxCommand)
		subj.Baseline.LinuxCommandLine = subj.Boot.LinuxCommand
	}
	if subj.Baseline.LinuxPath == "" || (allowChange && subj.Boot.LinuxFile != "") {
		change = change || subj.Baseline.LinuxPath != subj.Boot.LinuxFile
		subj.Baseline.LinuxPath = subj.Boot.LinuxFile
	}
	if subj.Baseline.InitrdPath == "" || (allowChange && subj.Boot.InitrdFile != "") {
		change = change || subj.Baseline.InitrdPath != subj.Boot.InitrdFile
		subj.Baseline.InitrdPath = subj.Boot.InitrdFile
	}

	subj.BaselineModified = subj.BaselineModified || change
}
