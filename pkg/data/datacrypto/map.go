package datacrypto

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"database/sql/driver"
)

type EncryptedMap struct {
	EncryptedRaw
	Data map[string]interface{} `json:"-"`
}

// Value implements driver.Valuer
func (d EncryptedMap) Value() (driver.Value, error) {
	// TODO do encryption
	if d.Ver < V2 {
		return nil, data.NewDataError(data.ErrorCodeOrmMapping, "ver 1 of encrypted data is not supported")
	}
	return pqx.JsonbValue(d)
}


//TODO


// Scan implements sql.Scanner
func (d *EncryptedRaw) Scan(src interface{}) error {
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

