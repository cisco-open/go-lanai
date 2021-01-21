package jwt

import (
	"encoding/json"
	"errors"
)

/*********************
	Abstracts
 *********************/
type Claims interface {
	Get(claim string) interface{}
	Has(claim string) bool
}

/*********************
	Implements
 *********************/
// MapClaims imlements Claims and jwt.Claims
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

func (c MapClaims) Valid() (err error) {
	if c == nil {
		err = errors.New("MapClaims is nil")
	}
	return
}

// jwtGoCompatibleClaims implements jwt.Claims and has its own json serialization/deserialization
type jwtGoCompatibleClaims struct {
	claims interface{}
}

func (c *jwtGoCompatibleClaims) Valid() error {
	if c.claims == nil {
		return errors.New("embedded claims are nil")
	}
	return nil
}

func (c *jwtGoCompatibleClaims) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.claims)
}

func (c *jwtGoCompatibleClaims) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, c.claims)
}

