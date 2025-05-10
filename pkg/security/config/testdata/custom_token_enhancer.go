package testdata

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
)

type CustomClaims struct {
	oauth2.FieldClaimsMapper
	oauth2.Claims
	MyClaim string `claim:"MyClaim"`
}

func (c *CustomClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *CustomClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *CustomClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *CustomClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *CustomClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *CustomClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}

type CustomTokenEnhancer struct{}

func NewCustomTokenEnhancer() auth.TokenEnhancer {
	return &CustomTokenEnhancer{}
}

func (c *CustomTokenEnhancer) Enhance(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	if t.Claims() == nil {
		return nil, oauth2.NewInternalError("need to be placed after BasicClaimsEnhancer")
	}

	customClaims := &CustomClaims{
		Claims: t.Claims(),
	}
	customClaims.MyClaim = "my_claim_value"

	t.SetClaims(customClaims)
	return t, nil
}
