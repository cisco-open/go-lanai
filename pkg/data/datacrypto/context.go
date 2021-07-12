package datacrypto

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

var (
	ErrUnsupportedVersion = data.NewDataError(data.ErrorCodeOrmMapping, "unsupported version of encrypted data format")
	ErrInvalidFormat = data.NewDataError(data.ErrorCodeOrmMapping, "invalid encrypted data")
)

/*************************
	Enums
 *************************/

const (
	// V1 is Java compatible data structure
	V1 Version = 1
	// V2 is Generic JSON version, default format of go-lanai
	V2 Version = 2

	defaultVersion = V2
	v1Text         = "1"
	v2Text         = "2"
)

type Version int

// UnmarshalText implements encoding.TextUnmarshaler
func (v *Version) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(string(text))
	switch str {
	case v1Text:
		*v = V1
	case v2Text:
		*v = V2
	case "":
		*v = defaultVersion
	default:
		return fmt.Errorf("unknown encrypted data version: %s", str)
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler with V1 support
func (v *Version) UnmarshalJSON(data []byte) (err error) {
	var i int
	if e := json.Unmarshal(data, &i); e != nil {
		return e
	}

	switch i {
	case int(V1), int(V2):
		*v = Version(i)
	case 0:
		*v = defaultVersion
	default:
		return fmt.Errorf("unknown encrypted data version: %d", i)
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
	Data
 *************************/

// EncryptedRaw is the carrier of encrypted data
// this data type implements gorm.Valuer, schema.GormDataTypeInterface
type EncryptedRaw struct {
	Ver   Version     `json:"v"`
	KeyID uuid.UUID   `json:"kid,omitempty"`
	Alg   Algorithm   `json:"alg,omitempty"`
	Raw   interface{} `json:"d,omitempty"`
}

// GormDataType implements schema.GormDataTypeInterface
func (EncryptedRaw) GormDataType() string {
	return "jsonb"
}

//// GormValue implements  gorm.Valuer
//func (d EncryptedRaw) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
//	return clause.Expr{SQL: "?", Vars: []interface{}{d}}
//}

/*************************
	Interface
 *************************/

type Encryptor interface {
	// Encrypt encrypt given "v" and populate EncryptedRaw.Raw
	// The process may read EncryptedRaw.Alg and EncryptedRaw.KeyID and update EncryptedRaw.Ver
	Encrypt(ctx context.Context, v interface{}, raw *EncryptedRaw) error

	// Decrypt reads EncryptedRaw and populate the decrypted data into given "v"
	// if v is not pointer type, this method may return error
	Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error
}
