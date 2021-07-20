package claims

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

var (
	IdTokenClaimSpecs = map[string]ClaimSpec{
		// Basic
		oauth2.ClaimIssuer:    {Func: Issuer, Req: true},
		oauth2.ClaimSubject:   {Func: Subject, Req: true},
		oauth2.ClaimAudience:  {Func: Audience, Req: true},
		oauth2.ClaimExpire:    {Func: ExpiresAt, Req: true},
		oauth2.ClaimIssueAt:   {Func: IssuedAt, Req: true},
		oauth2.ClaimAuthTime: {Func: AuthenticationTime, Req: false}, // TODO required vs optional
		oauth2.ClaimNonce: {Func: Unsupported, Req: false}, // TODO Nonce value and required vs optional (Required for implicit flow)
		oauth2.ClaimAuthCtxClassRef:     {Func: Unsupported, Req: false}, // TODO acr
		oauth2.ClaimAuthMethodRef:     {Func: Unsupported, Req: false}, // TODO amr
		oauth2.ClaimAuthorizedParty:     {Func: Unsupported, Req: false}, // TODO azp
		oauth2.ClaimAccessTokenHash:     {Func: Unsupported, Req: false}, // TODO hash

		// Profile Scope
		oauth2.ClaimFullName:          {Func: FullName, Req: false},
		oauth2.ClaimFirstName:         {Func: FirstName, Req: false},
		oauth2.ClaimLastName:          {Func: LastName, Req: false},
		oauth2.ClaimMiddleName:        {Func: Unsupported, Req: false},
		oauth2.ClaimNickname:          {Func: Unsupported, Req: false},
		oauth2.ClaimPreferredUsername: {Func: Username, Req: false},
		oauth2.ClaimProfileUrl:        {Func: Unsupported, Req: false},
		oauth2.ClaimPictureUrl:        {Func: Unsupported, Req: false},
		oauth2.ClaimWebsite:           {Func: Unsupported, Req: false},
		oauth2.ClaimGender:            {Func: Unsupported, Req: false},
		oauth2.ClaimBirthday:          {Func: Unsupported, Req: false},
		oauth2.ClaimZoneInfo:          {Func: ZoneInfo, Req: false},
		oauth2.ClaimLocale:            {Func: Locale, Req: false},
		oauth2.ClaimUpdatedAt:         {Func: Unsupported, Req: false}, // TODO
		oauth2.ClaimDefaultTenantId:   {Func: DefaultTenantId, Req: false},
		oauth2.ClaimAssignedTenants:   {Func: AssignedTenants, Req: false},
		oauth2.ClaimRoles:             {Func: Roles, Req: false},
		oauth2.ClaimPermissions:       {Func: Permissions, Req: false},

		// email scope
		oauth2.ClaimEmail:         {Func: Email, Req: false},
		oauth2.ClaimEmailVerified: {Func: EmailVerified, Req: false},

		// phone number scope
		oauth2.ClaimPhoneNumber:      {Func: Unsupported, Req: false},
		oauth2.ClaimPhoneNumVerified: {Func: Unsupported, Req: false},

		// address scope
		oauth2.ClaimAddress:           {Func: Address, Req: false},

		// Custom Profile
		oauth2.ClaimUserId:          {Func: UserId, Req: false},
		oauth2.ClaimAccountType:     {Func: AccountType, Req: false},
		oauth2.ClaimTenantId:        {Func: TenantId, Req: false},
		oauth2.ClaimTenantName:      {Func: TenantName, Req: false},
		oauth2.ClaimTenantSuspended: {Func: TenantSuspended, Req: false},
		oauth2.ClaimProviderId:      {Func: ProviderId, Req: false},
		oauth2.ClaimProviderName:    {Func: ProviderName, Req: false},
		oauth2.ClaimOrigUsername:    {Func: OriginalUsername, Req: false},
		oauth2.ClaimCurrency:        {Func: Currency, Req: false},
	}
)
