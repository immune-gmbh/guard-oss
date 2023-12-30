package change

import (
	"fmt"
	"time"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
)

type Row struct {
	Id             int64     `db:"id"`
	Type           string    `db:"type"`
	Actor          *string   `db:"actor"`
	Comment        *string   `db:"comment"`
	Timestamp      time.Time `db:"timestamp"`
	OrganizationId int64     `db:"organization_id"`
	DeviceId       *int64    `db:"device_id"`
	KeyId          *int64    `db:"key_id"`
}

func FromRow(row *Row) api.Change {
	var dev *api.Device
	if row.DeviceId != nil {
		dev = &api.Device{Id: fmt.Sprintf("%d", *row.DeviceId)}
	}
	return api.Change{
		Id:        fmt.Sprintf("%d", row.Id),
		Type:      row.Type,
		Actor:     row.Actor,
		Comment:   row.Comment,
		Timestamp: row.Timestamp,
		Device:    dev,
	}
}
