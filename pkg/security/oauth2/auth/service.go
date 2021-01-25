package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"github.com/google/uuid"
)

type AuthorizationService interface {
	CreateAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (oauth2.Authentication, error)
	CreateAccessToken(ctx context.Context, authentication oauth2.Authentication) (oauth2.AccessToken, error)
	RefreshAccessToken(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error)
}

/****************************
	Default implementation
 ****************************/
type DASOptions func(*DASOption)

type DASOption struct {
	DetailsFactory *common.ContextDetailsFactory
	TokenStore     TokenStore
	TokenEnhancer  TokenEnhancer
	// TODO...
}

// DefaultAuthorizationService implements AuthorizationService
type DefaultAuthorizationService struct {
	detailsFactory *common.ContextDetailsFactory
	tokenStore     TokenStore
	tokenEnhancer  TokenEnhancer
	// TODO...
}

func NewDefaultAuthorizationService(opts...DASOptions) *DefaultAuthorizationService {
	conf := DASOption{
		TokenEnhancer: NewCompositeTokenEnhancer(
			&ExpiryTokenEnhancer{},
			&BasicClaimsTokenEnhancer{},
			&LegacyTokenEnhancer{},
		),
	}
	for _, opt := range opts {
		opt(&conf)
	}
	return &DefaultAuthorizationService{
		detailsFactory: conf.DetailsFactory,
		tokenStore: conf.TokenStore,
		tokenEnhancer: conf.TokenEnhancer,
		// TODO...
	}
}

func (s *DefaultAuthorizationService) CreateAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (ret oauth2.Authentication, err error) {

	var details security.ContextDetails
	if userAuth != nil {
		if details, err = s.detailsFactory.New(ctx, request, userAuth); err != nil {
			return
		}
	} else {
		// TODO create basic details
	}

	oauth := oauth2.NewAuthentication(func(conf *oauth2.AuthConfig) {
		conf.Request = request
		conf.UserAuth = userAuth
		conf.Details = details
	})
	return oauth, nil
}

func (s *DefaultAuthorizationService) RefreshAccessToken(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error) {
	// TODO
	panic("implement me")
}

func (s *DefaultAuthorizationService) CreateAccessToken(c context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	var token *oauth2.DefaultAccessToken

	existing, e := s.tokenStore.ReusableAccessToken(c, oauth)
	if e != nil || existing == nil {
		token = oauth2.NewDefaultAccessToken(uuid.New().String())
	} else if t, ok := existing.(*oauth2.DefaultAccessToken); !ok {
		token = oauth2.FromAccessToken(t)
	} else {
		token = t
	}

	// TODO Enhance token
	enhanced, e := s.tokenEnhancer.Enhance(c, token, oauth)
	if e != nil {
		return nil, e
	}

	// save token
	return s.tokenStore.SaveAccessToken(c, enhanced, oauth)
}
