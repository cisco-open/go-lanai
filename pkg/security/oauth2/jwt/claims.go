package jwt

import "encoding/json"

/*********************
	Abstracts
 *********************/
type Claims interface {
	json.Marshaler
	json.Unmarshaler
	Value(key string) interface{}
}

/*********************
	Implements
 *********************/
// jwtGoCompatibleClaims implements jwt.Claims and has its own json serialization/deserialization
type jwtGoCompatibleClaims struct {
	claims Claims
}

func (c *jwtGoCompatibleClaims) Valid() error {
	return nil
}

func (c *jwtGoCompatibleClaims) MarshalJSON() ([]byte, error) {
	return c.claims.(json.Marshaler).MarshalJSON()
}

func (c *jwtGoCompatibleClaims) UnmarshalJSON(bytes []byte) error {
	return c.claims.(json.Unmarshaler).UnmarshalJSON(bytes)
}

// MapClaims imlements Claims
type MapClaims map[string]interface{}

func (c MapClaims) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(c))
}

func (c MapClaims) UnmarshalJSON(bytes []byte) error {
	m := map[string]interface{}(c)
	return json.Unmarshal(bytes, &m)
}

func (c MapClaims) Value(key string) interface{} {
	return c[key]
}

