package jwt

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

var staticJwkStore = NewStaticJwkStore()
var anotherJwkStore = NewSingleJwkStore("default")
var claims = MapClaims {
	"aud": []string{"target"},
	"exp": time.Now().Add(24 * time.Hour).Unix(),
	"jti": uuid.New().String(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
	"iss": "sandbox",
	"sub": "user",
}

func TestJwtWithKeyRotator(t *testing.T) {
	enc := NewRS256JwtEncoder(staticJwkStore, "default")
	dec := NewRS256JwtDecoder(staticJwkStore, "default")

	// encoding
	value, err := enc.Encode(context.Background(), claims)
	switch {
	case err != nil:
		t.Errorf("Encoder should not return error. But got %v \n", err)

	case len(value) == 0:
		t.Errorf("JWT should not be empty")

		//TODO more cases
	}
	
	t.Logf("JWT: %s", value)

	// decode
	parsed, err := dec.Decode(context.Background(), value, MapClaims{})
	switch {
	case err != nil:
		t.Errorf("Decoder should not return error. But got %v \n", err)

	case parsed == nil:
		t.Errorf("Decoder should return non-nil claims \n")

		//TODO more cases

	default:
		if _, ok := parsed.(MapClaims); !ok {
			t.Errorf("MapClaims is expected, but got %T \n", parsed)
		}

		if err := mapClaimsEquals(claims, parsed.(MapClaims)); err != nil {
			t.Errorf("Decoded claims doesn't match orginal: %v \n", err)
		}
	}
}

func mapClaimsEquals(expected MapClaims, actual MapClaims) error {
	for k,v := range expected {
		var equals bool
		switch actual[k].(type) {
		case int:
			equals = actual[k].(int) == v.(int)
		case float64:
			equals = actual[k].(float64) == float64(v.(int64))
		case []string:
			equals = true
			for i, val := range v.([]string) {
				if val != actual[k].([]string)[i] {
					equals = false
					break
				}
			}
		case []interface{}:
			equals = true
			for i, val := range v.([]string) {
				if val != actual[k].([]interface{})[i] {
					equals = false
					break
				}
			}
		default:
			equals = actual[k] == v
		}
		if !equals {
			return fmt.Errorf("claim[%s] expected %v, but got %v", k, v, actual[k])
		}
	}
	return nil
}
