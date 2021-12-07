package pqcrypt

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

const (
	PropertiesPrefix = "data.encryption"
)

type DataEncryptionProperties struct {
	Enabled bool          `json:"enabled"`
	Key     KeyProperties `json:"key"`
}

type KeyProperties struct {
	Type                 string `json:"type"`
	Exportable           bool   `json:"exportable"`
	AllowPlaintextBackup bool   `json:"allow-plaintext-backup"`
}

// https://www.vaultproject.io/api/secret/transit#create-key
const (
	KeyTypeAES128   = "aes128-gcm96"
	KeyTypeAES256   = "aes256-gcm96"
	KeyTypeChaCha20 = "chacha20-poly1305"
	KeyTypeED25519  = "ed25519"
	KeyTypeECDSA256 = "ecdsa-p256"
	KeyTypeECDSA384 = "ecdsa-p384"
	KeyTypeECDSA521 = "ecdsa-p521"
	KeyTypeRSA2048  = "rsa-2048"
	KeyTypeRSA3072  = "rsa-3072"
	KeyTypeRSA4096  = "rsa-4096"

	defaultKeyType = KeyTypeAES256
)

var supportedKeyTypes = utils.NewStringSet(
	KeyTypeAES128, KeyTypeAES256, KeyTypeChaCha20,
	KeyTypeED25519, KeyTypeECDSA256, KeyTypeECDSA384, KeyTypeECDSA521,
	KeyTypeRSA2048, KeyTypeRSA3072, KeyTypeRSA4096,
)

type KeyType string

// UnmarshalText implements encoding.TextUnmarshaler
func (t *KeyType) UnmarshalText(text []byte) error {
	str := strings.ToLower(strings.TrimSpace(string(text)))
	switch {
	case len(str) == 0:
		*t = defaultKeyType
	case supportedKeyTypes.Has(str):
		*t = KeyType(str)
	default:
		return fmt.Errorf("unknown encryption key type: %s", str)
	}
	return nil
}

//NewDataEncryptionProperties create a CockroachProperties with default values
func NewDataEncryptionProperties() *DataEncryptionProperties {
	return &DataEncryptionProperties{
		Enabled: false,
		Key: KeyProperties{
			Type:                 defaultKeyType,
			Exportable:           false,
			AllowPlaintextBackup: false,
		},
	}
}

//BindDataEncryptionProperties create and bind SessionProperties, with a optional prefix
func BindDataEncryptionProperties(ctx *bootstrap.ApplicationContext) DataEncryptionProperties {
	props := NewDataEncryptionProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DataEncryptionProperties"))
	}
	return *props
}
