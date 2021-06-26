package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "security.idp.saml"
)

type SamlAuthProperties struct {
	Enabled   bool                       `json:"enabled"`
	Endpoints SamlAuthEndpointProperties `json:"endpoints"`
}

type SamlAuthEndpointProperties struct {}

func NewSamlAuthProperties() *SamlAuthProperties {
	return &SamlAuthProperties{}
}

func BindSamlAuthProperties(ctx *bootstrap.ApplicationContext) SamlAuthProperties {
	props := NewSamlAuthProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SamlAuthProperties"))
	}
	return *props
}
