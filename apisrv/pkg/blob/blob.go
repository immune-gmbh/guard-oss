// Storage for large, binary objects like BIOS flash images. Items are stored
// in S3 compatible backends and are deduplicated based on their contents on a
// best effort basis. Their metadata is kept in a SQL database.
package blob

import (
	"fmt"
	"time"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/database"
)

const (
	BIOS              = "bios"
	WindowsExecutable = "windows-exeutable"
	UEFIApp           = "uefi-app"
	SMBIOS            = "smbios"
	ACPI              = "acpi"
	TXT               = "txt"
	Eventlog          = "eventlog"
)

// Metadata for a blob in S3. Contains necessary information to locate the blob
// but not the blob itself.
type Row struct {
	Id          string            `db:"id"`
	Digest      string            `db:"digest"`
	Snowflake   string            `db:"snowflake"`
	Namespace   string            `db:"namespace"`
	RawMetadata database.Document `db:"metadata"`
	CreatedAt   time.Time         `db:"created_at"`
}

func (r *Row) Filename() string {
	return fmt.Sprintf("%s-%s", r.Digest, r.Snowflake)
}

func (r *Row) Metadata() (*Metadata, error) {
	return metadataFromRow(r.RawMetadata)
}

const (
	MetadataType = "blob.metadata/1"
)

// Misc information on the blob contents and the events surrounding it
// creation. May change its schema frequently, unlike Row.
type Metadata struct {
	Type string `json:"type"`
	Kind string `json:"kind"`
}

func metadataFromRow(doc database.Document) (*Metadata, error) {
	if doc.IsNull() {
		return nil, nil
	}

	switch doc.Type() {
	case MetadataType:
		var metadata Metadata
		err := doc.Decode(&metadata)
		if err != nil {
			return nil, err
		}
		return &metadata, nil

	default:
		return nil, UnknownMetadataErr
	}
}

func (m *Metadata) ToRow() (database.Document, error) {
	return database.NewDocument(*m)
}
