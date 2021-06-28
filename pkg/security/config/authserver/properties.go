package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "security.auth"
)

//goland:noinspection GoNameStartsWithPackageName
type AuthServerProperties struct {
	Issuer            IssuerProperties    `json:"issuer"`
	RedirectWhitelist utils.StringSet     `json:"redirect-whitelist"`
	Endpoints         EndpointsProperties `json:"endpoints"`
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

type EndpointsProperties struct {
	// TODO check_session and /oauth/error, do we still need them?
	Authorize       string `json:"authorize"`
	Token           string `json:"token"`
	Approval        string `json:"approval"`
	CheckToken      string `json:"check-token"`
	TenantHierarchy string `json:"tenant-hierarchy"`
	Error           string `json:"error"`
	Logout          string `json:"logout"`
	UserInfo        string `json:"user-info"`
	JwkSet          string `json:"jwk-set"`
	SamlMetadata    string `json:"saml-metadata"`
}

//NewAuthServerProperties create a SessionProperties with default values
func NewAuthServerProperties() *AuthServerProperties {
	return &AuthServerProperties{
		Issuer: IssuerProperties{
			Protocol:    "http",
			Domain:      "locahost",
			Port:        8080,
			ContextPath: "",
			IncludePort: true,
		},
		RedirectWhitelist: utils.NewStringSet(),
		Endpoints: EndpointsProperties{
			Authorize:       "/v2/authorize",
			Token:           "/v2/token",
			Approval:        "/v2/approve",
			CheckToken:      "/v2/check_token",
			TenantHierarchy: "/v2/tenant_hierarchy",
			Error:           "/error",
			Logout:          "/v2/logout",
			UserInfo:        "/v2/userinfo",
			JwkSet:          "/v2/jwks",
			SamlMetadata:    "/metadata",
		},
	}
}

//BindAuthServerProperties create and bind AuthServerProperties, with a optional prefix
func BindAuthServerProperties(ctx *bootstrap.ApplicationContext) AuthServerProperties {
	props := NewAuthServerProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AuthServerProperties"))
	}
	return *props
}
