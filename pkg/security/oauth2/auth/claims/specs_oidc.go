package claims

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"

var (
	ProfileScopeSpecs = map[string]ClaimSpec{
		// Profile Scope
		oauth2.ClaimFullName:          Optional(FullName),
		oauth2.ClaimFirstName:         Optional(FirstName),
		oauth2.ClaimLastName:          Optional(LastName),
		oauth2.ClaimMiddleName:        Unsupported(),
		oauth2.ClaimNickname:          Unsupported(),
		oauth2.ClaimPreferredUsername: Optional(Username),
		oauth2.ClaimProfileUrl:        Unsupported(),
		oauth2.ClaimPictureUrl:        Unsupported(),
		oauth2.ClaimWebsite:           Unsupported(),
		oauth2.ClaimGender:            Unsupported(),
		oauth2.ClaimBirthday:          Unsupported(),
		oauth2.ClaimZoneInfo:          Optional(ZoneInfo),
		oauth2.ClaimLocale:            Optional(Locale),
		oauth2.ClaimCurrency:          Optional(Currency),
		oauth2.ClaimUpdatedAt:         Unsupported(),
		oauth2.ClaimDefaultTenantId:   Optional(DefaultTenantId),
		oauth2.ClaimAssignedTenants:   Optional(AccountAssignedTenants),
		oauth2.ClaimRoles:             Optional(Roles),
		oauth2.ClaimPermissions:       Optional(Permissions),
	}

	EmailScopeSpecs = map[string]ClaimSpec{
		oauth2.ClaimEmail:         Optional(Email),
		oauth2.ClaimEmailVerified: Optional(EmailVerified),
	}

	PhoneScopeSpecs = map[string]ClaimSpec{
		oauth2.ClaimPhoneNumber:      Unsupported(),
		oauth2.ClaimPhoneNumVerified: Unsupported(),
	}

	AddressScopeSpecs = map[string]ClaimSpec{
		oauth2.ClaimAddress: Optional(Address),
	}
)
