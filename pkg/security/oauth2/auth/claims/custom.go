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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

func UserId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.UserId(), errorMissingDetails)
}

func AccountType(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AccountType().String(), errorMissingDetails)
}

func Currency(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.CurrencyCode(), errorMissingDetails)
}

func DefaultTenantId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	acct := tryReloadAccount(ctx, opt)
	tenancy, ok := acct.(security.AccountTenancy)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(tenancy.DefaultDesignatedTenantId(), errorMissingDetails)
}

func TenantId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.TenantDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.TenantId(), errorMissingDetails)
}

func TenantExternalId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.TenantDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.TenantExternalId(), errorMissingDetails)
}

func TenantSuspended(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.TenantDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return utils.BoolPtr(details.TenantSuspended()), nil
}

func ProviderId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderId(), errorMissingDetails)
}

func ProviderName(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderName(), errorMissingDetails)
}

func ProviderDisplayName(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderDisplayName(), errorMissingDetails)
}

func ProviderDescription(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderDescription(), errorMissingDetails)
}

func ProviderEmail(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderEmail(), errorMissingDetails)
}

func ProviderNotificationType(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderNotificationType(), errorMissingDetails)
}

func Roles(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Roles(), errorMissingDetails)
}

func Permissions(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Permissions(), errorMissingDetails)
}

func OriginalUsername(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.ProxiedUserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	if details.Proxied() {
		return nonZeroOrError(details.OriginalUsername(), errorMissingDetails)
	} else {
		return nil, errorMissingDetails
	}
}
