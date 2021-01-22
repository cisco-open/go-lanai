package jwt

import (
	"encoding/json"
	"errors"
)

/*********************
	Implements
 *********************/
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

