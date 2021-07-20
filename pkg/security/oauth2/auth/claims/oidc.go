package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
)

// AddressClaim is defined at https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim
type AddressClaim struct {
	Formatted  string `json:"formatted,omitempty"`
	StreetAddr string `json:"street_address,omitempty"`
	City       string `json:"locality,omitempty"`
	Region     string `json:"region,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

func AuthenticationTime(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AuthenticationTime(), errorMissingDetails)
}

func FullName(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	name := strings.TrimSpace(strings.Join([]string{details.FirstName(), details.LastName()}, " "))
	return nonZeroOrError(name, errorMissingDetails)
}

func FirstName(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.FirstName(), errorMissingDetails)
}

func LastName(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LastName(), errorMissingDetails)
}

func Email(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Email(), errorMissingDetails)
}

func EmailVerified(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return utils.BoolPtr(strings.TrimSpace(details.Email()) != ""), nil
}

func ZoneInfo(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	// TODO maybe impelment this if possibile to extract it from locale
	return nil, errorMissingDetails
}

func Locale(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LocaleCode(), errorMissingDetails)
}

func Address(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	acct, ok := tryReloadAccount(ctx, opt).(security.AccountMetadata)
	if !ok || acct == nil {
		return nil, errorMissingDetails
	}
	addr := AddressClaim{
		Formatted:  acct.LocaleCode(),
		//StreetAddr: "",
		//City:       "",
		//Region:     "",
		//PostalCode: "",
		//Country:    "",
	}
	return &addr, nil
}

func Unsupported(_ context.Context, _ *FactoryOption) (v interface{}, err error) {
	return nil, errorMissingDetails
}
