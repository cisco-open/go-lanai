package claims

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

var (
	UserInfoClaimSpecs = map[string]ClaimSpec{
		// Basic
		oauth2.ClaimIssuer:   {Func: Issuer, Req: true},
		oauth2.ClaimAudience: {Func: Audience, Req: true},
		oauth2.ClaimSubject:  {Func: Subject, Req: false},

		// OIDC
		oauth2.ClaimFullName:          {Func: FullName, Req: false},
		oauth2.ClaimFirstName:         {Func: FirstName, Req: false},
		oauth2.ClaimLastName:          {Func: LastName, Req: false},
		//oauth2.ClaimMiddleName:        {Func: TBD, Req: false},
		//oauth2.ClaimNickname:          {Func: TBD, Req: false},
		oauth2.ClaimPreferredUsername: {Func: Username, Req: false},
		//oauth2.ClaimProfileUrl:        {Func: TBD, Req: false},
		//oauth2.ClaimPictureUrl:        {Func: TBD, Req: false},
		//oauth2.ClaimWebsite:           {Func: TBD, Req: false},
		oauth2.ClaimEmail:             {Func: Email, Req: false},
		oauth2.ClaimEmailVerified:     {Func: EmailVerified, Req: false},
		//oauth2.ClaimGender:            {Func: TBD, Req: false},
		//oauth2.ClaimBirthday:          {Func: TBD, Req: false},
		oauth2.ClaimZoneInfo:          {Func: ZoneInfo, Req: false},
		oauth2.ClaimLocale:            {Func: Locale, Req: false},
		//oauth2.ClaimPhoneNumber:       {Func: TBD, Req: false},
		//oauth2.ClaimPhoneNumVerified:  {Func: TBD, Req: false},
		oauth2.ClaimAddress:           {Func: Address, Req: false},
		//oauth2.ClaimUpdatedAt:         {Func: TBD, Req: false},

		// Custom
		oauth2.ClaimAccountType:     {Func: AccountType, Req: false},
		oauth2.ClaimDefaultTenantId: {Func: DefaultTenantId, Req: false},
		oauth2.ClaimAssignedTenants: {Func: AssignedTenants, Req: false},
		oauth2.ClaimRoles:           {Func: Roles, Req: false},
		oauth2.ClaimPermissions:     {Func: Permissions, Req: false},
	}
)
