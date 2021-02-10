package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

/*****************************
	legacyClaims Enhancer
 *****************************/
// legacyClaims imlements Claims and includes BasicClaims
type legacyClaims struct {
	oauth2.FieldClaimsMapper
	*oauth2.BasicClaims
	FirstName string `claim:"firstName"`
	LastName  string `claim:"lastName"`
	Email     string `claim:"email"`
	TenantId  string `claim:"tenantId"`
	Username  string `claim:"user_name"`
}

func (c *legacyClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *legacyClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *legacyClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *legacyClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *legacyClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *legacyClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}

// LegacyTokenEnhancer impelments order.Ordered and TokenEnhancer
type LegacyTokenEnhancer struct {

}

func (e *LegacyTokenEnhancer) Order() int {
	return TokenEnhancerOrderDetailsClaims
}

func (e *LegacyTokenEnhancer) Enhance(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	if t.Claims == nil {
		return nil, oauth2.NewInternalError("LegacyTokenEnhancer need to be placed immediately after BasicClaimsEnhancer")
	}

	basic, ok := t.Claims.(*oauth2.BasicClaims)
	if !ok {
		return nil, oauth2.NewInternalError("LegacyTokenEnhancer need to be placed immediately after BasicClaimsEnhancer")
	}

	legacy := &legacyClaims{
		BasicClaims: basic,
		Username: basic.Subject,
	}

	if ud, ok := oauth.Details().(security.UserDetails); ok {
		legacy.FirstName = ud.FirstName()
		legacy.LastName = ud.LastName()
		legacy.Email = ud.Email()
	}

	if td, ok := oauth.Details().(security.TenantDetails); ok {
		legacy.TenantId = td.TenantId()
	}

	t.Claims = legacy
	return t, nil
}




