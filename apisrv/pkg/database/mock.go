package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MockQuerier struct {
	Responses map[string]MockResult
}

type MockResult struct {
	Columns []string
	Rows    [][]interface{}

	iterator int
}

func (q MockQuerier) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	fmt.Println(query)
	for k, v := range q.Responses {
		m, err := regexp.MatchString(k, query)
		if err != nil {
			fmt.Println(err)
		} else if m {
			fmt.Printf("Match %s\n", k)
			return &v, nil
		}
	}
	fmt.Println("-- No match --")
	return nil, ErrNotFound
}

func (rows *MockResult) FieldDescriptions() []pgproto3.FieldDescription {
	desc := make([]pgproto3.FieldDescription, len(rows.Columns))
	for i, n := range rows.Columns {
		desc[i].Name = []byte(n)
	}
	return desc
}

func (rows *MockResult) Close() {
}

func (rows *MockResult) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag("NONE")
}

func (rows *MockResult) Err() error {
	return nil
}

func (rows *MockResult) Next() bool {
	rows.iterator += 1
	return rows.iterator <= len(rows.Rows)
}

func (rows *MockResult) Scan(dest ...interface{}) error {
	row := rows.Rows[rows.iterator-1]
	if len(dest) != len(row) {
		return errors.New("wrong column num")
	}

	for i, dst := range dest {
		if dst == nil {
			continue
		}
		fmt.Sscan(fmt.Sprint(row[i]), dest[i])
	}

	return nil
}

func (rows *MockResult) Values() ([]interface{}, error) {
	return rows.Rows[rows.iterator-1], nil
}

func (rows *MockResult) RawValues() [][]byte {
	return nil
}

type ExplainQuerier struct {
	MaxCost  float64
	Database *pgxpool.Pool
}

var explainAnalyze = regexp.MustCompile(`^(\w|\s)+\(cost=\d+\.\d+\.\.(\d+\.\d+)`) //\(cost=\d+\.\d+..(\d+\.\d+).+$`)

func (q ExplainQuerier) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	fmt.Println(query)
	tx, err := q.Database.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, "set enable_seqscan = off")
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, fmt.Sprint("explain analyze ", query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	maxCost := 0.0
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		fmt.Println(s)

		// parse explain analyze
		m := explainAnalyze.FindStringSubmatch(s)
		if len(m) >= 3 {
			cost, err := strconv.ParseFloat(m[2], 64)
			if err != nil {
				return nil, err
			}
			maxCost = math.Max(cost, maxCost)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	tx.Rollback(ctx)

	if q.MaxCost > 0 && maxCost > q.MaxCost {
		fmt.Println("query cost", maxCost, "limit is", q.MaxCost)
		return nil, errors.New("query too expensive")
	}
	return q.Database.Query(ctx, query, args...)
}
