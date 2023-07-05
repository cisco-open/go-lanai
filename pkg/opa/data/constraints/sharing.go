package constraints

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"database/sql/driver"
	"github.com/google/uuid"
)

const (
	SharedPermissionRead = SharedPermission(opa.OpRead)
	SharedPermissionWrite = SharedPermission(opa.OpWrite)
	SharedPermissionDelete  = SharedPermission(opa.OpDelete)
)

type SharedPermission opa.ResourceOperation

type Sharing map[uuid.UUID][]SharedPermission

// Value implements driver.Valuer
func (t Sharing) Value() (driver.Value, error) {
	return pqx.JsonbValue(t)
}

// Scan implements sql.Scanner
func (t *Sharing) Scan(src interface{}) error {
	return pqx.JsonbScan(src, t)
}

func (t Sharing) GormDataType() string {
	return "jsonb"
}
