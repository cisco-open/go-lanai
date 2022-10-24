package ittest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"fmt"
	"net/http"
	"net/url"
)

const CheckTokenPath = `/v2/check_token`

// RemoteTokenStoreReader implements oauth2.TokenStoreReader that leverage /check_token endpoint to load authentication
// Note: this implementation is not mocks. With proper refactoring, it can be potentially used in production
type RemoteTokenStoreReader struct {
	// JwtDecoder optional, when provided, token signature is pre-checked before sent to remote auth service
	JwtDecoder jwt.JwtDecoder
	// SkipRemoteCheck, if set to true, skip the remote check when JwtDecoder is provided and context details is not required.
	SkipRemoteCheck bool
	// HttpClient httpclient.Client to use for remote token check
	HttpClient httpclient.Client

	ClientId     string
	ClientSecret string
}

type RemoteTokenStoreOptions func(opt *RemoteTokenStoreOption)
type RemoteTokenStoreOption struct {
	JwtDecoder       jwt.JwtDecoder
	HttpClient       httpclient.Client
	HttpClientConfig *httpclient.ClientConfig
	BaseUrl          string
	ServiceName      string // auth service's name for LB
	ClientId         string
	ClientSecret     string
	SkipRemoteCheck  bool
}

func NewRemoteTokenStoreReader(opts ...RemoteTokenStoreOptions) oauth2.TokenStoreReader {
	var opt RemoteTokenStoreOption
	for _, fn := range opts {
		fn(&opt)
	}

	var client httpclient.Client
	var e error
	if opt.BaseUrl != "" {
		client, e = opt.HttpClient.WithBaseUrl(opt.BaseUrl)
	} else {
		client, e = opt.HttpClient.WithService(opt.ServiceName)
	}
	if e != nil {
		panic(e)
	}
	if opt.HttpClientConfig != nil {
		client = client.WithConfig(opt.HttpClientConfig)
	}
	return &RemoteTokenStoreReader{
		JwtDecoder:      opt.JwtDecoder,
		SkipRemoteCheck: opt.SkipRemoteCheck,
		HttpClient:      client.WithConfig(opt.HttpClientConfig),
		ClientId:        opt.ClientId,
		ClientSecret:    opt.ClientSecret,
	}
}

func (r *RemoteTokenStoreReader) ReadAuthentication(ctx context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	switch hint {
	case oauth2.TokenHintAccessToken:
		return r.readAuthenticationFromAccessToken(ctx, tokenValue)
	default:
		return nil, oauth2.NewUnsupportedTokenTypeError(fmt.Sprintf("token type [%s] is not supported", hint.String()))
	}
}

func (r *RemoteTokenStoreReader) ReadAccessToken(ctx context.Context, value string) (oauth2.AccessToken, error) {
	token, e := r.readAccessToken(ctx, value, nil)
	if e != nil {
		return nil, oauth2.NewInvalidAccessTokenError("token is invalid", e)
	}
	return token, nil
}

func (r *RemoteTokenStoreReader) ReadRefreshToken(ctx context.Context, value string) (oauth2.RefreshToken, error) {
	token, e := r.parseRefreshToken(ctx, value)
	switch {
	case e != nil:
		return nil, oauth2.NewInvalidGrantError("refresh token is invalid", e)
	case token.WillExpire() && token.Expired():
		return nil, oauth2.NewInvalidGrantError("refresh token is expired")
	}
	return token, nil
}

func (r *RemoteTokenStoreReader) readAccessToken(ctx context.Context, value string, detailedClaims *misc.CheckTokenClaims) (*oauth2.DefaultAccessToken, error) {
	var basicClaims oauth2.BasicClaims
	// pre-check signature if possible
	if r.JwtDecoder != nil {
		if e := r.JwtDecoder.DecodeWithClaims(ctx, value, &basicClaims); e != nil {
			return nil, e
		}
	}

	requireDetails := detailedClaims != nil || len(basicClaims.Id) == 0
	// Note, we only skip revocation check when we have token claims, details is not required and SkipRemoteCheck is true
	if r.SkipRemoteCheck && !requireDetails {
		return r.createAccessToken(&basicClaims, value), nil
	}

	// perform remote check
	if requireDetails && detailedClaims == nil {
		detailedClaims = &misc.CheckTokenClaims{}
	}
	if e := r.remoteAccessTokenCheck(ctx, value, detailedClaims); e != nil {
		return nil, e
	}
	return r.createAccessToken(&detailedClaims.BasicClaims, value), nil
}

func (r *RemoteTokenStoreReader) parseRefreshToken(_ context.Context, _ string) (*oauth2.DefaultRefreshToken, error) {
	return nil, fmt.Errorf("remote refresh token validation is not supported")
}

