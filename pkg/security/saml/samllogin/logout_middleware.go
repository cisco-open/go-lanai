package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	samlutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"errors"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	SLOInitiated SLOState = 1 << iota
	SLOCompletedFully
	SLOCompletedPartially
	SLOFailed
	SLOCompleted = SLOCompletedFully | SLOCompletedPartially | SLOFailed
)

type SLOState int

func (s SLOState) Is(mask SLOState) bool {
	return s&mask != 0 || mask == 0 && s == 0
}

const (
	kDetailsSLOState = "SP.SLOState"
)

func init() {
	gob.Register(SLOState(0))
}

type SPLogoutMiddleware struct {
	SPMetadataMiddleware
	successHandler     security.AuthenticationSuccessHandler
}

func NewLogoutMiddleware(sp saml.ServiceProvider,
	idpManager idp.IdentityProviderManager,
	clientManager *CacheableIdpClientManager,
	successHandler security.AuthenticationSuccessHandler) *SPLogoutMiddleware {

	return &SPLogoutMiddleware{
		SPMetadataMiddleware: SPMetadataMiddleware{
			internal:      sp,
			idpManager:    idpManager,
			clientManager: clientManager,
		},
		successHandler:     successHandler,
	}
}

// MakeSingleLogoutRequest initiate SLO at IdP by sending logout request with supported binding
func (m *SPLogoutMiddleware) MakeSingleLogoutRequest(ctx context.Context, r *http.Request, w http.ResponseWriter) error {
	// resolve SP client
	client, e := m.resolveIdpClient(ctx)
	if e != nil {
		return e
	}

	// resolve binding
	location, binding := m.resolveBinding(client.GetSLOBindingLocation)
	if location == "" {
		return security.NewExternalSamlAuthenticationError("idp does not have supported SLO bindings.")
	}

	// create and send SLO request.
	nameId, format := m.resolveNameId(ctx)
	// Note 1: MakeLogoutRequest doesn't handle Redirect properly as of crewjam/saml, we wrap it with a temporary fix
	// Note 2: SLO specs don't requires RelayState
	sloReq, e := samlutils.NewFixedLogoutRequest(client, location, nameId)
	if e != nil {
		return security.NewExternalSamlAuthenticationError("cannot make SLO request to binding location", e)
	}
	sloReq.NameID.Format = format

	// re-sign the request since we changed the format
	sloReq.Signature = nil
	if e := client.SignLogoutRequest(&sloReq.LogoutRequest); e != nil {
		return security.NewExternalSamlAuthenticationError("cannot sign SLO request", e)
	}

	switch binding {
	case saml.HTTPRedirectBinding:
		if e := m.redirectBindingExecutor(sloReq, "", client)(w, r); e != nil {
			return security.NewExternalSamlAuthenticationError("cannot send SLO request with HTTP redirect binding", e)
		}
	case saml.HTTPPostBinding:
		if e := m.postBindingExecutor(sloReq, "")(w, r); e != nil {
			return security.NewExternalSamlAuthenticationError("cannot send SLO request with HTTP post binding", e)
		}
	}
	return nil
}

// LogoutHandlerFunc returns the handler function that handles LogoutResponse/LogoutRequest sent by IdP.
// This is used to handle response of SP initiated SLO, if it's initiated by us.
// We need to continue our internal logout process
func (m *SPLogoutMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		var req saml.LogoutRequest
		var resp saml.LogoutResponse
		reqR := samlutils.ParseSAMLObject(gc, &req)
		respR := samlutils.ParseSAMLObject(gc, &resp)
		switch {
		case reqR.Err != nil && respR.Err != nil || reqR.Err == nil && respR.Err == nil:
			m.handleError(gc, security.NewExternalSamlAuthenticationError("Error reading SAMLRequest/SAMLResponse", reqR.Err, respR.Err))
			return
		case respR.Err == nil:
			m.handleLogoutResponse(gc, &resp, respR.Binding, respR.Encoded)
		case reqR.Err == nil:
			m.handleLogoutRequest(gc, &req, reqR.Binding, reqR.Encoded)
		}
	}
}

