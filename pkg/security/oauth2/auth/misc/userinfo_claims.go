package misc

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type UserInfoClaims struct {
	oauth2.FieldClaimsMapper

	/*******************************
	 * Standard JWT claims
	 *******************************/
	Issuer   string          `claim:"iss"`
	Audience utils.StringSet `claim:"aud"`
	Subject  string          `claim:"sub"`

	/*******************************
	* Standard OIDC claims
	*******************************/
	FullName          string               `claim:"name"`
	FirstName         string               `claim:"given_name"`
	LastName          string               `claim:"family_name"`
	MiddleName        string               `claim:"middle_name"`
	Nickname          string               `claim:"nickname"`
	PreferredUsername string               `claim:"preferred_username"`
	ProfileUrl        string               `claim:"profile"`
	PictureUrl        string               `claim:"picture"`
	Website           string               `claim:"website"`
	Email             string               `claim:"email"`
	EmailVerified     *bool                `claim:"email_verified"`
	Gender            string               `claim:"gender"`
	Birthday          string               `claim:"birthdate"`    // ISO 8601:2004 [ISO8601‑2004] YYYY-MM-DD format
	ZoneInfo          string               `claim:"zoneinfo"`     // Europe/Paris or America/Los_Angeles
	Locale            string               `claim:"locale"`       // Typically ISO 639-1 Alpha-2 [ISO639‑1] language code in lowercase and an ISO 3166-1
	PhoneNumber       string               `claim:"phone_number"` // RFC 3966 [RFC3966] e.g. +1 (604) 555-1234;ext=5678
	PhoneNumVerified  *bool                `claim:"phone_number_verified"`
	Address           *claims.AddressClaim `claim:"address"`
	UpdatedAt         time.Time            `claim:"updated_at"`

	/*******************************
	 * NFV Additional Claims
	 *******************************/
	AccountType     string          `claim:"account_type"`
	DefaultTenantId string          `claim:"default_tenant_id"`
	AssignedTenants utils.StringSet `claim:"assigned_tenants"`
	Roles           utils.StringSet `claim:"roles"`
	Permissions     utils.StringSet `claim:"permissions"`
}

func (c UserInfoClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *UserInfoClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c UserInfoClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c UserInfoClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *UserInfoClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

func (c UserInfoClaims) Values() map[string]interface{} {
	return c.FieldClaimsMapper.Values(c)
}
