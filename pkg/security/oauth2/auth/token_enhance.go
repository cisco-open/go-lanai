package auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"time"
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
	// resort the delegates
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

/*****************************
	BasicClaims Enhancer
 *****************************/
// ExpiryTokenEnhancer impelments order.Ordered and TokenEnhancer
type ExpiryTokenEnhancer struct {

}

func (e *ExpiryTokenEnhancer) Order() int {
	return TokenEnhancerOrderExpiry
}

func (e *ExpiryTokenEnhancer) Enhance(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	// we assume client is available at this point
	client := RetrieveAuthenticatedClient(c)

	// TODO check extensions
	t.SetIssueTime(time.Now())
	expire := t.IssueTime().Add(client.AccessTokenValidity())
	t.SetExpireTime(expire)
	return t, nil
}


/*****************************
	BasicClaims Enhancer
 *****************************/
// BasicClaimsTokenEnhancer impelments order.Ordered and TokenEnhancer
type BasicClaimsTokenEnhancer struct {

}

func (e *BasicClaimsTokenEnhancer) Order() int {
	return TokenEnhancerOrderBasicClaims
}

func (e *BasicClaimsTokenEnhancer) Enhance(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	request := oauth.OAuth2Request()
	basic := &oauth2.BasicClaims {
		Id:       uuid.New().String(),
		Audience: request.ClientId(),
		Issuer:   "localhost:8080", // TODO Issuer should be extracted for configuration
	}

	if t.Claims != nil {
		basic.Id = t.Claims.Get(oauth2.ClaimJwtId).(string)
	}

	if oauth.UserAuthentication() != nil {
		if sub, e := extractSubject(oauth.UserAuthentication()); e != nil {
			return nil, e
		} else {
			basic.Subject = sub
		}
	}

	if !t.ExpiryTime().IsZero() {
		basic.ExpiresAt = t.ExpiryTime()
	}

	if !t.IssueTime().IsZero() {
		basic.IssuedAt = t.IssueTime()
		basic.NotBefore = t.IssueTime()
	}

	t.Claims = basic
	return t, nil
}

func extractSubject(auth security.Authentication) (string, error) {
	p := auth.Principal()
	switch p.(type) {
	case string:
		return p.(string), nil
	case security.Account:
		return p.(security.Account).Username(), nil
	case fmt.Stringer:
		return p.(fmt.Stringer).String(), nil
	default:
		return "", oauth2.NewInternalError("unable to extract subject for authentication %T", auth)
	}
}


