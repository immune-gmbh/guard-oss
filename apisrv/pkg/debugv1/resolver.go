package debugv1

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	graphql1 "github.com/immune-gmbh/guard/apisrv/v2/internal/graphql"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/binarly"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/queue"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const (
	defaultEvidenceListLimit = 10
	defaultDeviceListLimit   = 10
)

var (
	errSingleAndFilter = errors.New("cannot filter a id or reference query")
	errIdXorRef        = errors.New("can only query by id or reference and type")
	errMultipleFilters = errors.New("cannot query with multiple filters")
	errDeviceIdOrRef   = errors.New("can only query by device id and organization id or reference")
)

func onlyOne(preds ...bool) bool {
	var ret bool
	for _, b := range preds {
		if b && ret {
			return false
		}
		ret = b || ret
	}
	return true
}

func convTime(ts *time.Time) *string {
	if ts != nil {
		str := ts.Format(time.RFC3339)
		return &str
	} else {
		return nil
	}
}

func convertJob(row *queue.Row) (*graphql1.Job, error) {
	var args *string
	if !row.Args.IsNull() {
		raw, err := row.Args.Bytes()
		if err != nil {
			return nil, err
		}
		str := string(raw)
		args = &str
	}

	job := graphql1.Job{
		ID:          fmt.Sprint(row.Id),
		Reference:   row.Reference,
		Type:        row.Type,
		Args:        args,
		ScheduledAt: row.ScheduledAt.Format(time.RFC3339),
		NextRunAt:   row.NextRunAt.Format(time.RFC3339),
		LastRunAt:   convTime(row.LastRunAt),
		LockedAt:    convTime(row.LockedAt),
		LockedUntil: convTime(row.LockedUntil),
		LockedBy:    row.LockedBy,
		ErrorCount:  row.ErrorCount,
		Successful:  row.Successful,
		FinishedAt:  convTime(row.FinishedAt),
	}
	return &job, nil
}

func convertEvidence(row *evidence.Row, report *binarly.Report) (*graphql1.Evidence, error) {
	vals, err := row.Values.Bytes()
	if err != nil {
		return nil, err
	}

	var repstr *string
	if report != nil {
		repdoc, err := database.NewDocument(*report)
		if err != nil {
			return nil, err
		}
		b, err := repdoc.Bytes()
		if err != nil {
			return nil, err
		}
		s := string(b)
		repstr = &s
	}

	ev := graphql1.Evidence{
		ID:         fmt.Sprint(row.Id),
		ReceivedAt: row.ReceivedAt.Format(time.RFC3339),
		RawValues:  string(vals),
		RawBinarly: repstr,
	}
	return &ev, nil
}

type Resolver struct {
	pool *pgxpool.Pool
}
