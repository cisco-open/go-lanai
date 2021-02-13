package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// DecodedAccessToken implements oauth2.AccessToken
type DecodedAccessToken struct {
	Claims     *ExtendedClaims
	TokenValue string
	ExpireAt   time.Time
	IssuedAt   time.Time
	ScopesSet  utils.StringSet
}

func NewDecodedAccessToken() *DecodedAccessToken {
	return &DecodedAccessToken{}
}

func (t *DecodedAccessToken) Value() string {
	return t.TokenValue
}

func (t *DecodedAccessToken) ExpiryTime() time.Time {
	return t.ExpireAt
}

func (t *DecodedAccessToken) Expired() bool {
	return !t.ExpireAt.IsZero() && t.ExpireAt.Before(time.Now())
}

func (t *DecodedAccessToken) Details() map[string]interface{} {
	return map[string]interface{}{}
}

func (t *DecodedAccessToken) Type() oauth2.TokenType {
	return oauth2.TokenTypeBearer
}

func (t *DecodedAccessToken) IssueTime() time.Time {
	return t.IssuedAt
}

func (t *DecodedAccessToken) Scopes() utils.StringSet {
	return t.ScopesSet
}

func (t *DecodedAccessToken) RefreshToken() oauth2.RefreshToken {
	return nil
}

// DecodedRefreshToken implements oauth2.RefreshToken
type DecodedRefreshToken struct {
	Claims     *ExtendedClaims
	TokenValue string
	ExpireAt   time.Time
	IssuedAt   time.Time
	ScopesSet  utils.StringSet
}

func (t *DecodedRefreshToken) Value() string {
	return t.TokenValue
}

func (t *DecodedRefreshToken) ExpiryTime() time.Time {
	return t.ExpireAt
}

func (t *DecodedRefreshToken) Expired() bool {
	return !t.ExpireAt.IsZero() && t.ExpireAt.Before(time.Now())
}

func (t *DecodedRefreshToken) Details() map[string]interface{} {
	return map[string]interface{}{}
}

func (t *DecodedRefreshToken) WillExpire() bool {
	return !t.ExpireAt.IsZero()
}

