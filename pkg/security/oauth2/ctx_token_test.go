package oauth2

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestAccessTokenJSONSerialization(t *testing.T) {
	refresh := NewDefaultRefreshToken("refresh token value").PutDetails("d1", "v1")
	token := NewDefaultAccessToken("token value")
	token.
		SetExpireTime(time.Now().Add(2 * time.Hour)).
		PutClaim("c1", "v1").
		PutClaim("c2", "v2").
		PutDetails("d1", "v1").
		PutDetails("d2", "v1").
		AddScopes("s1", "s2").
		SetRefreshToken(refresh)

	bytes , err := json.Marshal(token)
	str := string(bytes)
	fmt.Printf("JSON: %s\n", str)

	switch {
	case err != nil:
		t.Errorf("Marshalling should not return error. But got %v \n", err)

	case len(str) == 0:
		t.Errorf("json should not be empty")

	//TODO more cases
	}

	// Deserialize
	parsed := NewDefaultAccessToken("")
	err = json.Unmarshal([]byte(str), &parsed)

	switch {
	case err != nil:
		t.Errorf("Unmarshalling should not return error. But got %v \n", err)

	case parsed.Value() != "token value":
		t.Errorf("parsed value should be [%s], but is [%s]\n", "token value", parsed.Value())

	case parsed.Type().HttpHeader() != "Bearer":
		t.Errorf("parsed token http header should be [%s], but is [%s]\n", "Bearer", parsed.Type().HttpHeader())

	case parsed.IssueTime().IsZero():
		t.Errorf("parsed issue time should not be zero\n")

	case parsed.ExpiryTime().IsZero():
		t.Errorf("parsed expiry time should not be zero\n")

	case len(parsed.Scopes()) != 2:
		t.Errorf("parsed scopes should have [%d] items, but has [%d]\n", 2, len(parsed.Scopes()))

	case len(parsed.Details()) != 2:
		t.Errorf("parsed details should have [%d] items, but has [%d]\n", 2, len(parsed.Details()))

	case parsed.Claims != nil:
		t.Errorf("parsed claims should be empty (ignored), but got %v\n", parsed.Claims)

	case parsed.RefreshToken().Value() != "refresh token value":
		t.Errorf("parsed refresh token should be correct\n")

		//TODO more cases
	}
}
