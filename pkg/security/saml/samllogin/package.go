package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var logger = log.New("SAML.Auth")

var SamlAuthModule = &bootstrap.Module{
	Name: "saml authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(SamlAuthModule)

	gob.Register((*samlAssertionAuthentication)(nil))
}

func register(init security.Registrar, properties saml.SamlProperties,
	serverProps web.ServerProperties, idpManager idp.IdentityProviderManager,
	accountStore security.FederatedAccountStore) {

	configurer := newSamlAuthConfigurer(properties, serverProps, idpManager, accountStore)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}