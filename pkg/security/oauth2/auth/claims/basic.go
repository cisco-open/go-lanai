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

package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"reflect"
	"time"
)

func ClientId(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	return nonZeroOrError(opt.Source.OAuth2Request().ClientId(), errorMissingDetails)
}

func Audience(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	if opt.Source.OAuth2Request().ClientId() == "" {
		return nil, errorMissingDetails
	}
	return utils.NewStringSet(opt.Source.OAuth2Request().ClientId()), nil
}

func JwtId(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	return extractAccessTokenClaim(opt, oauth2.ClaimJwtId)
}

func ExpiresAt(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.AccessToken() != nil {
		v = opt.Source.AccessToken().ExpiryTime()
	}

	if details, ok := opt.Source.Details().(security.ContextDetails); ok {
		v = details.ExpiryTime()
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func IssuedAt(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.AccessToken() != nil {
		v = opt.Source.AccessToken().IssueTime()
	}

	if details, ok := opt.Source.Details().(security.ContextDetails); ok {
		v = details.IssueTime()
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func Issuer(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Issuer != nil {
		if id := opt.Issuer.Identifier(); id != "" {
			return id, nil
		}
	}

	// fall back to extract from access token
	return extractAccessTokenClaim(opt, oauth2.ClaimIssuer)
}

func NotBefore(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	return extractAccessTokenClaim(opt, oauth2.ClaimNotBefore)
}

func Subject(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	return Username(ctx, opt)
}

func Scopes(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	return nonZeroOrError(opt.Source.OAuth2Request().Scopes(), errorMissingDetails)
}

func Username(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.UserAuthentication() == nil || opt.Source.UserAuthentication().Principal() == nil {
		return nil, errorMissingUser
	}
	username, e := security.GetUsername(opt.Source.UserAuthentication())
	if e != nil {
		return nil, errorMissingUser
	}
	return nonZeroOrError(username, errorMissingDetails)
}

func nonZeroOrError(v interface{}, candidateError error) (interface{}, error) {
	var isZero bool
	switch v.(type) {
	case string:
		isZero = v.(string) == ""
	case time.Time:
		isZero = v.(time.Time).IsZero()
	case utils.StringSet:
		isZero = len(v.(utils.StringSet)) == 0
	default:
		isZero = reflect.ValueOf(v).IsZero()
	}

	if isZero {
		return nil, candidateError
	}
	return v, nil
}

func extractAccessToken(opt *FactoryOption) oauth2.AccessToken {
	token := opt.AccessToken
	if token == nil {
		token = opt.Source.AccessToken()
	}
	return token
}

func extractAccessTokenClaim(opt *FactoryOption, claim string) (v interface{}, err error) {
	container, ok := extractAccessToken(opt).(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	claims := container.Claims()
	if !claims.Has(claim) {
		return nil, errorMissingClaims
	}
	return claims.Get(claim), nil
}