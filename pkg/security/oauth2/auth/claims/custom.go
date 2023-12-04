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
