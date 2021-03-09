package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
)

var (
	errorMissingToken = errors.New("source authentication is missing token")
	errorMissingRequest = errors.New("source authentication is missing OAuth2 request")
	errorMissingUser = errors.New("source authentication is missing user")
	errorMissingDetails = errors.New("source authentication is missing required details")
	errorMissingClaims = errors.New("source authentication is missing required token claims")
)

type ClaimFactoryFunc func(ctx context.Context, opt *FactoryOption) (v interface{}, err error)

type ClaimSpec struct {
	Func ClaimFactoryFunc
	Req bool
}

type FactoryOptions func(opt *FactoryOption)

type FactoryOption struct {
	Source       oauth2.Authentication
	Issuer       security.Issuer
	AccountStore security.AccountStore
}

// options
func WithSource(oauth oauth2.Authentication) FactoryOptions {
	return func(opt *FactoryOption) {
		opt.Source = oauth
	}
}

func WithIssuer(issuer security.Issuer) FactoryOptions {
	return func(opt *FactoryOption) {
		opt.Issuer = issuer
	}
}

func WithAccountStore(accountStore security.AccountStore) FactoryOptions {
	return func(opt *FactoryOption) {
		opt.AccountStore = accountStore
	}
}

func Populate(ctx context.Context, claims oauth2.Claims, specs map[string]ClaimSpec, opts ...FactoryOptions) error {
	opt := FactoryOption{}
	for _, f := range opts {
		f(&opt)
	}
	for c, spec := range specs {
		if c == "" || spec.Func == nil {
			continue
		}
		v, e := spec.Func(ctx, &opt)
		if e != nil && spec.Req {
			return fmt.Errorf("unable to create claim [%s]: %v", c, e)
		} else if e != nil {
			continue
		}

		// check type and assign
		if e := safeSet(claims, c, v); e != nil {
			return e
		}
	}
	return nil
}

func safeSet(claims oauth2.Claims, claim string, value interface{}) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		if e, ok := r.(error); ok {
			err = fmt.Errorf("unable to create claim [%s]: %v", claim, e)
		} else {
			err = fmt.Errorf("unable to create claim [%s]: %v", claim, r)
		}
	}()

	claims.Set(claim, value)
	return nil
}

/*************************
	helpers
 *************************/
func tryReloadAccount(ctx context.Context, opt *FactoryOption) security.Account {
	if acct, ok := ctx.Value(oauth2.CtxKeyAuthenticatedAccount).(security.Account); ok {
		return acct
	}

	if opt.AccountStore == nil {
		return nil
	}

	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil
	}

	user, e := opt.AccountStore.LoadAccountById(ctx, details.UserId())
	if e != nil {
		return nil
	}

	// cache it in context if possible
	if mc, ok := ctx.(utils.MutableContext); ok {
		mc.Set(oauth2.CtxKeyAuthenticatedAccount, user)
	}
	return user
}


