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
	TokenTypeMac    = "mac"
	TokenTypeBasic  = "basic"
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

type Token interface {
	Value() string
	ExpiryTime() time.Time
	Expired() bool
	Details() map[string]interface{}
}

type AccessToken interface {
	Token
	Type() TokenType
	IssueTime() time.Time
	Scopes() utils.StringSet
	RefreshToken() RefreshToken
}

type RefreshToken interface {
	Token
	WillExpire() bool
}

/*******************************
	Common Impl. AccessToken
 *******************************/
// DefaultAccessToken implements AccessToken
type DefaultAccessToken struct {
	Claims       Claims
	tokenType    TokenType
	value        string
	expiryTime   time.Time
	issueTime	 time.Time
	scopes       utils.StringSet
	refreshToken *DefaultRefreshToken
	details      map[string]interface{}
}

func NewDefaultAccessToken(value string) *DefaultAccessToken {
	return &DefaultAccessToken{
		value:     value,
		tokenType: TokenTypeBearer,
		scopes:    utils.NewStringSet(),
		issueTime: time.Now(),
		details:   map[string]interface{}{},
	}
}

func FromAccessToken(token AccessToken) *DefaultAccessToken {
	if t, ok := token.(*DefaultAccessToken); ok {
		return &DefaultAccessToken{
			value:        t.value,
			tokenType:    t.tokenType,
			expiryTime:   t.expiryTime,
			issueTime:    t.issueTime,
			scopes:       t.scopes.Copy(),
			Claims:       t.Claims,
			details:      copyMap(t.details),
			refreshToken: t.refreshToken,
		}
	}

	copy := &DefaultAccessToken{
		value:      token.Value(),
		tokenType:  token.Type(),
		expiryTime: token.ExpiryTime(),
		issueTime:  token.IssueTime(),
		scopes:     token.Scopes().Copy(),
		details:    copyMap(token.Details()),
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
func (t *DefaultAccessToken) IssueTime() time.Time {
	return t.issueTime
}

// AccessToken
func (t *DefaultAccessToken) ExpiryTime() time.Time {
	return t.expiryTime
}

// AccessToken
func (t *DefaultAccessToken) Expired() bool {
	return !t.expiryTime.IsZero() && t.expiryTime.Before(time.Now())
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

// Setters
func (t *DefaultAccessToken) SetValue(v string) *DefaultAccessToken {
	t.value = v
	return t
}

func (t *DefaultAccessToken) SetIssueTime(v time.Time) *DefaultAccessToken {
	t.issueTime = v.UTC()
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
	t.Claims.Set(key, value)
	return t
}

/********************************
	Common Impl. RefreshToken
 ********************************/
type DefaultRefreshToken struct {
	Claims       Claims
	value        string
	expiryTime   time.Time
	details      map[string]interface{}
}

func NewDefaultRefreshToken(value string) *DefaultRefreshToken {
	return &DefaultRefreshToken{
		value: value,
		details: map[string]interface{}{},
	}
}

func FromRefreshToken(token RefreshToken) *DefaultRefreshToken {
	if t, ok := token.(*DefaultRefreshToken); ok {
		return &DefaultRefreshToken{
			value: t.value,
			details: copyMap(t.details),
			Claims: t.Claims,
		}
	}

	return &DefaultRefreshToken{
		value: token.Value(),
		details: copyMap(token.Details()),
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

// RefreshToken
func (t *DefaultRefreshToken) ExpiryTime() time.Time {
	return t.expiryTime
}

// RefreshToken
func (t *DefaultRefreshToken) Expired() bool {
	return !t.expiryTime.IsZero() && t.expiryTime.Before(time.Now())
}

// RefreshToken
func (t *DefaultRefreshToken) WillExpire() bool {
	return !t.expiryTime.IsZero()
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
	t.Claims.Set(key, value)
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


