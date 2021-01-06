package oauth2

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestAccessTokenJSONSerialization(t *testing.T) {
	token := NewDefaultAccessToken("token value")
	token.
		SetExpireTime(time.Now().Add(2 * time.Hour)).
		PutClaim("c1", "v1").
		PutClaim("c2", "v2").
		PutDetails("d1", "v1").
		PutDetails("d2", "v1").
		AddScopes("s1", "s2")

	bytes , err := json.Marshal(token)
	str := string(bytes)
	fmt.Printf("JSON: %s\n", str)

	switch {
	case err != nil:
		t.Errorf("Marshalling should not return error")

	case len(str) == 0:
		t.Errorf("json should not be empty")

	//TODO more cases
	}

	// Deserialize
	parsed := NewDefaultAccessToken("")
	err = json.Unmarshal([]byte(str), &parsed)

	switch {
	case err != nil:
		t.Errorf("Unmarshalling should not return error\n")

	case parsed.Value() != "token value":
		t.Errorf("parsed value should be [%s], but is [%s]\n", "token value", parsed.Value())

	case parsed.ExpiryTime().IsZero():
		t.Errorf("parsed expire time should not be zero\n")

	case len(parsed.Scopes()) != 2:
		t.Errorf("parsed scopes should have [%d] items, but has [%d]\n", 2, len(parsed.Scopes()))

	case len(parsed.Details()) != 2:
		t.Errorf("parsed details should have [%d] items, but has [%d]\n", 2, len(parsed.Details()))

	case len(parsed.Claims()) != 0:
		t.Errorf("parsed claims should be empty (ignored), but has [%d] items\n", len(parsed.Claims()))

		//TODO more cases
	}
}
