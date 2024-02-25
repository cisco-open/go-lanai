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
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
)

/*****************************
	Abstraction
 *****************************/

// TokenEnhancer modify given oauth2.AccessToken or return a new token based on given context and auth
// Most TokenEnhancer responsible to add/modify claims of given access token
// But it's not limited to do so. e.g. TokenEnhancer could be responsible to  install refresh token
// Usually if given token is not mutable, the returned token would be different instance
type TokenEnhancer interface {
	Enhance(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error)
}

/*****************************
	Common Implementations
 *****************************/

type CompositeTokenEnhancer struct {
	delegates []TokenEnhancer
}

func NewCompositeTokenEnhancer(delegates ...TokenEnhancer) *CompositeTokenEnhancer {
	return &CompositeTokenEnhancer{delegates: delegates}
}

func (e *CompositeTokenEnhancer) Enhance(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	for _, enhancer := range e.delegates {
		current, err := enhancer.Enhance(ctx, token, oauth)
		if err != nil {
			return nil, err
		}
		token = current
	}
	return token, nil
}

func (e *CompositeTokenEnhancer) Add(enhancers... TokenEnhancer) {
	e.delegates = append(e.delegates, flattenEnhancers(enhancers)...)
	// resort the extensions
	order.SortStable(e.delegates, order.OrderedFirstCompare)
}

func (e *CompositeTokenEnhancer) Remove(enhancer TokenEnhancer) {
	for i, item := range e.delegates {
		if item != enhancer {
			continue
		}

		// remove but keep order
		if i + 1 <= len(e.delegates) {
			copy(e.delegates[i:], e.delegates[i+1:])
		}
		e.delegates = e.delegates[:len(e.delegates) - 1]
		return
	}
}

// flattenEnhancers recursively flatten any nested CompositeTokenEnhancer
func flattenEnhancers(enhancers []TokenEnhancer) (ret []TokenEnhancer) {
	ret = make([]TokenEnhancer, 0, len(enhancers))
	for _, e := range enhancers {
		switch e.(type) {
		case *CompositeTokenEnhancer:
			flattened := flattenEnhancers(e.(*CompositeTokenEnhancer).delegates)
			ret = append(ret, flattened...)
		default:
			ret = append(ret, e)
		}
	}
	return
}
