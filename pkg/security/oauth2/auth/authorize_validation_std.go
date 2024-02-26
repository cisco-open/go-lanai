// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
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
