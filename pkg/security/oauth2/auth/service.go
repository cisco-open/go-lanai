package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/google/uuid"
	"time"
)

var (

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
	ClientStore    oauth2.OAuth2ClientStore
	AccountStore   security.AccountStore
	TenantStore    security.TenantStore
	ProviderStore  security.ProviderStore
	HierarchyStore security.TenantHierarchyStore
	TokenStore     TokenStore
	TokenEnhancer  TokenEnhancer
	// TODO...
}

// DefaultAuthorizationService implements AuthorizationService
type DefaultAuthorizationService struct {
	detailsFactory *common.ContextDetailsFactory
	clientStore    oauth2.OAuth2ClientStore
	accountStore   security.AccountStore
	tenantStore    security.TenantStore
	providerStore  security.ProviderStore
	hierarchyStore security.TenantHierarchyStore
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
		clientStore:    conf.ClientStore,
		accountStore:   conf.AccountStore,
		tenantStore:    conf.TenantStore,
		providerStore:  conf.ProviderStore,
		hierarchyStore: conf.HierarchyStore,
		tokenStore:     conf.TokenStore,
		tokenEnhancer:  conf.TokenEnhancer,
		// TODO...
	}
}

type authFacts struct {
	request  oauth2.OAuth2Request
	client   oauth2.OAuth2Client
	account  security.Account
	tenant   *security.Tenant
	provider *security.Provider
}

func (s *DefaultAuthorizationService) CreateAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (oauth oauth2.Authentication, err error) {

	now := time.Now().UTC()

	facts, e := s.loadAndVerifyFacts(ctx, request, userAuth)
	if e != nil {
		return nil, e
	}

	mutableCtx, ok := ctx.(utils.MutableContext);
	if  !ok {
		return nil, newImmutableContextError()
	}

	mutableCtx.Set(oauth2.CtxKeyAuthenticatedClient, facts.client)
	mutableCtx.Set(oauth2.CtxKeyAuthenticatedAccount, facts.account)
	mutableCtx.Set(oauth2.CtxKeyAuthorizedTenant, facts.tenant)
	mutableCtx.Set(oauth2.CtxKeyAuthorizedProvider, facts.provider)
	mutableCtx.Set(oauth2.CtxKeyAuthorizationIssueTime, now)

	// expiry
	expiry := s.determineExpiryTime(ctx, request, facts)
	if !expiry.IsZero() {
		mutableCtx.Set(oauth2.CtxKeyAuthorizationExpiryTime, expiry)
	}

	// auth time
	authTime := s.determineAuthenticationTime(ctx, userAuth)
	if !authTime.IsZero() {
		mutableCtx.Set(oauth2.CtxKeyAuthenticationTime, authTime)
	}

	if userAuth == nil {
		mutableCtx.Set(oauth2.CtxKeyAuthenticationTime, now)
	} else {
		// TODO, grab authentication time
	}

	// create context details
	details, err := s.detailsFactory.New(ctx, request)
	if err != nil {
		return
	}

	oauth = oauth2.NewAuthentication(func(conf *oauth2.AuthConfig) {
		conf.Request = request
		conf.UserAuth = userAuth
		conf.Details = details
	})
	return
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

	enhanced, e := s.tokenEnhancer.Enhance(c, token, oauth)
	if e != nil {
		return nil, e
	}

	// save token
	return s.tokenStore.SaveAccessToken(c, enhanced, oauth)
}

/****************************
	Helpers
 ****************************/
