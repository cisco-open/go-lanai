package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/google/uuid"
	"time"
)

var (
	endOfWorld = time.Date(2999, time.December, 31, 23, 59, 59, 0, time.UTC)
)

type AuthorizationService interface {
	CreateAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (oauth2.Authentication, error)
	SwitchAuthentication(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication, src oauth2.Authentication) (oauth2.Authentication, error)
	CreateAccessToken(ctx context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error)
	RefreshAccessToken(ctx context.Context, oauth oauth2.Authentication, refreshToken oauth2.RefreshToken) (oauth2.AccessToken, error)
}

/****************************
	Default implementation
 ****************************/

type DASOptions func(*DASOption)

type DASOption struct {
	DetailsFactory     *common.ContextDetailsFactory
	ClientStore        oauth2.OAuth2ClientStore
	AccountStore       security.AccountStore
	TenantStore        security.TenantStore
	ProviderStore      security.ProviderStore
	Issuer             security.Issuer
	TokenStore         TokenStore
	TokenEnhancers     []TokenEnhancer
	PostTokenEnhancers []TokenEnhancer
}

// DefaultAuthorizationService implements AuthorizationService
type DefaultAuthorizationService struct {
	detailsFactory    *common.ContextDetailsFactory
	clientStore       oauth2.OAuth2ClientStore
	accountStore      security.AccountStore
	tenantStore       security.TenantStore
	providerStore     security.ProviderStore
	tokenStore        TokenStore
	tokenEnhancer     TokenEnhancer
	postTokenEnhancer TokenEnhancer
}

func NewDefaultAuthorizationService(opts ...DASOptions) *DefaultAuthorizationService {
	basicEnhancer := BasicClaimsTokenEnhancer{}
	refreshTokenEnhancer := RefreshTokenEnhancer{}
	conf := DASOption{
		TokenEnhancers: []TokenEnhancer{
			&ExpiryTokenEnhancer{},
			&basicEnhancer,
			&LegacyTokenEnhancer{},
			&ResourceIdTokenEnhancer{},
			&DetailsTokenEnhancer{},
			&refreshTokenEnhancer,
		},
		PostTokenEnhancers: []TokenEnhancer{},
	}
	for _, opt := range opts {
		opt(&conf)
	}

	basicEnhancer.issuer = conf.Issuer
	refreshTokenEnhancer.issuer = conf.Issuer
	refreshTokenEnhancer.tokenStore = conf.TokenStore
	return &DefaultAuthorizationService{
		detailsFactory:    conf.DetailsFactory,
		clientStore:       conf.ClientStore,
		accountStore:      conf.AccountStore,
		tenantStore:       conf.TenantStore,
		providerStore:     conf.ProviderStore,
		tokenStore:        conf.TokenStore,
		tokenEnhancer:     NewCompositeTokenEnhancer(conf.TokenEnhancers...),
		postTokenEnhancer: NewCompositeTokenEnhancer(conf.PostTokenEnhancers...),
	}
}

func (s *DefaultAuthorizationService) CreateAuthentication(ctx context.Context,
	request oauth2.OAuth2Request, user security.Authentication) (oauth oauth2.Authentication, err error) {

	userAuth := ConvertToOAuthUserAuthentication(user)
	details, err := s.createContextDetails(ctx, request, userAuth, nil)
	if err != nil {
		return
	}

	// reconstruct user auth based on newly loaded facts (account may changed)
	if userAuth, err = s.createUserAuthentication(ctx, request, userAuth); err != nil {
		return
	}
	// create the result
	oauth = oauth2.NewAuthentication(func(conf *oauth2.AuthOption) {
		conf.Request = request
		conf.UserAuth = userAuth
		conf.Details = details
	})
	return
}

func (s *DefaultAuthorizationService) SwitchAuthentication(ctx context.Context,
	request oauth2.OAuth2Request, user security.Authentication,
	src oauth2.Authentication) (oauth oauth2.Authentication, err error) {

	userAuth := ConvertToOAuthUserAuthentication(user)
	details, err := s.createContextDetails(ctx, request, userAuth, src)
	if err != nil {
		return
	}

	// reconstruct user auth based on newly loaded facts (account may changed)
	if userAuth, err = s.createUserAuthentication(ctx, request, userAuth); err != nil {
		return
	}

	// create the result
	oauth = oauth2.NewAuthentication(func(conf *oauth2.AuthOption) {
		conf.Request = request
		conf.UserAuth = userAuth
		conf.Details = details
	})
	return
}

func (s *DefaultAuthorizationService) CreateAccessToken(c context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	token := s.reuseOrNewAccessToken(c, oauth)

	enhanced, e := s.tokenEnhancer.Enhance(c, token, oauth)
	if e != nil {
		return nil, e
	}

	// save token
	saved, e := s.tokenStore.SaveAccessToken(c, enhanced, oauth)
	if e != nil {
		return nil, e
	}
	return s.postTokenEnhancer.Enhance(c, saved, oauth)
}

