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

import "context"

type RequestedClaims interface {
	Get(claim string) (RequestedClaim, bool)
}

type RequestedClaim interface {
	Essential() bool
	Values() []string
	IsDefault() bool
}

type ClaimSpec interface {
	Calculate(ctx context.Context, opt *FactoryOption) (v interface{}, err error)
	Required(ctx context.Context, opt *FactoryOption) bool
}

type claimSpec struct {
	Func    ClaimFactoryFunc
	ReqFunc ClaimRequirementFunc
}

func (c claimSpec) Calculate(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if c.Func == nil {
		return nil, errorInvalidSpec
	}
	return c.Func(ctx, opt)
}

func (c claimSpec) Required(ctx context.Context, opt *FactoryOption) bool {
	if c.ReqFunc == nil {
		return false
	}
	return c.ReqFunc(ctx, opt)
}

func Required(fn ClaimFactoryFunc) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: requiredFunc,
	}
}

func Optional(fn ClaimFactoryFunc) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: optionalFunc,
	}
}

func RequiredIfParamsExists(fn ClaimFactoryFunc, requestParams ...string) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: func(ctx context.Context, opt *FactoryOption) bool {
			if opt.Source.OAuth2Request() == nil || opt.Source.OAuth2Request().Parameters() == nil {
				return false
			}
			req := opt.Source.OAuth2Request()
			for _, param := range requestParams {
				if _, ok := req.Parameters()[param]; ok {
					return true
				}
			}
			return false
		},
	}
}

func RequiredIfImplicitFlow(fn ClaimFactoryFunc) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: func(ctx context.Context, opt *FactoryOption) bool {
			if opt.Source.OAuth2Request() == nil || opt.Source.OAuth2Request().ResponseTypes() == nil {
				return false
			}
			return opt.Source.OAuth2Request().ResponseTypes().Has("token")
		},
	}
}

func Unsupported() ClaimSpec {
	return &claimSpec{
		Func: func(_ context.Context, _ *FactoryOption) (v interface{}, err error) {
			return nil, errorMissingDetails
		},
		ReqFunc: optionalFunc,
	}
}

func requiredFunc(_ context.Context, _ *FactoryOption) bool {
	return true
}

func optionalFunc(_ context.Context, _ *FactoryOption) bool {
	return false
}