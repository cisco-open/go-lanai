package internal

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"

// ExtendedClaims imlements oauth2.Claims. It's used only for access token decoding
type ExtendedClaims struct {
	oauth2.FieldClaimsMapper
	oauth2.BasicClaims
	oauth2.Claims
}

func NewExtendedClaims(claims ...oauth2.Claims) *ExtendedClaims {
	ptr := &ExtendedClaims{
		Claims: oauth2.MapClaims{},
	}
	for _, c := range claims {
		values := c.Values()
		for k, v := range values {
			ptr.Set(k, v)
		}
	}

	return ptr
}

func (c *ExtendedClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *ExtendedClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *ExtendedClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *ExtendedClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *ExtendedClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *ExtendedClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}

