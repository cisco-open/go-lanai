package pqx

import (
	"database/sql/driver"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UUIDArray []uuid.UUID

// driver.Valuer
func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return pq.StringArray(a.Strings()).Value()
}

// sql.Scanner
func (a *UUIDArray) Scan(src interface{}) error {
	if a == nil {
		return nil
	}

	strArray := &pq.StringArray{}
	if e := strArray.Scan(src); e != nil {
		return e
	}
	uuids := make(UUIDArray, len(*strArray))
	for i, v := range *strArray {
		var e error
		if uuids[i], e = uuid.Parse(v); e != nil {
			return fmt.Errorf("pq: cannot convert %T to UUIDArray - %v", src, e)
		}
	}
	*a = uuids
	return nil
}

func (a UUIDArray) Strings() []string {
	strArray := make(pq.StringArray, len(a))
	for i, v := range a{
		strArray[i] = v.String()
	}
	return strArray
}
