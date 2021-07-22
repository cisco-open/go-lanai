package claims

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

var (
	IdTokenBasicSpecs = map[string]ClaimSpec{
		// Basic
		oauth2.ClaimIssuer:          Required(Issuer),
		oauth2.ClaimSubject:         Required(Subject),
		oauth2.ClaimAudience:        Required(Audience),
		oauth2.ClaimExpire:          Required(ExpiresAt),
		oauth2.ClaimIssueAt:         Required(IssuedAt),
		oauth2.ClaimAuthTime:        RequiredIfParamsExists(AuthenticationTime, oauth2.ParameterMaxAge),
		oauth2.ClaimNonce:           RequiredIfParamsExists(Nonce, oauth2.ParameterNonce),
		oauth2.ClaimAuthCtxClassRef: Unsupported(),                // TODO acr
		oauth2.ClaimAuthMethodRef:   Unsupported(),                // TODO amr
		oauth2.ClaimAuthorizedParty: Optional(ClientId),                // TODO azp
		oauth2.ClaimAccessTokenHash: RequiredIfImplicitFlow(AccessTokenHash),

		// Custom Profile
		oauth2.ClaimUserId:          Optional(UserId),
		oauth2.ClaimAccountType:     Optional(AccountType),
		oauth2.ClaimTenantId:        Optional(TenantId),
		oauth2.ClaimTenantName:      Optional(TenantName),
		oauth2.ClaimTenantSuspended: Optional(TenantSuspended),
		oauth2.ClaimProviderId:      Optional(ProviderId),
		oauth2.ClaimProviderName:    Optional(ProviderName),
		oauth2.ClaimOrigUsername:    Optional(OriginalUsername),
		oauth2.ClaimCurrency:        Optional(Currency),
	}
)
