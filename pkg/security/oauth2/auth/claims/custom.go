package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

func UserId(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.UserId(), errorMissingDetails)
}

func AccountType(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AccountType().String(), errorMissingDetails)
}

func Currency(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.CurrencyCode(), errorMissingDetails)
}

func AssignedTenants(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AssignedTenantIds(), errorMissingDetails)
}

func TenantId(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.TenantDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.TenantId(), errorMissingDetails)
}

func TenantName(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.TenantDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.TenantName(), errorMissingDetails)
}

func TenantSuspended(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.TenantDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return utils.BoolPtr(details.TenantSuspended()), nil
}

func ProviderId(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderId(), errorMissingDetails)
}

func ProviderName(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.ProviderDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.ProviderName(), errorMissingDetails)
}

func Roles(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Roles(), errorMissingDetails)
}

func Permissions(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Permissions(), errorMissingDetails)
}

func OriginalUsername(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.ProxiedUserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	if details.Proxied() {
		return nonZeroOrError(details.OriginalUsername(), errorMissingDetails)
	} else {
		return nil, errorMissingDetails
	}
}