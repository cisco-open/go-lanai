package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
)

const (
	AuthServerPropertiesPrefix = "security.auth"
)

type AuthServerProperties struct {
	Issuer            IssuerProperties `json:"issuer"`
	RedirectWhitelist utils.StringSet  `json:"redirect-whitelist"`
}

type IssuerProperties struct {
	//  the protocol which is either http or https
	Protocol string `json:"protocol"`
	// This server's host name
	// Used to build the entity base url. The entity url identifies this auth server in a SAML exchange and OIDC exchange.
	Domain string `json:"domain"`
	Port   int    `json:"port"`
	// Context base path for this server
	// Used to build the entity base url. The entity url identifies this auth server in a SAML exchange.
	ContextPath string `json:"context-path"`
	IncludePort bool   `json:"include-port"`
}

//NewAuthServerProperties create a SessionProperties with default values
func NewAuthServerProperties() *AuthServerProperties {
	return &AuthServerProperties {
		Issuer: IssuerProperties{
			Protocol:    "http",
			Domain:      "locahost",
			Port:        8080,
			ContextPath: "",
			IncludePort: true,
		},
		RedirectWhitelist: utils.NewStringSet(),
	}
}

//BindAuthServerProperties create and bind AuthServerProperties, with a optional prefix
func BindAuthServerProperties(ctx *bootstrap.ApplicationContext) AuthServerProperties {
	props := NewAuthServerProperties()
	if err := ctx.Config().Bind(props, AuthServerPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AuthServerProperties"))
	}
	return *props
}