func (s *DefaultAuthorizationService) RefreshAccessToken(c context.Context, oauth oauth2.Authentication, refreshToken oauth2.RefreshToken) (oauth2.AccessToken, error) {

	// we first remove existing access token associated with this refresh token
	// this functionality is necessary so refresh tokens can't be used to create an unlimited number of access tokens.
	_ = s.tokenStore.RemoveAccessToken(c, refreshToken)

	token := s.reuseOrNewAccessToken(c, oauth)
	token.SetRefreshToken(refreshToken)

	enhanced, e := s.tokenEnhancer.Enhance(c, token, oauth)
	if e != nil {
		return nil, e
	}

	// save token
	saved, e := s.tokenStore.SaveAccessToken(c, enhanced, oauth)
	if e != nil {
		return nil, e
	}

	return s.postTokenEnhancer.Enhance(c, saved, oauth)
}

/*
***************************

		Authorization Helpers
	 ***************************
*/
type authFacts struct {
	request  oauth2.OAuth2Request
	client   oauth2.OAuth2Client
	account  security.Account
	tenant   *security.Tenant
	provider *security.Provider
	source   oauth2.Authentication
	userAuth oauth2.UserAuthentication
}

func (s *DefaultAuthorizationService) createContextDetails(ctx context.Context,
	request oauth2.OAuth2Request, userAuth oauth2.UserAuthentication,
	src oauth2.Authentication) (security.ContextDetails, error) {
	now := time.Now().UTC()

	facts, e := s.loadAndVerifyFacts(ctx, request, userAuth)
	if e != nil {
		return nil, e
	}

	mutableCtx, ok := ctx.(utils.MutableContext)
	if !ok {
		return nil, newImmutableContextError()
	}

	mutableCtx.Set(oauth2.CtxKeyAuthenticatedClient, facts.client)
	mutableCtx.Set(oauth2.CtxKeyAuthenticatedAccount, facts.account)
	mutableCtx.Set(oauth2.CtxKeyAuthorizedTenant, facts.tenant)
	mutableCtx.Set(oauth2.CtxKeyAuthorizedProvider, facts.provider)
	mutableCtx.Set(oauth2.CtxKeyUserAuthentication, facts.userAuth)
	mutableCtx.Set(oauth2.CtxKeyAuthorizationIssueTime, now)
	if src != nil {
		facts.source = src
		mutableCtx.Set(oauth2.CtxKeySourceAuthentication, src)
	}

	// expiry
	expiry := s.determineExpiryTime(ctx, request, facts)
	if !expiry.IsZero() {
		mutableCtx.Set(oauth2.CtxKeyAuthorizationExpiryTime, expiry)
	}

	// auth time
	authTime := s.determineAuthenticationTime(ctx, userAuth, facts)
	if !authTime.IsZero() {
		mutableCtx.Set(oauth2.CtxKeyAuthenticationTime, authTime)
	}

	// create context details
	return s.detailsFactory.New(mutableCtx, request) //nolint:contextcheck // this is expected usage of MutableCtx
}

func (s *DefaultAuthorizationService) createUserAuthentication(ctx context.Context, _ oauth2.OAuth2Request, userAuth oauth2.UserAuthentication) (oauth2.UserAuthentication, error) {
	if userAuth == nil {
		return nil, nil
	}

	account, ok := ctx.Value(oauth2.CtxKeyAuthenticatedAccount).(security.Account)
	if !ok {
		return userAuth, nil
	}

	permissions := map[string]interface{}{}
	for _, v := range account.Permissions() {
		permissions[v] = true
	}

	details, ok := userAuth.Details().(map[string]interface{})
	if !ok || details == nil {
		details = map[string]interface{}{}
	}

	return oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
		opt.Principal = account.Username()
		opt.Permissions = permissions
		opt.State = userAuth.State()
		opt.Details = details
	}), nil
}

