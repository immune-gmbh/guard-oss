package database

import (
	"context"
	"embed"
	"os"
	"testing"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/mock"
)

func TestNullMatching(t *testing.T) {
	examples := map[string]string{
		`"Hello, World"`:           `"Hello, World"`,
		`"Hello, \u0000 World"`:    `"Hello, %00 World"`,
		`"Hello, \\u0000 World"`:   `"Hello, \\u0000 World"`,
		`"Hello, \\\u0000 World"`:  `"Hello, \\%00 World"`,
		`"Hello, \\\\u0000 World"`: `"Hello, \\\\u0000 World"`,
		`"Hello, \n World"`:        `"Hello, \n World"`,
		`"Hello, \\n World"`:       `"Hello, \\n World"`,
		`"Hello, \\`:               `"Hello, \\`,
		`"Hello, \`:                `"Hello, \`,
		`"Hello, % World`:          `"Hello, %25 World`,
		`"Hello, %% World`:         `"Hello, %25%25 World`,
		`"Hello, %25\u0000 World`:  `"Hello, %2525%00 World`,
		`渀\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000`: `渀%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00`,
	}

	for k, v := range examples {
		t.Run(k, func(t *testing.T) {
			v2, err := encode([]byte(k))
			assert.NoError(t, err)
			assert.Equal(t, v, string(v2))
			k2, err := decode(v2)
			assert.NoError(t, err)
			assert.Equal(t, k, string(k2))
		})
	}
}

func TestDocumentType(t *testing.T) {
	examples := map[string]string{
		`{"type":"test/1"}`: `test/1`,
		`{"typ":"test/1"}`:  ``,
		`{type:"test/1"}`:   ``,
	}

	for k, v := range examples {
		t.Run(k, func(t *testing.T) {
			doc, err := NewDocumentRaw([]byte(k))
			assert.NoError(t, err)
			assert.Equal(t, v, doc.Type())
		})
	}
}

type row struct {
	Id  int64    `db:"id"`
	Doc Document `db:"doc"`
}

func TestDocumentSQL(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		"Select":     testSelectDoc,
		"Insert":     testInsertDoc,
		"InsertNull": testInsertNullDoc,
		"Sanitize":   testSanitize,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.Reset(t, ctx, embed.FS{})

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()

			_, err := conn.Exec(ctx, `create table test (id bigserial primary key, doc jsonb);`)
			assert.NoError(t, err)

			fn(t, conn)
		})
	}
}

func testSelectDoc(t *testing.T, conn *pgxpool.Pool) {
	ctx := context.Background()
	doc, err := NewDocumentRaw([]byte(`{"type":"test/1","a":1}`))
	assert.NoError(t, err)

	var row row
	err = pgxscan.Get(ctx, conn, &row, `select 42 as id, $1::jsonb as doc`, doc)
	assert.NoError(t, err)
	assert.False(t, row.Doc.IsNull())
	assert.Equal(t, "test/1", row.Doc.Type())
}

func testSanitize(t *testing.T, conn *pgxpool.Pool) {
	ctx := context.Background()
	doc, err := NewDocumentRaw([]byte(`{"type":"test/1","int":42,"string":"Hello, \\u0000 World%00"}`))
	assert.NoError(t, err)

	var row row
	err = pgxscan.Get(ctx, conn, &row, `insert into test (doc) values ($1::jsonb) returning *`, doc)
	assert.NoError(t, err)
	assert.False(t, row.Doc.IsNull())
	assert.Equal(t, "test/1", row.Doc.Type())

	var ddoc1, ddoc2 testdoc
	err = doc.Decode(&ddoc1)
	assert.NoError(t, err)
	err = row.Doc.Decode(&ddoc2)
	assert.NoError(t, err)
	assert.Equal(t, ddoc2, ddoc1)
}

func testInsertDoc(t *testing.T, conn *pgxpool.Pool) {
	ctx := context.Background()
	doc, err := NewDocumentRaw([]byte(`{"type":"test/1","a":1}`))
	assert.NoError(t, err)

	var row row
	err = pgxscan.Get(ctx, conn, &row, `insert into test (doc) values ($1::jsonb) returning *`, doc)
	assert.NoError(t, err)
	assert.False(t, row.Doc.IsNull())
	assert.Equal(t, "test/1", row.Doc.Type())
}

func testInsertNullDoc(t *testing.T, conn *pgxpool.Pool) {
	ctx := context.Background()
	doc, err := NewDocument(nil)
	assert.NoError(t, err)

	var row row
	err = pgxscan.Get(ctx, conn, &row, `insert into test (doc) values ($1::jsonb) returning *`, doc)
	assert.NoError(t, err)
	assert.True(t, row.Doc.IsNull())

	err = pgxscan.Get(ctx, conn, &row, `insert into test (doc) values (NULL) returning *`)
	assert.NoError(t, err)
	assert.True(t, row.Doc.IsNull())
}

type testdoc struct {
	Type   string `json:"type"`
	Int    int    `json:"int"`
	String string `json:"string"`
}

func TestDocumentDecodeEncode(t *testing.T) {
	d := testdoc{Type: "test/1", Int: 42, String: "Forty two"}
	doc, err := NewDocument(d)
	assert.NoError(t, err)
	assert.False(t, doc.IsNull())
	assert.Equal(t, "test/1", doc.Type())

	var dd testdoc
	err = doc.Decode(&dd)
	assert.NoError(t, err)
	assert.Equal(t, d, dd)
}

func TestDocumentDecodeEncodeNull(t *testing.T) {
	doc, err := NewDocument(nil)
	assert.NoError(t, err)
	assert.True(t, doc.IsNull())
	assert.Equal(t, "", doc.Type())

	var dd testdoc
	err = doc.Decode(&dd)
	assert.Equal(t, ErrIsNull, err)

	doc, err = NewDocumentRaw(nil)
	assert.NoError(t, err)
	assert.True(t, doc.IsNull())
	assert.Equal(t, "", doc.Type())
}

func TestLargeDocument(t *testing.T) {
	buf, err := os.ReadFile("../../test/DESKTOP-MIMU51J-before.json")
	assert.NoError(t, err)

	buf[len(buf)/2] = '\\'
	buf[len(buf)/2+1] = 'u'
	buf[len(buf)/2+2] = '0'
	buf[len(buf)/2+3] = '0'
	buf[len(buf)/2+4] = '0'
	buf[len(buf)/2+5] = '0'
	dst, err := encode(buf)
	assert.NoError(t, err)
	buf2, err := decode(dst)
	assert.NoError(t, err)
	assert.Equal(t, buf, buf2)
}
