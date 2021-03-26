package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/google/uuid"
)

var (
	refreshTokenAllowedGrants = utils.NewStringSet(
		oauth2.GrantTypeAuthCode,
		oauth2.GrantTypeImplicit,
		oauth2.GrantTypeRefresh,
		oauth2.GrantTypePassword, // this is for dev purpose, shouldn't be allowed
	)
)

/*****************************
	RefreshToken Enhancer
 *****************************/
// RefreshTokenEnhancer impelments order.Ordered and TokenEnhancer
// RefreshTokenEnhancer is responsible to create refresh token and associate it with the given access token
type RefreshTokenEnhancer struct {
	tokenStore TokenStore
	issuer     security.Issuer
}

func (te *RefreshTokenEnhancer) Order() int {
	return TokenEnhancerOrderRefreshToken
}

func (te *RefreshTokenEnhancer) Enhance(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	// step 1, check if refresh token is needed
	client, ok := ctx.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client)
	if !ok || !te.isRefreshTokenNeeded(ctx, token, oauth, client) {
		return token, nil
	}

	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	// step 2, create refresh token
	// Note: we don't reuse refresh token
	id := uuid.New().String()
	refresh := oauth2.NewDefaultRefreshToken(id)

	// step 3, set expriy time
	// Note: refresh token's validity is counted since authentication time
	details, ok := oauth.Details().(security.AuthenticationDetails)
	if ok && client.RereshTokenValidity() > 0 && !details.AuthenticationTime().IsZero() {
		expiry := details.AuthenticationTime().Add(client.RereshTokenValidity())
		refresh.SetExpireTime(expiry)
	}

	// step 4 create claims,
	request := oauth.OAuth2Request()
	claims := oauth2.BasicClaims{
		Id:       id,
		Audience: utils.NewStringSet(client.ClientId()),
		Issuer: te.issuer.Identifier(),
		Scopes: request.Scopes(),
	}

	if oauth.UserAuthentication() != nil {
		if sub, e := extractSubject(oauth.UserAuthentication()); e != nil {
			return nil, e
		} else {
			claims.Subject = sub
		}
	}

	if refresh.WillExpire() && !refresh.ExpiryTime().IsZero() {
		claims.Set(oauth2.ClaimExpire, refresh.ExpiryTime())
	}
	refresh.SetClaims(&claims)

	// step 5, save refresh token
	if saved, e := te.tokenStore.SaveRefreshToken(ctx, refresh, oauth); e == nil {
		t.SetRefreshToken(saved)
	}
	return t, nil
}

/*****************************
	Helpers
 *****************************/
func (te *RefreshTokenEnhancer) isRefreshTokenNeeded(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication, client oauth2.OAuth2Client) bool {
	// refresh grant should be allowed for the client
	if e := ValidateGrant(ctx, client, oauth2.GrantTypeRefresh); e != nil {
		return false
	}

	// only some grant types can return refresh token
	if !refreshTokenAllowedGrants.Has(oauth.OAuth2Request().GrantType()) {
		return false
	}

	// last, if given token already have an refresh token, no need to generate new
	return token.RefreshToken() == nil || token.RefreshToken().WillExpire() && token.RefreshToken().Expired()
}

func copyClaim(dest oauth2.Claims, src oauth2.Claims, claim string) {
	if src != nil && src.Has(claim) {
		dest.Set(claim, src.Get(claim))
	}
}
