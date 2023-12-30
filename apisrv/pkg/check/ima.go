package check

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"reflect"

	"github.com/google/go-tpm/tpm2"
	log "github.com/sirupsen/logrus"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/baseline"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/eventlog"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type imaLog struct{}

func (imaLog) String() string {
	return "IMA File Measurements"
}

func (imaLog) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if len(subj.IMALog) == 0 {
		return nil
	}
	if subjectHasDummyTPM(subj) {
		return nil
	}

	// verify log
	var hit bool
	var iss issuesv1.ImaInvalidLog
	iss.Common.Id = issuesv1.ImaInvalidLogId
	iss.Common.Aspect = issuesv1.ImaInvalidLogAspect
	iss.Common.Incident = true

	for algo, bank := range subj.Values.PCR {
		var h hash.Hash
		switch algo {
		case "4":
			h = sha1.New()
		case "11":
			h = sha256.New()
		default:
			tel.Log(ctx).WithField("algo bank", algo).Error("unknown hash algo")
			continue
		}

		_, err := eventlog.VerifyIMA(subj.IMALog, bank, h)
		if err != nil {
			var replayErr eventlog.ImaReplayErr
			iss.Args.Pcr = nil
			if errors.As(err, &replayErr) {
				for p, v := range replayErr.InvalidPCRs {
					iss.Args.Pcr = append(iss.Args.Pcr, issuesv1.ImaInvalidLogPcr{
						Quoted:   bank[p],
						Computed: v,
						Number:   p,
					})
				}
				tel.Log(ctx).WithField("bank", algo).WithError(err).Info("ima log replay error")
			} else {
				tel.Log(ctx).WithField("bank", algo).WithError(err).Error("verify ima log")
			}
			continue
		}
		hit = true
	}
	if !hit && !subj.Baseline.AllowInvalidImaLog {
		return &iss
	}

	return nil
}

func (imaLog) Update(ctx context.Context, overrides []string, subj *Subject) {
	allow := hasIssue(overrides, issuesv1.ImaInvalidLogId)
	if allow && !subj.Baseline.AllowInvalidImaLog {
		subj.Baseline.AllowInvalidImaLog = true
		subj.BaselineModified = true
	}
}

type imaBootAggregate struct{}

func (imaBootAggregate) String() string {
	return "IMA Boot Aggregate"
}

func bootAggregate(subj *Subject) (*baseline.Hash, error) {
	var agg baseline.Hash

	h1bank := fmt.Sprint(int(tpm2.AlgSHA1))
	if bank, ok := subj.Values.PCR[h1bank]; ok {
		h1 := sha1.New()
		for index := 0; index < 8; index += 1 {
			if s, ok := bank[fmt.Sprint(index)]; ok {
				if b, err := hex.DecodeString(s); err == nil {
					h1.Write(b)
				}
			}
		}

		hh, err := baseline.NewHash(h1.Sum(nil))
		if err != nil {
			return nil, err
		}
		agg.UnionWith(&hh)
	}

	return &agg, nil
}

func (imaBootAggregate) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	agg, err := bootAggregate(subj)
	if err != nil {
		tel.Log(ctx).WithError(err).Error("create hash")
		return nil
	}
	if !subj.Boot.BootAggregate.IntersectsWith(agg) {
		tel.Log(ctx).WithFields(log.Fields{"log": subj.Boot.BootAggregate, "computed": *agg}).Info("boot aggregate")
		var iss issuesv1.ImaBootAggregate
		computed, logged := baseline.BeforeAfter(agg, &subj.Boot.BootAggregate)
		iss.Common.Id = issuesv1.ImaBootAggregateId
		iss.Common.Aspect = issuesv1.ImaBootAggregateAspect
		iss.Common.Incident = true
		iss.Args.Computed = computed
		iss.Args.Logged = logged
		return &iss
	}
	return nil
}

func (imaBootAggregate) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.ImaBootAggregateId)
	change := false

	if allowChange {
		change = change || !reflect.DeepEqual(subj.Baseline.BootAggregate, subj.Boot.BootAggregate)
		subj.Baseline.BootAggregate = subj.Boot.BootAggregate
	} else {
		change = subj.Baseline.BootAggregate.UnionWith(&subj.Boot.BootAggregate) || change
	}

	subj.BaselineModified = subj.BaselineModified || change
}

type imaFiles struct{}

func (imaFiles) String() string {
	return "IMA File Measurements"
}

func (imaFiles) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	var iss issuesv1.ImaRuntimeMeasurements
	iss.Common.Id = issuesv1.ImaRuntimeMeasurementsId
	iss.Common.Aspect = issuesv1.ImaRuntimeMeasurementsAspect
	iss.Common.Incident = true

	for _, file := range subj.Policy.ProtectedFiles {
		hash, ok := subj.Boot.Files[file.Path]
		if !ok {
			continue
		}
		if len(subj.Baseline.FileMeasurements) > 0 {
			prevhash, ok := subj.Baseline.FileMeasurements[file.Path]
			if ok && !hash.IntersectsWith(&prevhash) {
				tel.Log(ctx).WithFields(log.Fields{"path": file.Path, "previous": prevhash.String(), "now": hash.String()}).Info("file changed")

				before, after := baseline.BeforeAfter(&prevhash, &hash)
				iss.Args.Files = append(iss.Args.Files, issuesv1.ImaRuntimeMeasurementsFile{
					Path:   file.Path,
					Before: before,
					After:  after,
				})
			}
		}
	}

	if len(iss.Args.Files) > 0 {
		return &iss
	} else {
		return nil
	}
}

func (imaFiles) Update(ctx context.Context, overrides []string, subj *Subject) {
	allowChange := hasIssue(overrides, issuesv1.ImaRuntimeMeasurementsId)
	change := false

	for _, file := range subj.Policy.ProtectedFiles {
		hash, ok := subj.Boot.Files[file.Path]
		if subj.Baseline.FileMeasurements == nil {
			subj.Baseline.FileMeasurements = make(map[string]baseline.Hash)
		}
		prevhash, prevok := subj.Baseline.FileMeasurements[file.Path]

		if allowChange {
			change = true
			if ok {
				subj.Baseline.FileMeasurements[file.Path] = hash
			} else {
				delete(subj.Baseline.FileMeasurements, file.Path)
			}
		} else if ok {
			if prevok {
				change = prevhash.UnionWith(&hash) || change
				subj.Baseline.FileMeasurements[file.Path] = prevhash
			} else {
				change = true
				subj.Baseline.FileMeasurements[file.Path] = hash
			}
		}
	}
	subj.BaselineModified = subj.BaselineModified || change
}
