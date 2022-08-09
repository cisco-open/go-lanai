package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"encoding/gob"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	SLOInitiated SLOState = 1 << iota
	SLOFullyCompleted
	SLOPartiallyCompleted
	SLOCompleted
)

type SLOState int

func (s SLOState) Is(mask SLOState) bool {
	return s ^ mask != 0
}

const (
	kDetailsSLOState = "SP.SLOState"
)

func init() {
	gob.Register(SLOState(0))
}

type SPLogoutMiddleware struct {
	SPMetadataMiddleware
	// TODO clean up this

	// supported SLO bindings, can be saml.HTTPPostBinding or saml.HTTPRedirectBinding. Order indicates preference
	bindings           []string
	successHandler     security.AuthenticationSuccessHandler
	fallbackEntryPoint security.AuthenticationEntryPoint
	//requestTracker samlsp.RequestTracker
	//authenticator security.Authenticator
}

func NewLogoutMiddleware(sp saml.ServiceProvider,
	idpManager idp.IdentityProviderManager,
	clientManager *CacheableIdpClientManager,
	successHandler security.AuthenticationSuccessHandler,
	errorPath string) *SPLogoutMiddleware {

	// TODO clean up this
	return &SPLogoutMiddleware{
		SPMetadataMiddleware: SPMetadataMiddleware{
			internal:      sp,
			idpManager:    idpManager,
			clientManager: clientManager,
		},
		bindings:           []string{saml.HTTPPostBinding, saml.HTTPRedirectBinding},
		successHandler:     successHandler,
		fallbackEntryPoint: redirect.NewRedirectWithURL(errorPath),
	}
}

// MakeSingleLogoutRequest initiate SLO at IdP by sending logout request with supported binding
func (sp *SPLogoutMiddleware) MakeSingleLogoutRequest(ctx context.Context, r *http.Request, w http.ResponseWriter) error {
	ep := redirect.NewRedirectWithRelativePath("/v2/logout/saml/slo/dummy", false)
	ep.Commence(ctx, r, w, nil)
	// TODO
	return nil
}

// LogoutRequestHandlerFunc returns the handler function that handles incoming LogoutRequest sent by IdP.
// This is used to handle IdP initiated SLO
// We need to initiate our internal logout process if this SLO process is not initiated by us
func (sp *SPLogoutMiddleware) LogoutRequestHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		// TODO we may need to initiate our internal logout process if this SLO process is not initiated by us
	}
}

// LogoutResponseHandlerFunc returns the handler function that handles LogoutResponse sent by IdP.
// This is used to handle response of SP initiated SLO, if it's initiated by us.
// We need to continue our internal logout process
func (sp *SPLogoutMiddleware) LogoutResponseHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		// TODO
		sp.updateSLOState(gc, func(current SLOState) SLOState {
			return current | SLOFullyCompleted | SLOCompleted
		})
		sp.handleSuccess(gc)
	}
}

// Commence implements security.AuthenticationEntryPoint. It's used when SP initiated SLO is required
func (sp *SPLogoutMiddleware) Commence(ctx context.Context, r *http.Request, w http.ResponseWriter, _ error) {
	if e := sp.MakeSingleLogoutRequest(ctx, r, w); e != nil {
		sp.fallbackEntryPoint.Commence(ctx, r, w, e)
		return
	}

	sp.updateSLOState(ctx, func(current SLOState) SLOState {
		return current | SLOInitiated
	})
}

func (sp *SPLogoutMiddleware) handleSuccess(gc *gin.Context) {
	auth := security.Get(gc)
	sp.successHandler.HandleAuthenticationSuccess(gc, gc.Request, gc.Writer, auth, auth)
	if gc.Writer.Written() {
		gc.Abort()
	}
}

func (sp *SPLogoutMiddleware) currentAuthDetails(ctx context.Context) map[string]interface{} {
	auth := security.Get(ctx)
	switch m := auth.Details().(type) {
	case map[string]interface{}:
		return m
	default:
		return nil
	}
}

func (sp *SPLogoutMiddleware) currentSLOState(ctx context.Context) SLOState {
	details := sp.currentAuthDetails(ctx)
	if details == nil {
		return 0
	}
	state, _ := details[kDetailsSLOState].(SLOState)
	return state
}

func (sp *SPLogoutMiddleware) updateSLOState(ctx context.Context, updater func(current SLOState) SLOState) {
	details := sp.currentAuthDetails(ctx)
	if details == nil {
		return
	}
	state, _ := details[kDetailsSLOState].(SLOState)
	details[kDetailsSLOState] = updater(state)
}

// DummySLOHandlerFunc mimic SLO flow for dev purpose, should be removed
func (sp *SPLogoutMiddleware) DummySLOHandlerFunc() gin.HandlerFunc {
	// TODO
	const PostHTMLTmpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Dummy Logout</title>
</head>
<body>
<form action="/auth/v2/logout/saml/slo/callback" method="get">
    <input type="submit" value="Proceed">
</form>
</body>
</html>
`
	return func(gc *gin.Context) {
		gc.Writer.Header().Set("Content-Type", "text/html")
		gc.Writer.WriteHeader(http.StatusOK)
		_, _ = gc.Writer.WriteString(PostHTMLTmpl)
	}
}
