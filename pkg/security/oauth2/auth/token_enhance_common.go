package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/google/uuid"
	"time"
)

/*****************************
	Expiry Time Enhancer
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

	if authDetails, ok := oauth.Details().(security.AuthenticationDetails); ok {
		t.SetIssueTime(authDetails.IssueTime())
		t.SetExpireTime(authDetails.ExpiryTime())
	} else {
		t.SetIssueTime(time.Now())
	}
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
		ClientId: request.ClientId(),
		Scopes:   request.Scopes().Copy(),
	}

	if t.Claims != nil {
		basic.Id = t.Claims.Get(oauth2.ClaimJwtId).(string)
		basic.Scopes = utils.NewStringSetFrom(t.Claims.Get(oauth2.ClaimScope))
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
