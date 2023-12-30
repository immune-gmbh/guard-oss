package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	tel "github.com/immune-gmbh/guard/apisrv/v2/pkg/telemetry"
)

type Row struct {
	Id           string            `db:"id"`
	Reference    string            `db:"reference"`
	Type         string            `db:"type"`
	Args         database.Document `db:"args"`
	ScheduledAt  time.Time         `db:"scheduled_at"`
	ScheduledCtx database.Document `db:"scheduled_ctx"`
	NextRunAt    time.Time         `db:"next_run_at"`

	LastRunAt   *time.Time        `db:"last_run_at"`
	LastRunCtx  database.Document `db:"last_run_ctx"`
	LockedAt    *time.Time        `db:"locked_at"`
	LockedUntil *time.Time        `db:"locked_until"`
	LockedBy    *string           `db:"locked_by"`
	ErrorCount  int               `db:"error_count"`

	Successful *bool      `db:"successful"`
	FinishedAt *time.Time `db:"finished_at"`
}

var (
	InvalidFilterErr = errors.New("invalid List filter")

	rowFields = map[string]bool{
		"id":            true,
		"reference":     true,
		"type":          true,
		"args":          true,
		"scheduled_at":  true,
		"scheduled_ctx": true,
		"next_run_at":   true,
		"last_run_at":   true,
		"last_run_ctx":  true,
		"locked_at":     true,
		"locked_until":  true,
		"locked_by":     true,
		"error_count":   true,
		"successful":    true,
		"finished_at":   true,
	}
)

func Enqueue(ctx context.Context, q pgxscan.Querier, ty string, ref string, args interface{}, runAt time.Time, now time.Time) (*Row, error) {
	var row Row
	var spanctx database.Document

	argsdoc, err := database.NewDocument(args)
	if err != nil {
		return nil, err
	}

	if sctx, err := tel.SpanContext(ctx); err == nil {
		spanctx, err = database.NewDocumentRaw([]byte(sctx))
		if err != nil {
			return nil, err
		}
	}

	err = pgxscan.Get(ctx, q, &row, `
		insert into v2.jobs (
      reference,
			type,
			args,
			scheduled_at,
			scheduled_ctx,
			next_run_at
		) values (
			$1,
			$2,
			$3,
			$4::timestamptz,
			$5,
			$6::timestamptz
		) returning *
	`,
		ref,     // $1
		ty,      // $2
		argsdoc, // $3
		now,     // $4
		spanctx, // $5
		runAt)   // $6

	if err != nil {
		return nil, err
	}

	return &row, nil
}

func AreReady(ctx context.Context, q pgxscan.Querier, refs []string, maxAge time.Time) (bool, error) {
	var flag *bool
	err := pgxscan.Get(ctx, q, &flag, `
    select
      bool_and(successful is not null or $2 > scheduled_at)
    from v2.jobs
    where reference = any($1::text[])
  `, refs, maxAge)
	if err != nil {
		return false, database.Error(err)
	}

	if flag == nil {
		return len(refs) == 0, nil
	} else {
		return *flag, nil
	}
}

func ByReference(ctx context.Context, q pgxscan.Querier, ty string, ref string) (*Row, error) {
	var row Row

	err := pgxscan.Get(ctx, q, &row, `
		select * from v2.jobs
		where type = $1 
		  and reference = $2
	`, ty, ref)
	if err != nil {
		return nil, database.Error(err)
	}

	return &row, nil
}

func ById(ctx context.Context, q pgxscan.Querier, id string) (*Row, error) {
	var row Row

	err := pgxscan.Get(ctx, q, &row, "select * from v2.jobs where id = $1", id)
	if err != nil {
		return nil, database.Error(err)
	}

	return &row, nil
}

type FilterType struct {
	Types []string
}

type FilterReference struct {
	Set      []string
	LessThan string
}

type FilterTimestamp struct {
	ScheduledAtLessThen string
	LockedAtLessThen    string
	NextRunAtLessThen   string
	FinishedAtLessThen  string
}

type FilterId struct {
	Set      []string
	LessThan string
}

type Status int

// List(..., limit, ...)
const (
	Running = 0
	Failed  = 1
	Queued  = 2
	Done    = 3
)

// List(..., orderBy, ...)
const (
	Id          = 0
	ScheduledAt = 1
	LockedAt    = 2
	FinishedAt  = 3
	NextRunAt   = 4
	Reference   = 5
)

type FilterStatus struct {
	Status Status
}

