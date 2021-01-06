package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
	"time"
)

/*****************************
	Abstractions
 *****************************/
type TokenType string
const(
	TokenTypeBearer = "bearer"
	TokenTypeMac = "mac"
	TokenTypeBasic = "basic"
)

func (t TokenType) HttpHeader() string {
	switch strings.ToLower(string(t)) {
	case TokenTypeMac:
		return "MAC"
	case TokenTypeBasic:
		return "Basic"
	default:
		return "Bearer"
	}
}

type AccessToken interface {
	Value() string
	Type() TokenType
	ExpiryTime() time.Time
	Expired() bool
	Scopes() utils.StringSet
	Details() map[string]interface{}
	RefreshToken() RefreshToken
}

type RefreshToken interface {
	Value() string
	Details() map[string]interface{}
}

/*****************************
	Common Implementations
 *****************************/
// DefaultAccessToken implements AccessToken, json.
type DefaultAccessToken struct {
	tokenType    TokenType
	value        string
	expiryTime   time.Time
	scopes       utils.StringSet
	refreshToken RefreshToken
	details      map[string]interface{}
	claims       map[string]interface{}
}

// embeded accessToken for marshaling
type accessToken struct {

}

func NewDefaultAccessToken(value string) *DefaultAccessToken {
	return &DefaultAccessToken{
		value: value,
		tokenType: TokenTypeBearer,
		scopes: utils.NewStringSet(),
		claims: map[string]interface{}{},
		details: map[string]interface{}{},
	}
}

// AccessToken
func (t *DefaultAccessToken) Value() string {
	return t.value
}

// AccessToken
func (t *DefaultAccessToken) Details() map[string]interface{} {
	return t.details
}

// AccessToken
func (t *DefaultAccessToken) Type() TokenType {
	return t.tokenType
}

// AccessToken
func (t *DefaultAccessToken) ExpiryTime() time.Time {
	return t.expiryTime
}

func (t *DefaultAccessToken) Expired() bool {
	return !t.expiryTime.IsZero() && t.expiryTime.After(time.Now())
}

// AccessToken
func (t *DefaultAccessToken) Scopes() utils.StringSet {
	return t.scopes
}

// AccessToken
func (t *DefaultAccessToken) RefreshToken() RefreshToken {
	return t.refreshToken
}

func (t *DefaultAccessToken) Claims() map[string]interface{} {
	return t.claims
}

// Setters
func (t *DefaultAccessToken) SetValue(v string) *DefaultAccessToken {
	t.value = v
	return t
}

func (t *DefaultAccessToken) SetExpireTime(v time.Time) *DefaultAccessToken {
	t.expiryTime = v.UTC()
	return t
}

func (t *DefaultAccessToken) SetRefreshToken(v RefreshToken) *DefaultAccessToken {
	t.refreshToken = v
	return t
}

func (t *DefaultAccessToken) AddScopes(scopes...string) *DefaultAccessToken {
	t.scopes.Add(scopes...)
	return t
}

func (t *DefaultAccessToken) RemoveScopes(scopes...string) *DefaultAccessToken {
	t.scopes.Remove(scopes...)
	return t
}

func (t *DefaultAccessToken) PutDetails(key string, value interface{}) *DefaultAccessToken {
	if value == nil {
		delete(t.details, key)
	} else {
		t.details[key] = value
	}

	return t
}

func (t *DefaultAccessToken) PutClaim(key string, value interface{}) *DefaultAccessToken {
	if value == nil {
		delete(t.claims, key)
	} else {
		t.claims[key] = value
	}
	return t
}




