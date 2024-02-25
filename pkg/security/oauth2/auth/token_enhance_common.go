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
    "github.com/google/uuid"
    "time"
)

const (
	errTmplUnsupportedToken = `unsupported token implementation %T`
)

/*****************************
	Expiry Time Enhancer
 *****************************/

// ExpiryTokenEnhancer implements order.Ordered and TokenEnhancer
type ExpiryTokenEnhancer struct {}

func (e *ExpiryTokenEnhancer) Order() int {
	return TokenEnhancerOrderExpiry
}

func (e *ExpiryTokenEnhancer) Enhance(_ context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError(errTmplUnsupportedToken, t)
	}

	if authDetails, ok := oauth.Details().(security.AuthenticationDetails); ok {
		t.SetIssueTime(authDetails.IssueTime())
		t.SetExpireTime(authDetails.ExpiryTime())
	} else {
		t.SetIssueTime(time.Now().UTC())
	}
	return t, nil
}

/*****************************
	Details Enhancer
 *****************************/

// DetailsTokenEnhancer implements order.Ordered and TokenEnhancer
// it populate token's additional metadata other than claims, issue/expiry time
type DetailsTokenEnhancer struct {}

func (e *DetailsTokenEnhancer) Order() int {
	return TokenEnhancerOrderTokenDetails
}

func (e *DetailsTokenEnhancer) Enhance(_ context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError(errTmplUnsupportedToken, t)
	}

	t.SetScopes(oauth.OAuth2Request().Scopes())
	return t, nil
}

/*****************************
	BasicClaims Enhancer
 *****************************/

// BasicClaimsTokenEnhancer impelments order.Ordered and TokenEnhancer
type BasicClaimsTokenEnhancer struct {
	issuer security.Issuer
}

func (te *BasicClaimsTokenEnhancer) Order() int {
	return TokenEnhancerOrderBasicClaims
}

func (te *BasicClaimsTokenEnhancer) Enhance(_ context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError(errTmplUnsupportedToken, t)
	}

	request := oauth.OAuth2Request()
	basic := &oauth2.BasicClaims {
		Id:       uuid.New().String(),
		Audience: oauth2.StringSetClaim(utils.NewStringSet(request.ClientId())),
		Issuer:   te.issuer.Identifier(),
		ClientId: request.ClientId(),
		Scopes:   request.Scopes().Copy(),
	}

	if t.Claims() != nil && t.Claims().Has(oauth2.ClaimJwtId) {
		basic.Id = t.Claims().Get(oauth2.ClaimJwtId).(string)
	}

	if oauth.UserAuthentication() != nil {
		if sub, e := extractSubject(oauth.UserAuthentication()); e != nil {
			return nil, e
		} else {
			basic.Subject = sub
		}
	}

	if !t.ExpiryTime().IsZero() {
		basic.ExpiresAt = t.ExpiryTime()
	}

	if !t.IssueTime().IsZero() {
		basic.IssuedAt = t.IssueTime()
		basic.NotBefore = t.IssueTime()
	}

	t.SetClaims(basic)
	return t, nil
}

func extractSubject(auth security.Authentication) (string, error) {
	p := auth.Principal()
	switch p.(type) {
	case string:
		return p.(string), nil
	case security.Account:
		return p.(security.Account).Username(), nil
	case fmt.Stringer:
		return p.(fmt.Stringer).String(), nil
	default:
		return "", oauth2.NewInternalError("unable to extract subject for authentication %T", auth)
	}
}
