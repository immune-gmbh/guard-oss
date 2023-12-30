package tag

import (
	"errors"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

var (
	// database columns fetched by default
	defaultColumns = []string{
		"id", "key", "value", "metadata", "organization_id",
	}
	allowedColumns = map[string]bool{
		"id": true, "key": true, "value": true, "metadata": true, "organization_id": true,
	}

	NoMetadataErr      = errors.New("no metadata fetched")
	UnknownMetadataErr = errors.New("unknown metadata type")
)

// database representation of a tag
type Row struct {
	// columns
	Id           string            `db:"id"`
	Key          string            `db:"key"`
	Value        *string           `db:"value"`
	Metadata     database.Document `db:"metadata"`
	Organization int64             `db:"organization_id"`

	// relations and "virtual" columns
	Devices []int64 `db:"devices"`
	Score   float32 `db:"score"` // similarity score when using Text()

	// cache
	parsedMetadata *Metadata
}

// returns the (cached) metadata. fails if the "metadata" column wasn't fetched
func (row *Row) GetMetadata() (*Metadata, error) {
	if row.parsedMetadata != nil {
		return row.parsedMetadata, nil
	} else if !row.Metadata.IsNull() {
		meta, err := parseMetadata(row.Metadata)
		row.parsedMetadata = meta
		return meta, err
	} else {
		return nil, NoMetadataErr
	}
}

var (
	MetadataType = "tag/1"
)

// free form tag metadata encoded as a JSONB column
type Metadata struct {
	Type string `json:"type"`
}

func NewMetadata() Metadata {
	return Metadata{Type: MetadataType}
}

func (m Metadata) Serialize() (database.Document, error) {
	return database.NewDocument(m)
}

func parseMetadata(doc database.Document) (*Metadata, error) {
	var metadata Metadata
	err := doc.Decode(&metadata)
	if err != nil {
		return nil, err
	}

	switch metadata.Type {
	case MetadataType:
		return &metadata, nil
	default:
		return nil, UnknownMetadataErr
	}
}

func FromRow(row *Row) (*api.Tag, error) {
	tag := api.Tag{
		Id:    row.Id,
		Key:   row.Key,
		Score: row.Score,
	}

	return &tag, nil
}
