package saml

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const SamlPropertiesPrefix = "security.auth.saml"

type SamlProperties struct {
	RootUrl string `json:"root-url"`
	CertificateFile string `json:"certificate-file"`
	KeyFile string  `json:"key-file"`
	KeyPassword string `json:"key-password"`
}

func NewSamlProperties() *SamlProperties {
	return &SamlProperties{}
}

func BindSamlProperties(ctx *bootstrap.ApplicationContext) SamlProperties {
	props := NewSamlProperties()
	if err := ctx.Config().Bind(props, SamlPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SamlProperties"))
	}
	return *props
}