func (r *RemoteTokenStoreReader) readAuthenticationFromAccessToken(ctx context.Context, tokenValue string) (oauth2.Authentication, error) {
	// parse JWT token
	var claims misc.CheckTokenClaims
	token, e := r.readAccessToken(ctx, tokenValue, &claims)
	if e != nil {
		return nil, e
	}

	// load context details
	details := r.createSecurityDetails(&claims)
	if e != nil {
		return nil, oauth2.NewInvalidAccessTokenError("token unknown", e)
	}

	// reconstruct request
	request := r.createOAuth2Request(&claims, details)

	// reconstruct user auth if available
	var userAuth security.Authentication
	if claims.Subject != "" {
		userAuth = r.createUserAuthentication(&claims, details)
	}

	return oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = request
		opt.UserAuth = userAuth
		opt.Token = token
		opt.Details = details
	}), nil
}

/*****************
	Helpers
 *****************/

func (r *RemoteTokenStoreReader) remoteAccessTokenCheck(ctx context.Context, value string, dest *misc.CheckTokenClaims) error {
	form := url.Values{
		"token": []string{value},
		"token_type_hint": []string{"access_token"},
		"no_details": []string{fmt.Sprintf("%v", dest == nil)},
	}
	req := httpclient.NewRequest(CheckTokenPath, http.MethodPost,
		httpclient.WithUrlEncodedBody(form),
		httpclient.WithBasicAuth(r.ClientId, r.ClientSecret),
	)

	claims := dest
	if dest == nil {
		claims = &misc.CheckTokenClaims{}
	}
	_, e := r.HttpClient.Execute(ctx, req, httpclient.JsonBody(claims))
	if e != nil {
		return e
	}

	if claims.Active == nil || !*claims.Active {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (r *RemoteTokenStoreReader) createAccessToken(claims *oauth2.BasicClaims, value string) *oauth2.DefaultAccessToken {
	token := oauth2.NewDefaultAccessToken(value)
	token.SetExpireTime(claims.ExpiresAt)
	token.SetIssueTime(claims.IssuedAt)
	token.SetScopes(claims.Scopes.Copy())
	token.SetClaims(claims)
	return token
}

func (r *RemoteTokenStoreReader) createSecurityDetails(claims *misc.CheckTokenClaims) security.ContextDetails {
	return sectest.NewMockedSecurityDetails(func(d *sectest.SecurityDetailsMock) {
		*d = sectest.SecurityDetailsMock{
			Username:                 claims.Username,
			UserId:                   claims.UserId,
			TenantExternalId:         claims.TenantExternalId,
			TenantId:                 claims.TenantId,
			ProviderName:             claims.ProviderName,
			ProviderId:               claims.ProviderId,
			ProviderDisplayName:      claims.ProviderDisplayName,
			ProviderDescription:      claims.ProviderDescription,
			ProviderEmail:            claims.ProviderEmail,
			ProviderNotificationType: claims.ProviderNotificationType,
			Exp:                      claims.ExpiresAt,
			Iss:                      claims.IssuedAt,
			Permissions:              claims.Permissions,
			Tenants:                  claims.AssignedTenants,
			OrigUsername:             claims.OrigUsername,
			UserFirstName:            claims.FirstName,
			UserLastName:             claims.LastName,
			KVs:                      map[string]interface{}{},
		}
	})
}

func (r *RemoteTokenStoreReader) createOAuth2Request(claims *misc.CheckTokenClaims, details security.ContextDetails) oauth2.OAuth2Request {
	clientId := claims.ClientId
	if clientId == "" && claims.Audience != nil && len(claims.Audience) != 0 {
		clientId = utils.StringSet(claims.Audience).Values()[0]
	}

	params := map[string]string{}
	reqParams, _ := details.Value(oauth2.DetailsKeyRequestParams)
	if m, ok := reqParams.(map[string]interface{}); ok {
		for k, v := range m {
			switch s := v.(type) {
			case string:
				params[k] = s
			}
		}
	}

	ext := claims.Values()
	reqExt, _ := details.Value(oauth2.DetailsKeyRequestExt)
	if m, ok := reqExt.(map[string]interface{}); ok {
		for k, v := range m {
			ext[k] = v
		}
	}

	return oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
		opt.Parameters = params
		opt.ClientId = clientId
		opt.Scopes = claims.Scopes
		opt.Approved = true
		opt.Extensions = ext
		//opt.GrantType =
		//opt.RedirectUri =
		//opt.ResponseTypes =
	})
}

func (r *RemoteTokenStoreReader) createUserAuthentication(claims *misc.CheckTokenClaims, details security.ContextDetails) security.Authentication {
	permissions := map[string]interface{}{}
	for k := range details.Permissions() {
		permissions[k] = true
	}

	return oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
		opt.Principal = claims.Subject
		opt.Permissions = permissions
		opt.State = security.StateAuthenticated
		opt.Details = map[string]interface{}{}
		if claims != nil {
			opt.Details = claims.Values()
		}
	})
}
