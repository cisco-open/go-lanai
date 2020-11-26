package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
)

const (
	PasswordAuthenticatorFeatureId = FeatureId("passwdAuth")
)

// FeatureId is ordered
type FeatureId string

// order.Ordered interface
func (FeatureId) Order() int {
	return order.Highest
}

// We currently don't have any stuff to configure
type PasswordAuthFeature struct {
	accountStore security.AccountStore
	passwordEncoder PasswordEncoder
}

func (f *PasswordAuthFeature) Identifier() security.FeatureIdentifier {
	return PasswordAuthenticatorFeatureId
}

func (f *PasswordAuthFeature) AccountStore(as security.AccountStore) *PasswordAuthFeature {
	f.accountStore = as
	return f
}

func (f *PasswordAuthFeature) PasswordEncoder(pe PasswordEncoder) *PasswordAuthFeature {
	f.passwordEncoder = pe
	return f
}

// Standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *PasswordAuthFeature {
	feature := &PasswordAuthFeature{}
	if fm, ok := ws.(security.FeatureModifier); ok {
		_ = fm.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *PasswordAuthFeature {
	return &PasswordAuthFeature{}
}

type PasswordAuthConfigurer struct {
	accountStore security.AccountStore
	passwordEncoder PasswordEncoder
}

func newPasswordAuthConfigurer(store security.AccountStore, encoder PasswordEncoder) *PasswordAuthConfigurer {
	return &PasswordAuthConfigurer {
		accountStore: store,
		passwordEncoder: encoder,
	}
}

func (pac *PasswordAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Validate
	if err := pac.validate(feature.(*PasswordAuthFeature), ws); err != nil {
		return err
	}
	f := feature.(*PasswordAuthFeature)


	// set defaults if necessary
	pac.setDefaultsIfNecessary(f)

	auth := NewAuthenticator(f.accountStore, f.passwordEncoder)
	ws.Authenticator().(*security.CompositeAuthenticator).Add(auth)
	return nil
}

func (pac *PasswordAuthConfigurer) validate(f *PasswordAuthFeature, ws security.WebSecurity) error {

	if _,ok := ws.Authenticator().(*security.CompositeAuthenticator); !ok {
		return fmt.Errorf("unable to add password authenticator to %T", ws.Authenticator())
	}

	if f.accountStore == nil && pac.accountStore == nil {
		return fmt.Errorf("unable to create password authenticator: account store is not set")
	}
	return nil
}

func (pac *PasswordAuthConfigurer) setDefaultsIfNecessary(f *PasswordAuthFeature) {
	if f.accountStore == nil {
		f.accountStore = pac.accountStore
	}

	if f.passwordEncoder == nil {
		f.passwordEncoder = pac.passwordEncoder
	}
}
