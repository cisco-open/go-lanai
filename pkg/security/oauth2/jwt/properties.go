package jwt

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
	"strings"
)

/***********************
	Crypto
************************/
const CryptoKeysPropertiesPrefix = "security"

const (
	KeyFileFormatPem KeyFormatType = "pem"
)

type KeyFormatType string

type CryptoProperties struct {
	Keys map[string]CryptoKeyProperties `json:"keys"`
	Jwt JwtProperties `json:"jwt"`
}

type JwtProperties struct {
	KeyName string `json:"key-name"`
}

type CryptoKeyProperties struct {
	Id        string `json:"id"`
	KeyFormat string `json:"format"`
	Location  string `json:"file"`
	Password  string `json:"password"`
}

func (p CryptoKeyProperties) Format() KeyFormatType {
	return KeyFormatType(strings.ToLower(p.KeyFormat))
}

//CryptoProperties create a SessionProperties with default values
func NewCryptoProperties() *CryptoProperties {
	return &CryptoProperties {
		Keys: map[string]CryptoKeyProperties{},
	}
}

//BindCryptoProperties create and bind CryptoProperties, with a optional prefix
func BindCryptoProperties(ctx *bootstrap.ApplicationContext) CryptoProperties {
	props := NewCryptoProperties()
	if err := ctx.Config().Bind(props, CryptoKeysPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CryptoProperties"))
	}
	return *props
}
