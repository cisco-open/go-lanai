package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
)

type AuthorizationService interface {
	CreateAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (oauth2.Authentication, error)
	CreateAccessToken(ctx context.Context, authentication oauth2.Authentication) (oauth2.AccessToken, error)
	RefreshAccessToken(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error)
}

/****************************
	Default implementation
 ****************************/
type AuthorizationServiceOptionsFunc func(*AuthServiceConfig)

type AuthServiceConfig struct {
	TokenStore TokenStore
	// TODO...
}

// DetailsPersistentAuthorizationService implements AuthorizationService
type DetailsPersistentAuthorizationService struct {
	tokenStore TokenStore
	// TODO...
}

func NewDetailsPersistentAuthorizationService(opts...AuthorizationServiceOptionsFunc) *DetailsPersistentAuthorizationService {
	conf := AuthServiceConfig{}
	for _, opt := range opts {
		opt(&conf)
	}
	return &DetailsPersistentAuthorizationService{
		tokenStore: conf.TokenStore,
		// TODO...
	}
}

func (s *DetailsPersistentAuthorizationService) CreateAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (oauth2.Authentication, error) {
	oauth := oauth2.NewAuthentication(func(conf *oauth2.AuthConfig) {
		conf.Request = request
		conf.UserAuth = userAuth
	})
	return oauth, nil
}

func (s *DetailsPersistentAuthorizationService) RefreshAccessToken(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error) {
	panic("implement me")
}

func (s *DetailsPersistentAuthorizationService) CreateAccessToken(c context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	return oauth2.NewDefaultAccessToken(fmt.Sprintf("TODO for [%v]", oauth.Principal())), nil
}
