package claims

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

var (
	CheckTokenClaimSpecs = map[string]ClaimSpec {
		// Basic
		oauth2.ClaimAudience:  {Func: Audience, Req: true},
		oauth2.ClaimExpire:    {Func: ExpiresAt, Req: false},
		oauth2.ClaimJwtId:     {Func: JwtId, Req: false},
		oauth2.ClaimIssueAt:   {Func: IssuedAt, Req: false},
		oauth2.ClaimIssuer:    {Func: Issuer, Req: true},
		oauth2.ClaimNotBefore: {Func: NotBefore, Req: false},
		oauth2.ClaimSubject:   {Func: Subject, Req: false},
		oauth2.ClaimScope:     {Func: Scopes, Req: false},
		oauth2.ClaimClientId:  {Func: ClientId, Req: true},
		oauth2.ClaimUsername:  {Func: Username, Req: false},

		// OIDC
		oauth2.ClaimAuthTime:  {Func: AuthenticationTime, Req: false},
		oauth2.ClaimFirstName: {Func: FirstName, Req: false},
		oauth2.ClaimLastName:  {Func: LastName, Req: false},
		oauth2.ClaimEmail:     {Func: Email, Req: false},
		oauth2.ClaimLocale:    {Func: Locale, Req: false},

		// Custom
		oauth2.ClaimUserId:          {Func: UserId, Req: false},
		oauth2.ClaimAccountType:     {Func: AccountType, Req: false},
		oauth2.ClaimCurrency:        {Func: Currency, Req: false},
		oauth2.ClaimAssignedTenants: {Func: AssignedTenants, Req: false},
		oauth2.ClaimTenantId:        {Func: TenantId, Req: false},
		oauth2.ClaimTenantName:      {Func: TenantName, Req: false},
		oauth2.ClaimTenantSuspended: {Func: TenantSuspended, Req: false},
		oauth2.ClaimProviderId:      {Func: ProviderId, Req: false},
		oauth2.ClaimProviderName:    {Func: ProviderName, Req: false},
		oauth2.ClaimRoles:           {Func: Roles, Req: false},
		oauth2.ClaimPermissions:     {Func: Permissions, Req: false},
		oauth2.ClaimOrigUsername:    {Func: OriginalUsername, Req: false},
	}
)
