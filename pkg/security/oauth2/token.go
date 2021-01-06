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

/*******************************
	Common Impl. AccessToken
 *******************************/
// DefaultAccessToken implements AccessToken
type DefaultAccessToken struct {
	tokenType    TokenType
	value        string
	expiryTime   time.Time
	scopes       utils.StringSet
	refreshToken *DefaultRefreshToken
	details      map[string]interface{}
	claims       map[string]interface{}
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

func FromAccessToken(token AccessToken) *DefaultAccessToken {
	if t, ok := token.(*DefaultAccessToken); ok {
		return &DefaultAccessToken{
			value: t.value,
			tokenType: t.tokenType,
			expiryTime: t.expiryTime,
			scopes: t.scopes.Copy(),
			claims: copyMap(t.claims),
			details: copyMap(t.details),
			refreshToken: t.refreshToken,
		}
	}

	copy := &DefaultAccessToken{
		value: token.Value(),
		tokenType: token.Type(),
		expiryTime: token.ExpiryTime(),
		scopes: token.Scopes().Copy(),
		claims: map[string]interface{}{},
		details: copyMap(token.Details()),
	}
	copy.SetRefreshToken(token.RefreshToken())
	return copy
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
	if t.refreshToken == nil {
		return nil
	}
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
	if refresh, ok := v.(*DefaultRefreshToken); ok {
		t.refreshToken = refresh
	} else {
		t.refreshToken = FromRefreshToken(v)
	}
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

/********************************
	Common Impl. RefreshToken
 ********************************/
type DefaultRefreshToken struct {
	value        string
	details      map[string]interface{}
	claims       map[string]interface{} // TODO
}

func NewDefaultRefreshToken(value string) *DefaultRefreshToken {
	return &DefaultRefreshToken{
		value: value,
		details: map[string]interface{}{},
		claims: map[string]interface{}{},
	}
}

func FromRefreshToken(token RefreshToken) *DefaultRefreshToken {
	if t, ok := token.(*DefaultRefreshToken); ok {
		return &DefaultRefreshToken{
			value: t.value,
			details: copyMap(t.details),
			claims: copyMap(t.claims),
		}
	}

	return &DefaultRefreshToken{
		value: token.Value(),
		details: copyMap(token.Details()),
		claims: map[string]interface{}{},
	}
}

// RefreshToken
func (t *DefaultRefreshToken) Value() string {
	return t.value
}

// RefreshToken
func (t *DefaultRefreshToken) Details() map[string]interface{} {
	return t.details
}

func (t *DefaultRefreshToken) Claims() map[string]interface{} {
	return t.claims
}

// Setters
func (t *DefaultRefreshToken) SetValue(v string) *DefaultRefreshToken {
	t.value = v
	return t
}

func (t *DefaultRefreshToken) PutDetails(key string, value interface{}) *DefaultRefreshToken {
	if value == nil {
		delete(t.details, key)
	} else {
		t.details[key] = value
	}
	return t
}

func (t *DefaultRefreshToken) PutClaim(key string, value interface{}) *DefaultRefreshToken {
	if value == nil {
		delete(t.claims, key)
	} else {
		t.claims[key] = value
	}
	return t
}

/********************************
	Helpers
 ********************************/
func copyMap(src map[string]interface{}) map[string]interface{} {
	dest := map[string]interface{}{}
	for k,v := range src {
		dest[k] = v
	}
	return dest
}