func (s *DefaultAuthorizationService) loadAndVerifyFacts(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (*authFacts, error) {
	client := RetrieveAuthenticatedClient(ctx)
	if client == nil {
		return nil, newInvalidClientError()
	}

	account, err := s.loadAccount(ctx, request, userAuth)
	if err != nil {
		return nil, err
	} else if account != nil && (account.Locked() || account.Disabled()) {
		return nil, newInvalidUserError("unsupported user's account locked or disabled")
	}

	defaultTenantId, assignedTenants, err := common.ResolveClientUserTenants(ctx, account, client)
	if err != nil {
		return nil, newInvalidTenantForUserError(fmt.Errorf("can't resolve account [%T] and client's [%T] tenants", account, client))
	}

	tenant, err := s.loadTenant(ctx, request, defaultTenantId)
	if err != nil {
		return nil, err
	}

	if err = s.verifyTenantAccess(ctx, tenant, assignedTenants); err != nil {
		return nil, err
	}

	provider, err := s.loadProvider(ctx, request, tenant)
	if err != nil {
		return nil, err
	}

	if account == nil { // at this point we have all the information we need if it's only client auth
		return &authFacts{
			request:  request,
			client:   client,
			tenant:   tenant,
			provider: provider,
		}, nil
	}

	if finalizer, ok := s.accountStore.(security.AccountFinalizer); ok {
		newAccount, err := finalizer.Finalize(ctx, account, security.FinalizeWithTenant(tenant))
		if err != nil {
			return nil, err
		}
		// Check that the ID and username have not been tampered with
		if newAccount.ID() != account.ID() || newAccount.Username() != account.Username() {
			return nil, newTamperedIDOrUsernameError()
		}
		// Check tenancy has not been tampered with
		if _, ok := newAccount.(security.AccountTenancy); !ok {
			return nil, newTamperedTenancyError()
		}
		if newAccount.(security.AccountTenancy).DefaultDesignatedTenantId() != account.(security.AccountTenancy).DefaultDesignatedTenantId() {
			return nil, newTamperedTenancyError()
		}
		if !utils.NewStringSet(newAccount.(security.AccountTenancy).DesignatedTenantIds()...).Equals(utils.NewStringSet(account.(security.AccountTenancy).DesignatedTenantIds()...)) {
			return nil, newTamperedTenancyError()
		}

		account = newAccount
	}

	// after account finalizer, we can re-create the userAuth security.Authentication,
	// and then return it from here
	// The Principal and State cannot change. Details and Permissions may change
	// can use something similar to auth.ConvertToOAuthUserAuthentication to grab things from, but then
	// edit the permissions
	// So keep everything from userAuth, we only need permissions from account
	// Check that the account userID and username did not change from the finalizer

	newUserAuth := ConvertToOAuthUserAuthentication(
		userAuth,
		ConvertWithSkipTypeCheck(true),
		func(option *ConvertOptions) {
			option.AppendUserAuthOptions(func(userAuth security.Authentication) oauth2.UserAuthOptions {
				return func(opt *oauth2.UserAuthOption) {
					opt.Permissions = userAuth.Permissions()
				}
			})
		})

	return &authFacts{
		request:  request,
		client:   client,
		account:  account,
		tenant:   tenant,
		provider: provider,
		userAuth: newUserAuth,
	}, nil
}

func (s *DefaultAuthorizationService) loadAccount(
	ctx context.Context,
	req oauth2.OAuth2Request,
	userAuth security.Authentication,
) (security.Account, error) {
	if userAuth == nil {
		return nil, nil
	}

	// sanity check, this should not happen
	if userAuth.State() < security.StateAuthenticated || userAuth.Principal() == nil {
		return nil, newUnauthenticatedUserError()
	}

	username, err := security.GetUsername(userAuth)
	if err != nil {
		return nil, newInvalidUserError(err)
	}

	acct, err := s.accountStore.LoadAccountByUsername(ctx, username)
	if err != nil {
		return nil, newInvalidUserError(err)
	}

	return acct, nil
}

func (s *DefaultAuthorizationService) loadTenant(
	ctx context.Context,
	request oauth2.OAuth2Request,
	defaultTenantId string,
) (*security.Tenant, error) {
	// extract tenant id or name
	tenantId, idOk := request.Parameters()[oauth2.ParameterTenantId]
	tenantExternalId, nOk := request.Parameters()[oauth2.ParameterTenantExternalId]
	if (!idOk || tenantId == "") && (!nOk || tenantExternalId == "") {
		tenantId = defaultTenantId
	}

	var tenant *security.Tenant
	var e error
	if tenantId != "" {
		tenant, e = s.tenantStore.LoadTenantById(ctx, tenantId)
		if e != nil {
			return nil, newInvalidTenantForUserError(fmt.Sprintf("error loading tenant with id [%s]", tenantId))
		}
	}

	if tenantExternalId != "" {
		tenant, e = s.tenantStore.LoadTenantByExternalId(ctx, tenantExternalId)
		if e != nil {
			return nil, newInvalidTenantForUserError(fmt.Sprintf("error loading tenant with externalId [%s]", tenantExternalId))
		}
	}

	return tenant, nil
}

func (s *DefaultAuthorizationService) verifyTenantAccess(ctx context.Context, tenant *security.Tenant, assignedTenantIds []string) error {
	if tenant == nil {
		return nil
	}

	tenantIds := utils.NewStringSet(assignedTenantIds...)

	if tenantIds.Has(security.SpecialTenantIdWildcard) {
		return nil
	}

	if !tenancy.AnyHasDescendant(ctx, tenantIds, tenant.Id) {
		return oauth2.NewInvalidGrantError("user does not have access to specified tenant")
	}

	return nil
}

func (s *DefaultAuthorizationService) loadProvider(ctx context.Context, _ oauth2.OAuth2Request, tenant *security.Tenant) (*security.Provider, error) {
	if tenant == nil {
		return nil, nil
	}

	providerId := tenant.ProviderId
	if providerId == "" {
		return nil, newInvalidProviderError("provider ID is not avalilable")
	}

	provider, e := s.providerStore.LoadProviderById(ctx, providerId)
	if e != nil {
		return nil, newInvalidProviderError(fmt.Sprintf("tenant [%s]'s provider is invalid", tenant.DisplayName))
	}
	return provider, nil
}

func (s *DefaultAuthorizationService) determineExpiryTime(ctx context.Context, _ oauth2.OAuth2Request, facts *authFacts) (expiry time.Time) {

	max := endOfWorld
	// When switching context, expiry should no later than original expiry time
	if facts.source != nil {
		if srcAuth, ok := facts.source.Details().(security.AuthenticationDetails); ok {
			max = srcAuth.ExpiryTime()
		}
	}

	if facts.client.AccessTokenValidity() == 0 {
		if max == endOfWorld {
			return
		} else {
			return max
		}
	}

	issueTime := ctx.Value(oauth2.CtxKeyAuthorizationIssueTime).(time.Time)
	expiry = issueTime.Add(facts.client.AccessTokenValidity()).UTC()
	return minTime(expiry, max)
}

func (s *DefaultAuthorizationService) determineAuthenticationTime(ctx context.Context, userAuth security.Authentication, facts *authFacts) (authTime time.Time) {
	if facts.source != nil {
		if srcAuth, ok := facts.source.Details().(security.AuthenticationDetails); ok {
			return srcAuth.AuthenticationTime()
		}
	}

	authTime = security.DetermineAuthenticationTime(ctx, userAuth)
	return
}

/*
***************************

		Helpers
	 ***************************
*/
func (s *DefaultAuthorizationService) reuseOrNewAccessToken(c context.Context, oauth oauth2.Authentication) *oauth2.DefaultAccessToken {
	existing, e := s.tokenStore.ReusableAccessToken(c, oauth)
	if e != nil || existing == nil {
		return oauth2.NewDefaultAccessToken(uuid.New().String())
	} else if t, ok := existing.(*oauth2.DefaultAccessToken); !ok {
		return oauth2.FromAccessToken(t)
	} else {
		return t
	}
}

func minTime(t1, t2 time.Time) time.Time {
	if t1.IsZero() || t1.Before(t2) {
		return t1
	} else {
		return t2
	}
}

//func ConvertToUserAuthenticationWithPermissions(
//	userAuth security.Authentication,
//	account security.Account,
//) oauth2.UserAuthentication {
//	principal, e := security.GetUsername(userAuth)
//	if e != nil {
//		principal = fmt.Sprintf("%v", userAuth)
//	}
//
//	details, ok := userAuth.Details().(map[string]interface{})
//	if !ok {
//		details = map[string]interface{}{
//			"Literal": userAuth.Details(),
//		}
//	}
//	permissions := make(map[string]interface{})
//	for _, permission := range account.Permissions() {
//		permissions[permission] = nil
//	}
//
//	return oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
//		opt.Principal = principal
//		opt.Permissions = permissions
//		opt.State = userAuth.State()
//		opt.Details = details
//	})
//}

/*
***************************

		Errors
	 ***************************
*/

func newTamperedIDOrUsernameError(reasons ...interface{}) error {
	return oauth2.NewInternalError("finalizer tampered with the ID or Username field", reasons...)
}

func newTamperedTenancyError(reasons ...interface{}) error {
	return oauth2.NewInternalError("finalizer tampered with the tenancy of the account", reasons...)
}

func newImmutableContextError(reasons ...interface{}) error {
	return oauth2.NewInternalError("context is not mutable", reasons...)
}

func newInvalidClientError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("trying authroize with unknown client", reasons...)
}

func newInvalidTenantForClientError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("authenticated client doesn't have access to the requested tenant", reasons...)
}

func newUnauthenticatedUserError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("trying authroize with unauthenticated user", reasons...)
}

func newInvalidUserError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("invalid authorizing user", reasons...)
}

func newInvalidTenantForUserError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("authenticated user does not have access to the requested tenant", reasons...)
}

func newInvalidProviderError(reasons ...interface{}) error {
	return oauth2.NewInvalidGrantError("authenticated user does not have access to the requested provider", reasons...)
}
