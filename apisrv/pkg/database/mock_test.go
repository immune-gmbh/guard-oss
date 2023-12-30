package database

import (
	"context"
	"testing"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/stretchr/testify/assert"
)

func TestMockSimple(t *testing.T) {
	var rows []int

	mock := MockQuerier{
		Responses: map[string]MockResult{
			"select": {
				Columns: []string{"a"},
				Rows: [][]interface{}{
					{1},
					{1},
					{1},
					{1},
					{1},
					{1},
					{1},
					{1},
					{1},
					{1},
				},
			},
		},
	}
	ctx := context.Background()
	err := pgxscan.Select(ctx, mock, &rows, "select * from test")
	assert.NoError(t, err)
	assert.Len(t, rows, 10)
	for _, n := range rows {
		assert.Equal(t, 1, n)
	}
}

func TestMockStruct(t *testing.T) {
	var rows []struct {
		A int
		B string
	}

	mock := MockQuerier{
		Responses: map[string]MockResult{
			"select": {
				Columns: []string{"a", "b"},
				Rows: [][]interface{}{
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
					{1, "AAA"},
				},
			},
		},
	}
	ctx := context.Background()
	err := pgxscan.Select(ctx, mock, &rows, "select * from test")
	assert.NoError(t, err)
	assert.Len(t, rows, 10)
	for _, n := range rows {
		assert.Equal(t, 1, n.A)
		assert.Equal(t, "AAA", n.B)
	}
}
