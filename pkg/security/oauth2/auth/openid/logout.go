package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"net/http"
	"strings"
)

var ParameterRedirectUri = "post_logout_redirect_uri"
var ParameterIdTokenHint = "id_token_hint"

type SuccessOptions func(opt *SuccessOption)

type SuccessOption struct {
	ClientStore         oauth2.OAuth2ClientStore
	WhitelabelErrorPath string
}

type OidcSuccessHandler struct {
	clientStore oauth2.OAuth2ClientStore
	fallback    security.AuthenticationErrorHandler
}

func (o *OidcSuccessHandler) Order() int {
	return order.Highest
}

func NewOidcSuccessHandler(opts ...SuccessOptions) *OidcSuccessHandler {
	opt := SuccessOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OidcSuccessHandler{
		clientStore: opt.ClientStore,
		fallback:    redirect.NewRedirectWithURL(opt.WhitelabelErrorPath),
	}
	return &OidcSuccessHandler{}
}

func (o *OidcSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	if r.FormValue(ParameterRedirectUri) == "" {
		return
	}

	redirectUri := r.FormValue(ParameterRedirectUri)
	if redirectUri == "" {
		// as OIDC success handler, we only care about this redirect
		return
	}

	// since the corresponding logout handler already validated the logout request and the redirect uri, we just need to do the redirect.
	http.Redirect(rw, r, redirectUri, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

type HandlerOptions func(opt *HandlerOption)

type HandlerOption struct {
	Dec         jwt.JwtDecoder
	Issuer      security.Issuer
	ClientStore oauth2.OAuth2ClientStore
}

type OidcLogoutHandler struct {
	dec         jwt.JwtDecoder
	issuer      security.Issuer
	clientStore oauth2.OAuth2ClientStore
}

func NewOidcLogoutHandler(opts ...HandlerOptions) *OidcLogoutHandler {
	opt := HandlerOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OidcLogoutHandler{
		dec:         opt.Dec,
		issuer:      opt.Issuer,
		clientStore: opt.ClientStore,
	}
}

func (o *OidcLogoutHandler) Order() int {
	return order.Highest
}

func (o *OidcLogoutHandler) ShouldLogout(ctx context.Context, request *http.Request, writer http.ResponseWriter, authentication security.Authentication) error {
	switch request.Method {
	case http.MethodGet:
		fallthrough
	case http.MethodPost:
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		fallthrough
	default:
		return security.NewInternalError(fmt.Sprintf("unsupported http verb %v", request.Method))
	}

	redirectUri := request.FormValue(ParameterRedirectUri)
	if redirectUri == "" {
		return nil
	}

	idTokenValue := request.FormValue(ParameterIdTokenHint)
	if strings.TrimSpace(idTokenValue) == "" {
		return fmt.Errorf(`id token is required from parameter "%s"`, ParameterIdTokenHint)
	}

	claims, err := o.dec.Decode(ctx, idTokenValue)
	if err != nil {
		return security.NewInternalError("id token cannot be decoded", err)
	}

	iss := claims.Get(oauth2.ClaimIssuer)
	if iss != o.issuer.Identifier() {
		return security.NewInternalError("id token is not issued by this auth server")
	}

	sub := claims.Get(oauth2.ClaimSubject)
	username, err := security.GetUsername(authentication)

	if err != nil {
		return security.NewInternalError("Couldn't identify current session user", err)
	} else if sub != username {
		return security.NewInternalError("logout request rejected because id token is not from the current session's user.")
	}

	clientId := claims.Get(oauth2.ClaimAudience).(string)
	client, err := auth.LoadAndValidateClientId(ctx, clientId, o.clientStore)
	if err != nil {
		return security.NewInternalError(fmt.Sprintf("error loading client %s", clientId), err)
	}
	_, err = auth.ResolveRedirectUri(ctx, redirectUri, client)
	if err != nil {
		return security.NewInternalError(fmt.Sprintf("redirect url %s is not registered by client %s", redirectUri, clientId))
	}

	return nil
}

func (o *OidcLogoutHandler) HandleLogout(ctx context.Context, request *http.Request, writer http.ResponseWriter, authentication security.Authentication) error {
	//no op, because the default logout handler is sufficient (deleting the current session etc.)
	return nil
}
