package oauth2

const (
	JsonFieldAccessTokenValue  = "access_token"
	JsonFieldTokenType         = "token_type"
	JsonFieldIssueTime         = "iat"
	JsonFieldExpiryTime        = "expiry"
	JsonFieldExpiresIn         = "expires_in"
	JsonFieldScope             = "scope"
	JsonFieldRefreshTokenValue = "refresh_token"
	JsonFieldIDTokenValue      = "id_token"
)

const (
	ParameterClientId            = "client_id"
	ParameterClientSecret        = "client_secret"
	ParameterResponseType        = "response_type"
	ParameterRedirectUri         = "redirect_uri"
	ParameterScope               = "scope"
	ParameterState               = "state"
	ParameterGrantType           = "grant_type"
	ParameterUsername            = "username"
	ParameterPassword            = "password"
	ParameterTenantId            = "tenant_id"
	ParameterTenantExternalId    = "tenant_name" //for backward compatibility we map it to tenant_name
	ParameterNonce               = "nonce"
	ParameterMaxAge              = "max_age"
	ParameterError               = "error"
	ParameterErrorDescription    = "error_description"
	ParameterCodeChallenge       = "code_challenge"
	ParameterCodeChallengeMethod = "code_challenge_method"
	ParameterCodeVerifier        = "code_verifier"
	ParameterRequestObj          = "request"
	ParameterRequestUri          = "request_uri"
	ParameterAuthCode            = "code"
	ParameterUserApproval        = "user_oauth_approval"
	ParameterRefreshToken        = "refresh_token"
	ParameterAccessToken         = "access_token"
	ParameterSwitchUsername      = "switch_username"
	ParameterSwitchUserId        = "switch_user_id"
	ParameterDisplay             = "display"
	ParameterACR                 = "acr_values"
	ParameterPrompt              = "prompt"
	ParameterClaims              = "claims"
	//Parameter = ""
)

const (
	ExtUseSessionTimeout = "use_session_timeout"
	//Ext     = ""
)

const (
	GrantTypeClientCredentials = "client_credentials"
	GrantTypePassword          = "password"
	GrantTypeAuthCode          = "authorization_code"
	GrantTypeImplicit          = "implicit"
	GrantTypeRefresh           = "refresh_token"
	GrantTypeSwitchUser        = "urn:cisco:nfv:oauth:grant-type:switch-user"
	GrantTypeSwitchTenant      = "urn:cisco:nfv:oauth:grant-type:switch-tenant"
	GrantTypeSamlSSO           = "urn:ietf:params:oauth:grant-type:saml2-bearer"
)

const (
	ScopeRead            = "read"
	ScopeWrite           = "write"
	ScopeTokenDetails    = "token_details"
	ScopeTenantHierarchy = "tenant_hierarchy"
	ScopeOidc            = "openid"
	ScopeOidcProfile     = "profile"
	ScopeOidcEmail       = "email"
	ScopeOidcAddress     = "address"
	ScopeOidcPhone       = "phone"
)

const (
	CtxKeyAuthenticatedClient       = "kAuthenticatedClient"
	CtxKeyAuthenticatedAccount      = "kAuthenticatedAccount"
	CtxKeyAuthorizedTenant          = "kAuthorizedTenant"
	CtxKeyAuthorizedProvider        = "kAuthorizedProvider"
	CtxKeyUserAuthentication        = "kUserAuthentication"
	CtxKeyAuthorizationExpiryTime   = "kAuthorizationExpiryTime"
	CtxKeyAuthorizationIssueTime    = "kAuthorizationIssueTime"
	CtxKeyAuthenticationTime        = "kAuthenticationTime"
	CtxKeyReceivedAuthorizeRequest  = "kReceivedAuthRequest"
	CtxKeyValidatedAuthorizeRequest = "kValidatedAuthRequest"
	CtxKeyResolvedAuthorizeRedirect = "kResolvedRedirect"
	CtxKeyResolvedAuthorizeState    = "kResolvedState"
	CtxKeySourceAuthentication      = "kSourceAuthentication"
	//CtxKeyRefreshToken              = "kRefreshToken"
)

