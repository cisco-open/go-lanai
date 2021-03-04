package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"reflect"
	"time"
)

const (
	ClaimTag = "claim"
)

type Claims interface {
	Get(claim string) interface{}
	Has(claim string) bool
	Set(claim string, value interface{})
	Values() map[string]interface{}
}

/*********************
	Implements
 *********************/
// MapClaims imlements Claims & claimsMapper
type MapClaims map[string]interface{}

func (c MapClaims) MarshalJSON() ([]byte, error) {
	m, e := c.toMap(true)
	if e != nil {
		return nil, e
	}
	return json.Marshal(m)
}

func (c MapClaims) UnmarshalJSON(bytes []byte) error {
	m := map[string]interface{}{}
	if e := json.Unmarshal(bytes, &m); e != nil {
		return e
	}
	c.fromMap(m)
	return nil
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

func (c MapClaims) Values() map[string]interface{} {
	ret, e := c.toMap(false)
	if e != nil {
		return map[string]interface{}{};
	}
	return ret
}

func (c MapClaims) toMap(convert bool) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for k, v := range c {
		if convert {
			value, e := claimMarshalConvert(reflect.ValueOf(v));
			if e != nil {
				return nil, e
			}
			ret[k] = value.Interface()
		} else {
			ret[k] = v
		}
	}
	return ret, nil
}

func (c MapClaims) fromMap(src map[string]interface{}) error {

	for k, v := range src {
		value, e := claimUnmarshalConvert(reflect.ValueOf(v), anyType)
		if e != nil {
			return e
		}
		c[k] = value.Interface()
	}
	return nil
}

// BasicClaims imlements Claims
type BasicClaims struct {
	FieldClaimsMapper
	Audience  utils.StringSet `claim:"aud"`
	ExpiresAt time.Time       `claim:"exp"`
	Id        string          `claim:"jti"`
	IssuedAt  time.Time       `claim:"iat"`
	Issuer    string          `claim:"iss"`
	NotBefore time.Time       `claim:"nbf"`
	Subject   string          `claim:"sub"`
	Scopes    utils.StringSet `claim:"scope"`
	ClientId  string          `claim:"client_id"`
}

func (c *BasicClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *BasicClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *BasicClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *BasicClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *BasicClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c *BasicClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}
