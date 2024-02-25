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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/utils"
)

/***********************
	Common Functions
 ***********************/

func RetrieveAuthenticatedClient(c context.Context) oauth2.OAuth2Client {
	if client, ok := c.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client); ok {
		return client
	}

	sec := security.Get(c)
	if sec.State() < security.StatePrincipalKnown {
		return nil
	}

	if client, ok := sec.Principal().(oauth2.OAuth2Client); ok {
		return client
	}
	return nil
}

func RetrieveFullyAuthenticatedClient(c context.Context) (oauth2.OAuth2Client, error) {
	sec := security.Get(c)
	if sec.State() < security.StateAuthenticated {
		return nil, oauth2.NewInvalidGrantError("client is not fully authenticated")
	}

	if client, ok := sec.Principal().(oauth2.OAuth2Client); ok {
		return client, nil
	}
	return nil, oauth2.NewInvalidGrantError("client is not fully authenticated")
}

func ValidateResponseTypes(ctx context.Context, request *AuthorizeRequest, supported utils.StringSet) error {
	if request.ResponseTypes == nil {
		return oauth2.NewInvalidAuthorizeRequestError("response_type is required")
	}

	// shortcut if already validated
	if v := request.Context().Value(ctxKeyValidResponseType); v != nil {
		return nil
	}

	if ok, invalid := IsSubSet(ctx, supported, request.ResponseTypes); !ok {
		return oauth2.NewInvalidResponseTypeError(fmt.Sprintf("unsupported response type: %s", invalid))
	}

	// mark validated
	request.Context().Set(ctxKeyValidResponseType, true)
	return nil
}

func ValidateGrant(_ context.Context, client oauth2.OAuth2Client, grantType string) error {
	if grantType == "" {
		return oauth2.NewInvalidTokenRequestError("missing grant_type")
	}

	if !client.GrantTypes().Has(grantType) {
		return oauth2.NewUnauthorizedClientError(fmt.Sprintf("grant type '%s' is not allowed by this client '%s'", grantType, client.ClientId()))
	}

	return nil
}

func ValidateScope(c context.Context, client oauth2.OAuth2Client, scopes ...string) error {
	for _, scope := range scopes {
		if !client.Scopes().Has(scope) {
			return oauth2.NewInvalidScopeError("invalid scope: " + scope)
		}
	}
	return nil
}

func ValidateAllScopes(c context.Context, client oauth2.OAuth2Client, scopes utils.StringSet) error {
	if ok, invalid := IsSubSet(c, client.Scopes(), scopes); !ok {
		return oauth2.NewInvalidScopeError("invalid scope: " + invalid)
	}
	return nil
}

func ValidateAllAutoApprovalScopes(c context.Context, client oauth2.OAuth2Client, scopes utils.StringSet) error {
	if ok, invalid := IsSubSet(c, client.AutoApproveScopes(), scopes); !ok {
		return oauth2.NewAccessRejectedError("scope not auto approved: " + invalid)
	}
	return nil
}

func IsSubSet(_ context.Context, superset utils.StringSet, subset utils.StringSet) (ok bool, invalid string) {
	for scope, _ := range subset {
		if !superset.Has(scope) {
			return false, scope
		}
	}
	return true, ""
}

// ValidateApproval approval param is a map with scope as keys and approval status as values
func ValidateApproval(c context.Context, approval map[string]bool, client oauth2.OAuth2Client, scopes utils.StringSet) error {
	if e := ValidateAllScopes(c, client, scopes); e != nil {
		return e
	}

	for scope, _ := range scopes {
		if approved, ok := approval[scope]; !ok || !approved {
			return oauth2.NewAccessRejectedError(fmt.Sprintf("user disapproved scope [%s]", scope))
		}
	}
	return nil
}

func LoadAndValidateClientId(c context.Context, clientId string, clientStore oauth2.OAuth2ClientStore) (oauth2.OAuth2Client, error) {
	if clientId == "" {
		return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Sprintf("A client id must be provided"))
	}

	client, e := clientStore.LoadClientByClientId(c, clientId)
	if e != nil {
		return nil, oauth2.NewClientNotFoundError("invalid client")
	}
	return client, nil
}

func ResolveRedirectUri(_ context.Context, redirectUri string, client oauth2.OAuth2Client) (string, error) {
	if client.RedirectUris() == nil || len(client.RedirectUris()) == 0 {
		return "", oauth2.NewInvalidAuthorizeRequestError(
			"at least one redirectUri must be registered in the client")
	}

	// The resolved redirect URI is either the redirect_uri from the parameters or the one from
	// clientDetails.
	if redirectUri == "" && len(client.RedirectUris()) == 1 {
		// single registered redirect URI
		return client.RedirectUris().Values()[0], nil
	} else if redirectUri == "" {
		return "", oauth2.NewInvalidRedirectUriError("the redirect_uri must be proveded because the client have multiple registered redirect URI")
	}

	for registered, _ := range client.RedirectUris() {
		matcher, e := NewWildcardUrlMatcher(registered)
		if e != nil {
			continue
		}
		if matches, e := matcher.Matches(redirectUri); e == nil && matches {
			return redirectUri, nil
		}
	}

	return "", oauth2.NewInvalidRedirectUriError("the redirect_uri must be registered with the client")
}

type ConvertOptions struct {
	SkipTypeCheck   bool
	userAuthOptions []OverrideAuthOptions
}

func (c *ConvertOptions) AppendUserAuthOptions(option OverrideAuthOptions) {
	c.userAuthOptions = append(c.userAuthOptions, option)
}

type ConvertOption func(option *ConvertOptions)

func ConvertWithSkipTypeCheck(skipTypeCheck bool) ConvertOption {
	return func(option *ConvertOptions) {
		option.SkipTypeCheck = skipTypeCheck
	}
}

// OverrideAuthOptions allows the oauth2.UserAuthOptions to be overridden during the
// conversion when creating and returning a new user authentication.
type OverrideAuthOptions func(userAuth security.Authentication) oauth2.UserAuthOptions

// ConvertToOAuthUserAuthentication takes any type of authentication and convert it into oauth2.Authentication
func ConvertToOAuthUserAuthentication(userAuth security.Authentication, options ...ConvertOption) oauth2.UserAuthentication {
	var opts ConvertOptions
	for _, opt := range options {
		opt(&opts)
	}
	if !opts.SkipTypeCheck {
		switch ua := userAuth.(type) {
		case nil:
			return nil
		case oauth2.UserAuthentication:
			return ua
		}
	}

	principal, e := security.GetUsername(userAuth)
	if e != nil {
		principal = fmt.Sprintf("%v", userAuth)
	}

	details, ok := userAuth.Details().(map[string]interface{})
	if !ok {
		details = map[string]interface{}{
			"Literal": userAuth.Details(),
		}
	}

	defaultOption := func(opt *oauth2.UserAuthOption) {
		opt.Principal = principal
		opt.Permissions = userAuth.Permissions()
		opt.State = userAuth.State()
		opt.Details = details
	}

	var wrappedAuthOptions []oauth2.UserAuthOptions
	for _, opt := range opts.userAuthOptions {
		wrappedOption := opt(userAuth)
		wrappedAuthOptions = append(wrappedAuthOptions, wrappedOption)
	}

	var authenticationOptions []oauth2.UserAuthOptions
	authenticationOptions = append(authenticationOptions, defaultOption)
	authenticationOptions = append(authenticationOptions, wrappedAuthOptions...)
	return oauth2.NewUserAuthentication(authenticationOptions...)
}
