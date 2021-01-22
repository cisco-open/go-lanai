package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

var logger = log.GetNamedLogger("samllogin")

var SamlAuthModule = &bootstrap.Module{
	Name: "saml authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Provide(bindProperties),
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(SamlAuthModule)

	gob.Register((*samlAssertionAuthentication)(nil))
}

func register(init security.Registrar, properties ServiceProviderProperties,
	serverProps web.ServerProperties, idpManager IdentityProviderManager,
	accountStore security.FederatedAccountStore) {

	configurer := newSamlAuthConfigurer(properties, serverProps, idpManager, accountStore)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
func bindProperties(ctx *bootstrap.ApplicationContext) ServiceProviderProperties {
	props := NewServiceProviderProperties()
	if err := ctx.Config().Bind(props, ServiceProviderPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ServiceProviderProperties"))
	}
	return *props
}
