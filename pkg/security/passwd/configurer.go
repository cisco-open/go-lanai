package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

var (
	PasswordAuthenticatorFeatureId = security.FeatureId("passwdAuth", security.FeatureOrderAuthenticator)
)

type PasswordAuthConfigurer struct {
	accountStore security.AccountStore
	passwordEncoder PasswordEncoder
	redisClient redis.Client
}

func newPasswordAuthConfigurer(store security.AccountStore, encoder PasswordEncoder, redisClient redis.Client) *PasswordAuthConfigurer {
	return &PasswordAuthConfigurer {
		accountStore:    store,
		passwordEncoder: encoder,
		redisClient:     redisClient,
	}
}

func (pac *PasswordAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := pac.validate(feature.(*PasswordAuthFeature), ws); err != nil {
		return err
	}
	f := feature.(*PasswordAuthFeature)

	// Build authenticator
	ctx := context.Background()
	defaults := &builderDefaults{
		accountStore: pac.accountStore,
		passwordEncoder: pac.passwordEncoder,
		redisClient: pac.redisClient,
	}
	authenticator, err := NewAuthenticatorBuilder(f, defaults).Build(ctx)
	if err != nil {
		return err
	}

	// Add authenticator to WS, flatten if multiple
	if composite, ok := authenticator.(*security.CompositeAuthenticator); ok {
		ws.Authenticator().(*security.CompositeAuthenticator).Merge(composite)
	} else {
		ws.Authenticator().(*security.CompositeAuthenticator).Add(authenticator)
	}

	return nil
}

func (pac *PasswordAuthConfigurer) validate(f *PasswordAuthFeature, ws security.WebSecurity) error {

	if _,ok := ws.Authenticator().(*security.CompositeAuthenticator); !ok {
		return fmt.Errorf("unable to add password authenticator to %T", ws.Authenticator())
	}

	if f.accountStore == nil && pac.accountStore == nil {
		return fmt.Errorf("unable to create password authenticator: account accountStore is not set")
	}
	return nil
}



