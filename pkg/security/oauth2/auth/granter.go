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
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
)

type TokenGranter interface {
	// Grant create oauth2.AccessToken based on given TokenRequest
	// returns
	// 	- (nil, nil) if the TokenGranter doesn't support given request
	// 	- (non-nil, nil) if the TokenGranter support given request and created a token without error
	// 	- (nil, non-nil) if the TokenGranter support given request but rejected the request
	Grant(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error)
}

type TokenGranterOption struct {
	AuthService AuthorizationService
}

type CustomizableTokenGranter interface {
	Customize(options ...func(o *TokenGranterOption))
}

// CompositeTokenGranter implements TokenGranter
type CompositeTokenGranter struct {
	delegates []TokenGranter
}

func NewCompositeTokenGranter(delegates ...TokenGranter) *CompositeTokenGranter {
	return &CompositeTokenGranter{
		delegates: delegates,
	}
}

func (g *CompositeTokenGranter) Grant(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error) {
	for _, granter := range g.delegates {
		if token, e := granter.Grant(ctx, request); e != nil {
			return nil, e
		} else if token != nil {
			return token, nil
		}
	}
	return nil, oauth2.NewGranterNotAvailableError(fmt.Sprintf("grant type [%s] is not supported", request.GrantType))
}

func (g *CompositeTokenGranter) Add(granter TokenGranter) *CompositeTokenGranter {
	g.delegates = append(g.delegates, granter)
	return g
}

func (g *CompositeTokenGranter) Delegates() []TokenGranter {
	return g.delegates
}