func (f *DefaultAuthorizationService) loadAndVerifyFacts(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (*authFacts, error) {
	client := RetrieveAuthenticatedClient(ctx)
	if client == nil {
		return nil, newInvalidClientError()
	}

	if userAuth == nil {
		return &authFacts{ client: client }, nil
	}

	account, e := f.loadAccount(ctx, request, userAuth)
	if e != nil {
		return nil, e
	}

	tenant, e := f.loadTenant(ctx, request, account)
	if e != nil {
		return nil, e
	}

	provider, e := f.loadProvider(ctx, request, tenant)
	if e != nil {
		return nil, e
	}

	return &authFacts{
		request: request,
		client: client,
		account: account,
		tenant: tenant,
		provider: provider,
	}, nil
}

func (f *DefaultAuthorizationService) loadAccount(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (security.Account, error) {
	// sanity check, this should not happen
	if userAuth.State() < security.StateAuthenticated || userAuth.Principal() == nil {
		return nil, newUnauthenticatedUserError()
	}

	// we want to reload user's account
	principal := userAuth.Principal()
	var username string
	switch principal.(type) {
	case security.Account:
		username = principal.(security.Account).Username()
	case string:
		username = principal.(string)
	case fmt.Stringer:
		username = principal.(fmt.Stringer).String()
	default:
		return nil, newInvalidUserError(fmt.Sprintf("unsupported principal type %T", principal))
	}

	acct, e := f.accountStore.LoadAccountByUsername(ctx, username)
	if e != nil {
		return nil, newInvalidUserError(e)
	}
	return acct, nil
}

func (f *DefaultAuthorizationService) loadTenant(ctx context.Context, request oauth2.OAuth2Request, account security.Account) (*security.Tenant, error) {
	tenancy, ok := account.(security.AccountTenancy)
	if !ok {
		return nil, newInvalidTenantForUserError(fmt.Sprintf("account [%T] does not provide tenancy information", account))
	}

	// extract tenant id or name
	tenantId, idOk := request.Parameters()[oauth2.ParameterTenantId]
	tenantName, nOk := request.Parameters()[oauth2.ParameterTenantName]
	if (!idOk || tenantId == "") && (!nOk || tenantName == "") {
		tenantId = tenancy.DefaultTenantId()
	}

	var tenant *security.Tenant
	var e error
	if tenantId != "" {
		tenant, e = f.tenantStore.LoadTenantById(ctx, tenantId)
		if e != nil {
			return nil, newInvalidTenantForUserError(fmt.Sprintf("user [%s] does not access tenant with id [%s]", account.Username(), tenantId))
		}
	} else {
		tenant, e = f.tenantStore.LoadTenantByName(ctx, tenantName)
		if e != nil {
			return nil, newInvalidTenantForUserError(fmt.Sprintf("user [%s] does not access tenant with name [%s]", account.Username(), tenantName))
		}
	}

	// TODO check tenant access here (both client and user)

	return tenant, nil
}

func (f *DefaultAuthorizationService) loadProvider(ctx context.Context, request oauth2.OAuth2Request, tenant *security.Tenant) (*security.Provider, error) {
	providerId := tenant.ProviderId
	if providerId == "" {
		return nil, newInvalidProviderError("provider ID is not avalilable")
	}

	provider, e := f.providerStore.LoadProviderById(ctx, providerId)
	if e != nil {
		return nil, newInvalidProviderError(fmt.Sprintf("tenant [%s]'s provider is invalid", tenant.Name))
	}
	return provider, nil
}

func (f *DefaultAuthorizationService) determineExpiryTime(ctx context.Context, request oauth2.OAuth2Request, facts *authFacts) (expiry time.Time) {
	if facts.client.AccessTokenValidity() == 0 {
		return
	}

	issueTime := ctx.Value(oauth2.CtxKeyAuthorizationIssueTime).(time.Time)

	// TODO When switching context, expiry should no later than original expiry time
	return issueTime.Add(facts.client.AccessTokenValidity())
}

func (f *DefaultAuthorizationService) determineAuthenticationTime(ctx context.Context, userAuth security.Authentication) (authTime time.Time) {
	if userAuth == nil {
		return
	}

	details, ok := userAuth.Details().(map[string]interface{})
	if !ok {
		return
	}

	v, ok := details[security.DetailsKeyAuthTime]
	if !ok {
		return
	}

	if t, ok := v.(time.Time); ok {
		return t
	}
	return
}

/****************************
	Errors
 ****************************/
func newImmutableContextError(reasons ...interface{}) error {
	return oauth2.NewInternalError("context is not mutable", reasons...)
}

func newInvalidClientError(reasons ...interface{}) error {
	return oauth2.NewInternalError("trying authroize with unknown client", reasons...)
}

func newUnauthenticatedUserError(reasons ...interface{}) error {
	return oauth2.NewInternalError("trying authroize with unauthenticated user", reasons...)
}

func newInvalidUserError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("invalid authorizing user", reasons...)
}

func newInvalidTenantForClientError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("authenticated client does not have access to the requested tenant", reasons...)
}

func newInvalidTenantForUserError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("authenticated user does not have access to the requested tenant", reasons...)
}

func newInvalidProviderError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("authenticated user does not have access to the requested provider", reasons...)
}
