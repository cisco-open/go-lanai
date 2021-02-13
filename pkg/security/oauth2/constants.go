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
	ParameterClientId = "client_id"
	ParameterClientSecret = "client_secret"
	ParameterResponseType = "response_type"
	ParameterRedirectUri = "redirect_uri"
	ParameterScope = "scope"
	ParameterState = "state"
	ParameterGrantType = "grant_type"
	ParameterUsername = "username"
	ParameterPassword = "password"
	ParameterTenantId = "tenant_id"
	ParameterTenantName = "tenant_name"
	ParameterNonce = "nonce"
	ParameterError = "error"
	ParameterErrorDescription = "error_description"
	ParameterCodeChallenge = "code_challenge"
	ParameterCodeChallengeMethod = "code_challenge_method"
	ParameterAuthCode = "code"
	ParameterUserApproval = "user_oauth_approval"
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
)

const (
	GrantTypeClientCredentials = "client_credentials"
	GrantTypePassword = "password"
	GrantTypeAuthCode = "authorization_code"
	GrantTypeImplicit = "implicit"
	GrantTypeRefresh = "refresh_token"
)

const (
	/**
	 * JWT standard
	 * https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-13#section-4.1
	 */
	ClaimIssuer = "iss"
	ClaimSubject = "sub"
	ClaimAudience = "aud"
	ClaimExpire = "exp"
	ClaimNotBefore = "nbf"
	ClaimIssueAt = "iat"
	ClaimJwtId = "jti"
	//Claim = ""

	/**
	 * Standard CheckToken
	 * https://tools.ietf.org/html/rfc7662#section-2.2
	 */
	ClaimActive = "active"
	ClaimScope = "scope"
	ClaimClientId = "client_id"
	ClaimUsername = "username"
	ClaimTokenType = "token_type"
	//Claim = ""

	/**
	 * NFV Additions - Legacy
	 */
	ClaimLegacyTenantId = "tenantId"
	ClaimLegacyFirstName = "firstName"
	ClaimLegacyLastName = "lastName"
	ClaimLegacyUsername = "user_name"
)