// Commence implements security.AuthenticationEntryPoint. It's used when SP initiated SLO is required
func (m *SPLogoutMiddleware) Commence(ctx context.Context, r *http.Request, w http.ResponseWriter, err error) {
	if !errors.Is(err, ErrSamlSloRequired) {
		return
	}

	logger.WithContext(ctx).Infof("trying to start SAML SP-Initiated SLO")
	if e := m.MakeSingleLogoutRequest(ctx, r, w); e != nil {
		m.handleError(ctx, e)
		return
	}

	updateSLOState(ctx, func(current SLOState) SLOState {
		return current | SLOInitiated
	})
}

func (m *SPLogoutMiddleware) handleLogoutResponse(gc *gin.Context, resp *saml.LogoutResponse, binding, encoded string) {
	client, ok := m.clientManager.GetClientByEntityId(resp.Issuer.Value)
	if !ok {
		m.handleError(gc, security.NewExternalSamlAuthenticationError("cannot find idp metadata corresponding for logout response"))
		return
	}

	// perform validate, handle if success
	var e error
	if binding == saml.HTTPRedirectBinding {
		e = client.ValidateLogoutResponseRedirect(encoded)
	} else {
		e = client.ValidateLogoutResponseForm(encoded)
	}
	if e == nil {
		m.handleSuccess(gc)
		return
	}

	// handle error
	m.handleError(gc, e)
}

func (m *SPLogoutMiddleware) handleLogoutRequest(gc *gin.Context, req *saml.LogoutRequest, binding, encoded string) {
	// TODO Handle Logout Request for IDP-initiated SLO
}

func (m *SPLogoutMiddleware) resolveIdpClient(ctx context.Context) (*saml.ServiceProvider, error) {
	var entityId string
	auth := security.Get(ctx)
	if samlAuth, ok := auth.(*samlAssertionAuthentication); ok {
		entityId = samlAuth.SamlAssertion.Issuer.Value
	}
	if sp, ok := m.clientManager.GetClientByEntityId(entityId); ok {
		return sp, nil
	}
	return nil, security.NewExternalSamlAuthenticationError("Unable to initiate SLO as SP: unknown SAML Issuer")
}

func (m *SPLogoutMiddleware) resolveNameId(ctx context.Context) (nameId, format string) {
	auth := security.Get(ctx)
	if samlAuth, ok := auth.(*samlAssertionAuthentication); ok &&
		samlAuth.SamlAssertion != nil && samlAuth.SamlAssertion.Subject != nil && samlAuth.SamlAssertion.Subject.NameID != nil {
		nameId = samlAuth.SamlAssertion.Subject.NameID.Value
		format = samlAuth.SamlAssertion.Subject.NameID.Format
		//format = string(saml.EmailAddressNameIDFormat)
	}
	return
}

func (m *SPLogoutMiddleware) handleSuccess(ctx context.Context) {
	updateSLOState(ctx, func(current SLOState) SLOState {
		return current | SLOCompletedFully
	})
	gc := web.GinContext(ctx)
	auth := security.Get(ctx)
	m.successHandler.HandleAuthenticationSuccess(ctx, gc.Request, gc.Writer, auth, auth)
	if gc.Writer.Written() {
		gc.Abort()
	}
}

func (m *SPLogoutMiddleware) handleError(ctx context.Context, e error) {
	logger.WithContext(ctx).Infof("SAML Single Logout failed with error: %v", e)
	updateSLOState(ctx, func(current SLOState) SLOState {
		return current | SLOFailed
	})
	// We always let logout continues
	gc := web.GinContext(ctx)
	auth := security.Get(ctx)
	m.successHandler.HandleAuthenticationSuccess(ctx, gc.Request, gc.Writer, auth, auth)
	if gc.Writer.Written() {
		gc.Abort()
	}
}

/***********************
	Helper Funcs
 ***********************/

func currentAuthDetails(ctx context.Context) map[string]interface{} {
	auth := security.Get(ctx)
	switch m := auth.Details().(type) {
	case map[string]interface{}:
		return m
	default:
		return nil
	}
}

func currentSLOState(ctx context.Context) SLOState {
	details := currentAuthDetails(ctx)
	if details == nil {
		return 0
	}
	state, _ := details[kDetailsSLOState].(SLOState)
	return state
}

func updateSLOState(ctx context.Context, updater func(current SLOState) SLOState) {
	details := currentAuthDetails(ctx)
	if details == nil {
		return
	}
	state, _ := details[kDetailsSLOState].(SLOState)
	details[kDetailsSLOState] = updater(state)
}
