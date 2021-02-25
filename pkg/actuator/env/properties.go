package env

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
)

const (
	EnvPropertiesPrefix = "management.endpoint.env"
)

type EnvProperties struct {
	// KeysToSanitize holds list of regular expressions
	KeysToSanitize utils.StringSet `json:"keys-to-sanitize"`
}

//NewSessionProperties create a SessionProperties with default values
func NewEnvProperties() *EnvProperties {
	return &EnvProperties{
		KeysToSanitize: utils.NewStringSet(
			`.*password.*`, `.*secret.*`, `key`,
			`.*credentials.*`, `vcap_services`, `sun.java.command`,
		),
	}
}

//BindHealthProperties create and bind SessionProperties, with a optional prefix
func BindEnvProperties(ctx *bootstrap.ApplicationContext) EnvProperties {
	props := NewEnvProperties()
	if err := ctx.Config().Bind(props, EnvPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind EnvProperties"))
	}
	return *props
}

