package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
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

type ClaimFactoryFunc func(ctx context.Context, auth oauth2.Authentication) (v interface{}, err error)

type ClaimSpec struct {
	Func ClaimFactoryFunc
	Req bool
}

func Populate(ctx context.Context, claims oauth2.Claims, specs map[string]ClaimSpec, src oauth2.Authentication) error {
	for c, spec := range specs {
		if c == "" || spec.Func == nil {
			continue
		}
		v, e := spec.Func(ctx, src)
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
	Factory Functions
 *************************/