func List(ctx context.Context, q pgxscan.Querier, limit int, orderBy int, fields []string, filters ...interface{}) ([]*Row, error) {
	clauses := []string{}
	args := []interface{}{limit}
	for _, f := range filters {
		switch f := f.(type) {
		case FilterId:
			if f.LessThan != "" && len(f.Set) > 0 {
				return nil, InvalidFilterErr
			} else if f.LessThan != "" {
				clauses = append(clauses, fmt.Sprintf("v2.jobs.id < $%d", len(args)+1))
				args = append(args, f.LessThan)
			} else if len(f.Set) > 0 {
				clauses = append(clauses, fmt.Sprintf("v2.jobs.id = any($%d::text[])", len(args)+1))
				args = append(args, f.Set)
			}
		case FilterReference:
			if f.LessThan != "" && len(f.Set) > 0 {
				return nil, InvalidFilterErr
			} else if f.LessThan != "" {
				clauses = append(clauses, fmt.Sprintf("v2.jobs.reference < $%d::text", len(args)+1))
				args = append(args, f.LessThan)
			} else if len(f.Set) > 0 {
				clauses = append(clauses, fmt.Sprintf("v2.jobs.reference = any($%d::text[])", len(args)+1))
				args = append(args, f.Set)
			}
		case FilterType:
			clauses = append(clauses, fmt.Sprintf("v2.jobs.type = any($%d::text[])", len(args)+1))
			args = append(args, f.Types)
		case FilterStatus:
			switch f.Status {
			case Running:
				clauses = append(clauses, "v2.jobs.locked_by is not null")
			case Queued:
				clauses = append(clauses, "v2.jobs.finished_at is null and v2.jobs.locked_by is null")
			case Failed:
				clauses = append(clauses, "v2.jobs.successful = false")
			case Done:
				clauses = append(clauses, "v2.jobs.successful = true")
			}
		}
	}
	columns := "*"
	var filtedFields []string
	for _, f := range fields {
		if _, ok := rowFields[f]; ok {
			filtedFields = append(filtedFields, f)
		}
	}
	if len(filtedFields) > 0 {
		columns = strings.Join(filtedFields, ", ")
	}
	where := ""
	if len(clauses) > 0 {
		where = fmt.Sprint("where ", strings.Join(clauses, " and "))
	}
	var order string
	switch orderBy {
	case Id:
		fallthrough
	default:
		order = "v2.jobs.id"
	case ScheduledAt:
		order = "v2.jobs.scheduled_at"
	case LockedAt:
		order = "v2.jobs.locked_at"
	case NextRunAt:
		order = "v2.jobs.next_run_at"
	case FinishedAt:
		order = "v2.jobs.finished_at"
	case Reference:
		order = "v2.jobs.reference"
	}

	sql := fmt.Sprintf("select %s from v2.jobs %s order by %s desc limit $1", columns, where, order)
	rows := []*Row{}
	err := pgxscan.Select(ctx, q, &rows, sql, args...)
	if err != nil {
		return nil, database.Error(err)
	}

	return rows, nil
}

func lockJob(ctx context.Context, q pgxscan.Querier, instanceName string, types []string, timeout time.Time, now time.Time) (*Row, error) {
	var row Row
	var spanctx database.Document

	if sctx, err := tel.SpanContext(ctx); err == nil {
		spanctx, err = database.NewDocumentRaw([]byte(sctx))
		if err != nil {
			return nil, err
		}
	}

	// change v2.jobs_waiting index if this is modified
	err := pgxscan.Get(ctx, q, &row, `
		with job as (
			select id as selected from v2.jobs
			where next_run_at <= $1
				and locked_by is null
				and successful is null
        and type = any($5)
			order by random()
			for update skip locked
			limit 1
		)
		update v2.jobs 
		set (
      locked_at,
      locked_until,
      locked_by,
      last_run_ctx,
      last_run_at
    ) = (
      $1::timestamptz,
      $2::timestamptz,
      $3,
      $4::jsonb,
      $1::timestamptz
    )
    from job
		where v2.jobs.id = job.selected
		returning v2.jobs.*
	`,
		now,          // $1
		timeout,      // $2
		instanceName, // $3
		spanctx,      // $4
		types)        // $5
	if err != nil {
		return nil, database.Error(err)
	}

	return &row, nil
}

func unlockJob(ctx context.Context, q pgxscan.Querier, job *Row) error {
	var affected []string
	err := pgxscan.Select(ctx, q, &affected, `
		update v2.jobs 
		set (
			locked_at,
			locked_until,
			locked_by, 
			next_run_at,
			error_count,
			finished_at,
			successful
		) = (
			null,
			null,
			null, 
			$1::timestamptz,
			$2,
      $3,
      $4
		)
		where id = $5 and (locked_by = $6 or locked_by is null)
    returning id
	`,
		job.NextRunAt,  // $1
		job.ErrorCount, // $2
		job.FinishedAt, // $3
		job.Successful, // $4
		job.Id,         // $5
		job.LockedBy)   // $6
	if err != nil {
		return database.Error(err)
	}
	if len(affected) == 0 {
		return database.ErrNotFound
	}
	return nil
}

func unlockExpired(ctx context.Context, q pgxscan.Querier, now time.Time) (int, error) {
	// change v2.jobs_locked index if this is modified
	var affected []string
	err := pgxscan.Select(ctx, q, &affected, `
		update v2.jobs 
		set (
			locked_at,
			locked_until,
			locked_by, 
			error_count
		) = (
			null,
			null,
			null, 
			error_count + 1
		)
		where (locked_at is not null or locked_by is not null)
			and locked_until <= $1::timestamptz
    returning id
	`, now)
	if err != nil {
		return 0, database.Error(err)
	}
	return len(affected), nil
}
