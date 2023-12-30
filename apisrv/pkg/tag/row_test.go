package tag

import (
	"testing"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestParseMetadata(t *testing.T) {
	md := NewMetadata()
	assert.Equal(t, md.Type, MetadataType)
	doc, err := md.Serialize()
	assert.NoError(t, err)
	assert.False(t, doc.IsNull())
	md2, err := parseMetadata(doc)
	assert.NoError(t, err)
	assert.Equal(t, md, *md2)
}

func TestRowGetMetadata(t *testing.T) {
	doc, err := database.NewDocumentRaw([]byte(`{"type":"tag/1"}`))
	assert.NoError(t, err)
	row := Row{Metadata: doc}
	md, err := row.GetMetadata()
	assert.NoError(t, err)
	assert.Equal(t, md.Type, MetadataType)
	md2, _ := row.GetMetadata()
	assert.Same(t, md, md2)
}

func TestRowNoMetadata(t *testing.T) {
	doc, err := database.NewDocument(nil)
	assert.NoError(t, err)
	row := Row{Metadata: doc}
	_, err = row.GetMetadata()
	assert.Equal(t, NoMetadataErr, err)
}

func TestRowIllformedMetadata(t *testing.T) {
	doc, err := database.NewDocumentRaw([]byte("{aaaa"))
	assert.NoError(t, err)
	row := Row{Metadata: doc}
	_, err = row.GetMetadata()
	assert.Error(t, err)
}

func TestRowUnknownMetadata(t *testing.T) {
	doc, err := database.NewDocumentRaw([]byte(`{"type":"tag/9999"}`))
	assert.NoError(t, err)
	row := Row{Metadata: doc}
	_, err = row.GetMetadata()
	assert.Equal(t, UnknownMetadataErr, err)
}
