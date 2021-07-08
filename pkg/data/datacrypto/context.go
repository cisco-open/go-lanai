package datacrypto

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

/*************************
	Enums
 *************************/

const (
	// V1 is Java compatible data structure
	V1 Version = "1"
	// V2 is Generic JSON version, default format of go-lanai
	V2 Version = "2"

	defaultVersion = V2
)

type Version string

// UnmarshalText implements encoding.TextUnmarshaler
func (v *Version) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(string(text))
	switch str {
	case string(V1):
		*v = V1
	case string(V2):
		*v = V2
	case "":
		*v = defaultVersion
	default:
		return fmt.Errorf("unknown encrypted data version: %s", str)
	}
	return nil
}

const (
	AlgPlain   Algorithm = "p"
	AlgVault   Algorithm = "e" // this value is compatible with Java counterpart
	defaultAlg           = AlgPlain
)

type Algorithm string

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Algorithm) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(string(text))
	switch str {
	case string(AlgPlain):
		*a = AlgPlain
	case string(AlgVault):
		*a = AlgVault
	case "":
		*a = defaultAlg
	default:
		return fmt.Errorf("unknown encrypted data algorithm: %s", str)
	}
	return nil
}

/*************************
	Types
 *************************/

type EncryptedMap struct {
	EncryptedData
	Data map[string]interface{} `json:"-"`
}

//TODO

// Value implements driver.Valuer
func (d EncryptedMap) Value() (driver.Value, error) {
	// TODO
	return d, nil
}

// Scan implements sql.Scanner
func (d *EncryptedData) Scan(src interface{}) error {
	//switch src := src.(type) {
	//case []byte:
	//	*d = Duration(utils.ParseDuration(string(src)))
	//case string:
	//	*d = Duration(utils.ParseDuration(src))
	//case int, int8, int16, int32, int64:
	//	// TODO review how convert numbers to Duration
	//	*d = Duration(src.(int64))
	//case nil:
	//	return nil
	//default:
	//	return data.NewDataError(data.ErrorCodeOrmMapping,
	//		fmt.Sprintf("pqx: unable to convert data type %T to Duration", src))
	//}
	return nil
}
