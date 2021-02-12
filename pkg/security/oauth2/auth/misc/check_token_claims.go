package misc

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// CheckTokenClaims implemnts oauth2.Claims
type CheckTokenClaims struct {
	oauth2.FieldClaimsMapper

	/*******************************
	 * Standard Check Token claims
	 *******************************/
	oauth2.BasicClaims
	Active   bool   `claim:"active"`
	Username string `claim:"username"`

	/*******************************
	* Standard OIDC claims
	*******************************/
	FirstName string    `claim:"given_name"`
	LastName  string    `claim:"family_name"`
	Email     string    `claim:"email"`
	Locale    string    `claim:"locale"` // Typically ISO 639-1 Alpha-2 [ISO639‑1] language code in lowercase and an ISO 3166-1
	AuthTime  time.Time `claim:"auth_time"`

	/*******************************
	 * NFV Additional Claims
	 *******************************/
	UserId          string          `claim:"user_id"`
	AccountType     string          `claim:"account_type"`
	Currency        string          `claim:"currency"`
	TenantId        string          `claim:"tenant_id"`
	TenantName      string          `claim:"tenant_name"`
	TenantSuspended bool            `claim:"tenant_suspended"`
	ProviderId      string          `claim:"provider_id"`
	ProviderName    string          `claim:"provider_name"`
	AssignedTenants utils.StringSet `claim:"assigned_tenants"`
	Roles           utils.StringSet `claim:"roles"`
	Permissions     utils.StringSet `claim:"permissions"`
	OrigUsername    string          `claim:"original_username"`
}
