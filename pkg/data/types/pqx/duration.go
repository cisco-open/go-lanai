package pqx

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"database/sql/driver"
	"fmt"
	"time"
)

// utils.Duration is also an alias of time.Duration
type Duration utils.Duration

// driver.Valuer
func (d Duration) Value() (driver.Value, error) {
	return time.Duration(d).String(), nil
}

// sql.Scanner
func (d *Duration) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		*d = Duration(utils.ParseDuration(string(src)))
	case string:
		*d = Duration(utils.ParseDuration(src))
	case int, int8, int16, int32, int64:
		// TODO review how convert numbers to Duration
		*d = Duration(src.(int64))
	case nil:
		return nil
	default:
		return data.NewDataError(data.ErrorCodeOrmMapping,
			fmt.Sprintf("pqx: unable to convert data type %T to Duration", src))
	}
	return nil
}

// encoding.TextMarshaler
func (d Duration) MarshalText() (text []byte, err error) {
	return utils.Duration(d).MarshalText()
}

// encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	return (*utils.Duration)(d).UnmarshalText(text)
}