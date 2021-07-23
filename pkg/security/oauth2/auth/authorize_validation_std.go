package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"strings"
)

var (
	redirectGrantTypes = []string{oauth2.GrantTypeAuthCode, oauth2.GrantTypeImplicit}
)

// StandardAuthorizeRequestProcessor implements AuthorizeRequestProcessor and order.Ordered
// it validate auth request against standard oauth2 specs
type StandardAuthorizeRequestProcessor struct {
	responseTypes   utils.StringSet
	clientStore     oauth2.OAuth2ClientStore
}

type StdARPOptions func(*StdARPOption)

type StdARPOption struct {
	ResponseTypes   utils.StringSet
	ClientStore     oauth2.OAuth2ClientStore
}

func NewStandardAuthorizeRequestProcessor(opts...StdARPOptions) *StandardAuthorizeRequestProcessor {
	opt := StdARPOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &StandardAuthorizeRequestProcessor{
		responseTypes: opt.ResponseTypes,
		clientStore: opt.ClientStore,
	}
}

func (p *StandardAuthorizeRequestProcessor) Process(ctx context.Context, request *AuthorizeRequest) (validated *AuthorizeRequest, err error) {
	if e := p.validateResponseTypes(ctx, request); e != nil {
		return nil, e
	}

	client, e := p.validateClientId(ctx, request)
	if  e != nil {
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

	// TODO check client tenant restrictions
	_ = client.TenantRestrictions()

	return request, nil
}

func (p *StandardAuthorizeRequestProcessor) validateResponseTypes(_ context.Context, request *AuthorizeRequest) error {
	if request.ResponseTypes == nil {
		return oauth2.NewInvalidAuthorizeRequestError("response_type is required")
	}

	for k, _ := range request.ResponseTypes {
		if !p.responseTypes.Has(strings.ToLower(k)) {
			return oauth2.NewInvalidResponseTypeError(fmt.Sprintf("unsupported response type: %s", k))
		}
	}

	return nil
}

func (p *StandardAuthorizeRequestProcessor) validateClientId(c context.Context, request *AuthorizeRequest) (oauth2.OAuth2Client, error) {
	return LoadAndValidateClientId(c, request.ClientId, p.clientStore)
}

func (p *StandardAuthorizeRequestProcessor) validateRedirectUri(c context.Context, request *AuthorizeRequest, client oauth2.OAuth2Client) error {
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
	redirect, e := ResolveRedirectUri(c, request.RedirectUri, client)
	if e != nil {
		return e
	}
	request.RedirectUri = redirect
	return nil
}

func (p *StandardAuthorizeRequestProcessor) validateScope(c context.Context, request *AuthorizeRequest, client oauth2.OAuth2Client) error {
	if request.Scopes == nil || len(request.Scopes) == 0 {
		request.Scopes = client.Scopes().Copy()
	} else if e := ValidateAllScopes(c, client, request.Scopes); e != nil {
		return e
	}
	return nil
}
