package oauth2

const (
	JsonFieldAccessTokenValue  = "access_token"
	JsonFieldTokenType         = "token_type"
	JsonFieldIssueTime         = "iat"
	JsonFieldExpiryTime        = "expiry"
	JsonFieldExpiresIn         = "expires_in"
	JsonFieldScope             = "scope"
	JsonFieldRefreshTokenValue = "refresh_token"
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
	ParameterTenantName          = "tenant_name"
	ParameterNonce               = "nonce"
	ParameterError               = "error"
	ParameterErrorDescription    = "error_description"
	ParameterCodeChallenge       = "code_challenge"
	ParameterCodeChallengeMethod = "code_challenge_method"
	ParameterAuthCode            = "code"
	ParameterUserApproval        = "user_oauth_approval"
	ParameterRefreshToken        = "refresh_token"
	//Parameter = ""
)

const (
//Extension     = ""
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
	CtxKeyAuthorizationExpiryTime   = "kAuthorizationExpiryTime"
	CtxKeyAuthorizationIssueTime    = "kAuthorizationIssueTime"
	CtxKeyAuthenticationTime        = "kAuthenticationTime"
	CtxKeyReceivedAuthorizeRequest  = "kReceivedAuthRequest"
	CtxKeyValidatedAuthorizeRequest = "kValidatedAuthRequest"
	CtxKeyResolvedAuthorizeRedirect = "kResolvedRedirect"
	CtxKeyResolvedAuthorizeState    = "kResolvedState"
	CtxKeyRefreshToken              = "kRefreshToken"
)

const (
	GrantTypeClientCredentials = "client_credentials"
	GrantTypePassword          = "password"
	GrantTypeAuthCode          = "authorization_code"
	GrantTypeImplicit          = "implicit"
	GrantTypeRefresh           = "refresh_token"
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
	ClaimBirthday          = "birthdate"
	ClaimZoneInfo          = "zoneinfo"
	ClaimLocale            = "locale"
	ClaimPhoneNumber       = "phone_number"
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
	ClaimUserId          = "user_id"
	ClaimAccountType     = "account_type"
	ClaimCurrency        = "currency"
	ClaimTenantId        = "tenant_id"
	ClaimTenantName      = "tenant_name"
	ClaimTenantSuspended = "tenant_suspended"
	ClaimProviderId      = "provider_id"
	ClaimProviderName    = "provider_name"
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
