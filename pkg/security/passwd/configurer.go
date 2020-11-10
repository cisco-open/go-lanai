package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"reflect"
)

var PasswordAuthConfigurerType = reflect.TypeOf((*PasswordAuthConfigurer)(nil))

// We currently don't have any stuff to configure
type PasswordAuthFeature struct {
	accountStore security.AccountStore
	passwordEncoder PasswordEncoder
}

func (f *PasswordAuthFeature) ConfigurerType() reflect.Type {
	return PasswordAuthConfigurerType
}

func (f *PasswordAuthFeature) AccountStore(as security.AccountStore) *PasswordAuthFeature {
	f.accountStore = as
	return f
}

func (f *PasswordAuthFeature) PasswordEncoder(pe PasswordEncoder) *PasswordAuthFeature {
	f.passwordEncoder = pe
	return f
}

func Configure(ws security.WebSecurity) *PasswordAuthFeature {
	feature := &PasswordAuthFeature{}
	if fm, ok := ws.(security.FeatureModifier); ok {
		_ = fm.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

type PasswordAuthConfigurer struct {
	authenticator security.Authenticator
}

func newPasswordAuthConfigurer(auth security.Authenticator) *PasswordAuthConfigurer {
	return &PasswordAuthConfigurer{
		authenticator: auth,
	}
}

func (pac *PasswordAuthConfigurer) Build(f security.Feature) ([]security.MiddlewareTemplate, error) {
	// TODO error handling
	passwdFeature := f.(*PasswordAuthFeature)
	auth := NewAuthenticator(passwdFeature.accountStore, passwdFeature.passwordEncoder)
	pac.authenticator.(*security.CompositeAuthenticator).Add(auth)
	return []security.MiddlewareTemplate{}, nil
}
