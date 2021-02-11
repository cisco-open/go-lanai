package tokenauth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"reflect"
)

/******************************
	security.Authenticator
******************************/
type Authenticator struct {
	tokenStoreReader oauth2.TokenStoreReader
}

type AuthenticatorOptions func(opt *AuthenticatorOption)

type AuthenticatorOption struct {
	TokenStoreReader oauth2.TokenStoreReader
}

func NewAuthenticator(options ...AuthenticatorOptions) *Authenticator {
	opt := AuthenticatorOption{}
	for _, f := range options {
		if f != nil {
			f(&opt)
		}
	}
	return &Authenticator{
		tokenStoreReader:      opt.TokenStoreReader,
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, candidate security.Candidate) (security.Authentication, error) {
	can, ok := candidate.(*BearerToken)
	if !ok {
		return nil, nil
	}

	// TODO add remote check_token endpoint support
	token, e := a.tokenStoreReader.ReadAccessToken(ctx, can.Token)
	if e != nil {
		return nil, oauth2.NewInvalidAccessTokenError("token is invalid", e)
	} else if token.Expired() {
		return nil, oauth2.NewInvalidAccessTokenError("token is expired", e)
	}

	auth, e := a.tokenStoreReader.ReadAuthentication(ctx, token)
	if e != nil {
		return nil, oauth2.NewInvalidAccessTokenError("token unknown", e)
	}

	// perform some checks
	switch {
	case auth.State() < security.StateAuthenticated:
		return nil, oauth2.NewInvalidAccessTokenError("token is not associated with an authenticated session")
	case auth.OAuth2Request().ClientId() == "":
		return nil, oauth2.NewInvalidAccessTokenError("token is not issued to a valid client")
	case auth.UserAuthentication() != nil && reflect.ValueOf(auth.UserAuthentication().Principal()).IsZero():
		return nil, oauth2.NewInvalidAccessTokenError("token is not authorized by a valid user")
	}

	return auth, nil
}