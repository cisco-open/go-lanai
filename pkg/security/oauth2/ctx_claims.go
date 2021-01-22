package oauth2

import (
	"encoding/json"
	"time"
)

const (
	ClaimTag = "claim"
)

type Claims interface {
	Get(claim string) interface{}
	Has(claim string) bool
	Set(claim string, value interface{})
}

/*********************
	Implements
 *********************/
// MapClaims imlements Claims
type MapClaims map[string]interface{}

func (c MapClaims) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(c))
}

func (c MapClaims) UnmarshalJSON(bytes []byte) error {
	m := map[string]interface{}(c)
	return json.Unmarshal(bytes, &m)
}

func (c MapClaims) Get(claim string) interface{} {
	return c[claim]
}

func (c MapClaims) Has(claim string) bool {
	_,ok := c[claim]
	return ok
}

func (c MapClaims) Set(claim string, value interface{}) {
	c[claim] = value
}

// BasicClaims imlements Claims
type BasicClaims struct {
	StructClaimsMapper
	Audience  string    `claim:"aud"`
	ExpiresAt time.Time `claim:"exp"`
	Id        string    `claim:"jti"`
	IssuedAt  time.Time `claim:"iat"`
	Issuer    string    `claim:"iss"`
	NotBefore time.Time `claim:"nbf"`
	Subject   string    `claim:"sub"`
}

func (c *BasicClaims) MarshalJSON() ([]byte, error) {
	return c.StructClaimsMapper.DoMarshalJSON(c)
}

func (c *BasicClaims) UnmarshalJSON(bytes []byte) error {
	return c.StructClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *BasicClaims) Get(claim string) interface{} {
	return c.StructClaimsMapper.Get(c, claim)
}

func (c *BasicClaims) Has(claim string) bool {
	return c.StructClaimsMapper.Has(c, claim)
}

func (c *BasicClaims) Set(claim string, value interface{}) {
	c.StructClaimsMapper.Set(c, claim, value)
}

