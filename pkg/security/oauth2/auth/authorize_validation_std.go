package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

var (
	redirectGrantTypes     = []string{oauth2.GrantTypeAuthCode, oauth2.GrantTypeImplicit}
	supportedResponseTypes = utils.NewStringSet("token", "code")
)

// StandardAuthorizeRequestProcessor implements ChainedAuthorizeRequestProcessor and order.Ordered
// it validate auth request against standard oauth2 specs
type StandardAuthorizeRequestProcessor struct {
	clientStore  oauth2.OAuth2ClientStore
	accountStore security.AccountStore
}

type StdARPOptions func(*StdARPOption)

type StdARPOption struct {
	ClientStore  oauth2.OAuth2ClientStore
	AccountStore security.AccountStore
}

func NewStandardAuthorizeRequestProcessor(opts ...StdARPOptions) *StandardAuthorizeRequestProcessor {
	opt := StdARPOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &StandardAuthorizeRequestProcessor{
		clientStore:  opt.ClientStore,
		accountStore: opt.AccountStore,
	}
}

func (p *StandardAuthorizeRequestProcessor) Process(ctx context.Context, request *AuthorizeRequest, chain AuthorizeRequestProcessChain) (validated *AuthorizeRequest, err error) {
	if e := p.validateResponseTypes(ctx, request); e != nil {
		return nil, e
	}

	client, e := p.validateClientId(ctx, request)
	if e != nil {
		return nil, e
	}
	request.Context().Set(oauth2.CtxKeyAuthenticatedClient, client)

	if e := p.validateRedirectUri(ctx, request, client); e != nil {
		return nil, e
	}

	// starting from this point, we know that redirect uri can be used
	request.Context().Set(oauth2.CtxKeyResolvedAuthorizeRedirect, request.RedirectUri)
	if request.State != "" {
		request.Context().Set(oauth2.CtxKeyResolvedAuthorizeState, request.State)
	}

	if e := p.validateScope(ctx, request, client); e != nil {
		return nil, e
	}

	if e := p.validateClientTenancy(ctx, client); e != nil {
		return nil, e
	}

	return chain.Next(ctx, request)
}

func (p *StandardAuthorizeRequestProcessor) validateResponseTypes(ctx context.Context, request *AuthorizeRequest) error {
	return ValidateResponseTypes(ctx, request, supportedResponseTypes)
}

func (p *StandardAuthorizeRequestProcessor) validateClientId(ctx context.Context, request *AuthorizeRequest) (oauth2.OAuth2Client, error) {
	return LoadAndValidateClientId(ctx, request.ClientId, p.clientStore)
}

func (p *StandardAuthorizeRequestProcessor) validateRedirectUri(ctx context.Context, request *AuthorizeRequest, client oauth2.OAuth2Client) error {
	// first, we check for client's grant type to see if redirect URI is allowed
	if client.GrantTypes() == nil || len(client.GrantTypes()) == 0 {
		return oauth2.NewInvalidAuthorizeRequestError("client must have at least one authorized grant type")
	}

	found := false
	for _, grant := range redirectGrantTypes {
		found = found || client.GrantTypes().Has(grant)
	}
	if !found {
		return oauth2.NewInvalidAuthorizeRequestError(
			"redirect_uri can only be used by implicit or authorization_code grant types")
	}

	// Resolve redirect URI
	// The resolved redirect URI is either the redirect_uri from the parameters or the one from
	// clientDetails. Either way we need to store it on the AuthorizationRequest.
	redirect, e := ResolveRedirectUri(ctx, request.RedirectUri, client)
	if e != nil {
		return e
	}
	request.RedirectUri = redirect
	return nil
}

func (p *StandardAuthorizeRequestProcessor) validateScope(ctx context.Context, request *AuthorizeRequest, client oauth2.OAuth2Client) error {
	if request.Scopes == nil || len(request.Scopes) == 0 {
		request.Scopes = client.Scopes().Copy()
	} else if e := ValidateAllScopes(ctx, client, request.Scopes); e != nil {
		return e
	}
	return nil
}

func (p *StandardAuthorizeRequestProcessor) validateClientTenancy(ctx context.Context, client oauth2.OAuth2Client) error {
	userAuth := security.Get(ctx)
	// Note if current security doesn't have valid username, we don't return error here. We let access handler to deal with it
	username, e := security.GetUsername(userAuth)
	if !security.IsFullyAuthenticated(userAuth) || e != nil {
		return nil //nolint:nilerr // intended behaviour
	}

	acct, e := p.accountStore.LoadAccountByUsername(ctx, username)
	if e != nil || acct == nil {
		security.Clear(ctx)
		return security.NewUsernameNotFoundError("cannot retrieve account from current session")
	}
	acct, e = WrapAccount(ctx, acct, client)
	if e != nil || acct == nil {
		security.Clear(ctx)
		return security.NewUsernameNotFoundError("cannot resolve user and client tenancy")
	}

	if len(acct.(security.AccountTenancy).DesignatedTenantIds()) == 0 && !client.Scopes().Has(oauth2.ScopeCrossTenant) {
		security.Clear(ctx)
		return security.NewUsernameNotFoundError("user has no access to tenants of this client")
	}
	return nil
}
