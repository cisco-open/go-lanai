package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
)

type AddressClaim struct {
	// TODO OIDC Address claim
}

func AuthenticationTime(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AuthenticationTime(), errorMissingDetails)
}

func FullName(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	name := strings.TrimSpace(strings.Join([]string{details.FirstName(), details.LastName()}, " "))
	return nonZeroOrError(name, errorMissingDetails)
}

func FirstName(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.FirstName(), errorMissingDetails)
}

func LastName(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LastName(), errorMissingDetails)
}

func Email(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Email(), errorMissingDetails)
}

func EmailVerified(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return utils.BoolPtr(strings.TrimSpace(details.Email()) != ""), nil
}

func ZoneInfo(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	// TODO maybe impelment this if possibile to extract it from locale
	return nil, errorMissingDetails
}

func Locale(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LocaleCode(), errorMissingDetails)
}

func Address(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LocaleCode(), errorMissingDetails)
}


