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

// Sharing is a Model type that stores mapping between user IDs and a list of allowed permissions as JSONB map
// This type works with OPA sharing policy
type Sharing map[uuid.UUID][]SharedPermission

// Value implements driver.Valuer
func (s Sharing) Value() (driver.Value, error) {
	return pqx.JsonbValue(s)
}

// Scan implements sql.Scanner
func (s *Sharing) Scan(src interface{}) error {
	return pqx.JsonbScan(src, s)
}

func (s Sharing) GormDataType() string {
	return "jsonb"
}

func (s Sharing) Share(userID uuid.UUID, perms ...SharedPermission) {
	if len(perms) == 0 {
		delete(s, userID)
	} else {
		s[userID] = perms
	}
}