const (
	DetailsKeyRequestExt    = "kOAuth2Ext"
	DetailsKeyRequestParams = "kOAuth2Params"
)

const (
	/**
	 * JWT standard
	 * https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-13#section-4.1
	 */
	ClaimIssuer    = "iss"
	ClaimSubject   = "sub"
	ClaimAudience  = "aud"
	ClaimExpire    = "exp"
	ClaimNotBefore = "nbf"
	ClaimIssueAt   = "iat"
	ClaimJwtId     = "jti"
	//Claim = ""

	/**
	 * ID TOKEN
	 * https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	 */
	ClaimAuthTime        = "auth_time"
	ClaimNonce           = "nonce"
	ClaimAuthCtxClassRef = "acr"
	ClaimAuthMethodRef   = "amr"
	ClaimAuthorizedParty = "azp"
	ClaimAccessTokenHash = "at_hash"
	//Claim = ""

	/**
	 * Standard UserInfo
	 * https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	 */
	ClaimFullName          = "name"
	ClaimFirstName         = "given_name"
	ClaimLastName          = "family_name"
	ClaimMiddleName        = "middle_name"
	ClaimNickname          = "nickname"
	ClaimPreferredUsername = "preferred_username"
	ClaimProfileUrl        = "profile"
	ClaimPictureUrl        = "picture"
	ClaimWebsite           = "website"
	ClaimEmail             = "email"
	ClaimEmailVerified     = "email_verified"
	ClaimGender            = "gender"
	ClaimBirthday          = "birthdate"    // ISO 8601:2004 [ISO8601‑2004] YYYY-MM-DD format
	ClaimZoneInfo          = "zoneinfo"     // Europe/Paris or America/Los_Angeles
	ClaimLocale            = "locale"       // Typically ISO 639-1 Alpha-2 [ISO639‑1] language code in lowercase and an ISO 3166-1
	ClaimPhoneNumber       = "phone_number" // RFC 3966 [RFC3966] e.g. +1 (604) 555-1234;ext=5678
	ClaimPhoneNumVerified  = "phone_number_verified"
	ClaimAddress           = "address"
	ClaimUpdatedAt         = "updated_at"
	//Claim = ""

	/**
	 * Standard CheckToken
	 * https://tools.ietf.org/html/rfc7662#section-2.2
	 */
	ClaimActive    = "active"
	ClaimScope     = "scope"
	ClaimClientId  = "client_id"
	ClaimUsername  = "username"
	ClaimTokenType = "token_type"
	//Claim = ""

	/**
	 * NFV Additions - custom
	 */
	ClaimUserId                   = "user_id"
	ClaimAccountType              = "account_type"
	ClaimCurrency                 = "currency"
	ClaimTenantId                 = "tenant_id"
	ClaimTenantExternalId         = "tenant_name" //for backward compatibility we map it to tenant_name
	ClaimTenantSuspended          = "tenant_suspended"
	ClaimProviderId               = "provider_id"
	ClaimProviderName             = "provider_name"
	ClaimProviderDisplayName      = "provider_display_name"
	ClaimProviderDescription      = "provider_description"
	ClaimProviderEmail            = "provider_email"
	ClaimProviderNotificationType = "provider_notification_type"

	ClaimAssignedTenants = "assigned_tenants"
	ClaimRoles           = "roles"
	ClaimPermissions     = "permissions"
	ClaimOrigUsername    = "original_username"
	ClaimDefaultTenantId = "default_tenant_id"
	//Claim = ""

	/**
	 * NFV Additions - Legacy
	 */
	ClaimLegacyTenantId  = "tenantId"
	ClaimLegacyFirstName = "firstName"
	ClaimLegacyLastName  = "lastName"
	ClaimLegacyUsername  = "user_name"
)

const (
	LegacyResourceId = "nfv-api"
)
