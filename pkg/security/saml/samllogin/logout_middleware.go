package samllogin

import (
	"bytes"
	"compress/flate"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/base64"
	"encoding/gob"
	"encoding/xml"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
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
	return s&mask != 0
}

const (
	kDetailsSLOState = "SP.SLOState"
)

func init() {
	gob.Register(SLOState(0))
}

type SPLogoutMiddleware struct {
	SPMetadataMiddleware
	bindings           []string // supported SLO bindings, can be saml.HTTPPostBinding or saml.HTTPRedirectBinding. Order indicates preference
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
		bindings:           []string{saml.HTTPRedirectBinding, saml.HTTPPostBinding},
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
	location, binding := m.resolveBinding(m.bindings, client.GetSLOBindingLocation)
	if location == "" {
		return security.NewExternalSamlAuthenticationError("idp does not have supported SLO bindings.")
	}

	// create and send SLO request.
	nameId, format := m.resolveNameId(ctx)
	// Note 1: MakeLogoutRequest doesn't handle Redirect properly as of crewjam/saml, we wrap it with a temporary fix
	// Note 2: SLO specs don't requires RelayState
	sloReq, e := MakeFixedLogoutRequest(client, location, nameId)
	if e != nil {
		return security.NewExternalSamlAuthenticationError("cannot make SLO request to binding location", e)
	}
	sloReq.NameID.Format = format

	switch binding {
	case saml.HTTPRedirectBinding:
		if e := m.redirectBindingExecutor(sloReq, "", client)(w, r); e != nil {
			return security.NewExternalSamlAuthenticationError("cannot make SLO request with HTTP redirect binding", e)
		}
	case saml.HTTPPostBinding:
		if e := m.postBindingExecutor(sloReq, "")(w, r); e != nil {
			return security.NewExternalSamlAuthenticationError("cannot post SLO request", e)
		}
	}
	return nil
}

// LogoutRequestHandlerFunc returns the handler function that handles incoming LogoutRequest sent by IdP.
// This is used to handle IdP initiated SLO
// We need to initiate our internal logout process if this SLO process is not initiated by us
func (m *SPLogoutMiddleware) LogoutRequestHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		// TODO Handle Logout Request for IDP-initiated SLO
		body, e := ioutil.ReadAll(gc.Request.Body)
		logger.WithContext(gc).Infof("LogoutRequestHandlerFunc: [%v]%s", e, body)
		return
	}
}

// LogoutResponseHandlerFunc returns the handler function that handles LogoutResponse sent by IdP.
// This is used to handle response of SP initiated SLO, if it's initiated by us.
// We need to continue our internal logout process
func (m *SPLogoutMiddleware) LogoutResponseHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		// TODO Handle Logout Request for IDP-initiated SLO
		var encoded string
		var isRedirect bool
		if encoded, isRedirect = gc.GetQuery("SAMLResponse"); len(encoded) == 0 {
			encoded = gc.PostForm("SAMLResponse")
		}

		decoded, e := base64.StdEncoding.DecodeString(encoded)
		if e != nil {
			m.handleError(gc, security.NewExternalSamlAuthenticationError("cannot parse logout response body", e))
			return
		}

		// do some validation first before we decrypt
		resp := saml.LogoutResponse{}
		if e := xml.Unmarshal(decoded, &resp); e != nil {
			m.handleError(gc, security.NewExternalSamlAuthenticationError("Error unmarshalling SAMLResponse as xml", e))
			return
		}

		client, ok := m.clientManager.GetClientByEntityId(resp.Issuer.Value)
		if !ok {
			m.handleError(gc, security.NewExternalSamlAuthenticationError("cannot find idp metadata corresponding for logout response"))
			return
		}

		// perform validate, handle if success
		if isRedirect {
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
}

// Commence implements security.AuthenticationEntryPoint. It's used when SP initiated SLO is required
func (m *SPLogoutMiddleware) Commence(ctx context.Context, r *http.Request, w http.ResponseWriter, _ error) {
	if e := m.MakeSingleLogoutRequest(ctx, r, w); e != nil {
		gc := web.GinContext(ctx)
		m.handleError(gc, e)
		return
	}

	m.updateSLOState(ctx, func(current SLOState) SLOState {
		return current | SLOInitiated
	})
}

func (m *SPLogoutMiddleware) resolveIdpClient(ctx context.Context) (*saml.ServiceProvider, error) {
	var entityId string
	auth := security.Get(ctx)
	if samlAuth, ok := auth.(*samlAssertionAuthentication); ok {
		entityId = samlAuth.Assertion.Issuer.Value
	}
	if sp, ok := m.clientManager.GetClientByEntityId(entityId); ok {
		return sp, nil
	}
	return nil, security.NewExternalSamlAuthenticationError("Unable to initiate SLO as SP: unknown SAML Issuer")
}

func (m *SPLogoutMiddleware) resolveNameId(ctx context.Context) (nameId, format string) {
	auth := security.Get(ctx)
	if samlAuth, ok := auth.(*samlAssertionAuthentication); ok &&
		samlAuth.Assertion != nil && samlAuth.Assertion.Subject != nil && samlAuth.Assertion.Subject.NameID != nil {
		nameId = samlAuth.Assertion.Subject.NameID.Value
		format = samlAuth.Assertion.Subject.NameID.Format
		//format = string(saml.EmailAddressNameIDFormat)
	}
	return
}

func (m *SPLogoutMiddleware) handleSuccess(gc *gin.Context) {
	m.updateSLOState(gc, func(current SLOState) SLOState {
		return current | SLOCompletedFully
	})
	auth := security.Get(gc)
	m.successHandler.HandleAuthenticationSuccess(gc, gc.Request, gc.Writer, auth, auth)
	if gc.Writer.Written() {
		gc.Abort()
	}
}

func (m *SPLogoutMiddleware) handleError(gc *gin.Context, e error) {
	logger.WithContext(gc).Infof("SAML Single Logout failed with error: %v", e)
	m.updateSLOState(gc, func(current SLOState) SLOState {
		return current | SLOFailed
	})
	// We always let logout continues
	auth := security.Get(gc)
	m.successHandler.HandleAuthenticationSuccess(gc, gc.Request, gc.Writer, auth, auth)
	if gc.Writer.Written() {
		gc.Abort()
	}
}

func (m *SPLogoutMiddleware) currentAuthDetails(ctx context.Context) map[string]interface{} {
	auth := security.Get(ctx)
	switch m := auth.Details().(type) {
	case map[string]interface{}:
		return m
	default:
		return nil
	}
}

func (m *SPLogoutMiddleware) currentSLOState(ctx context.Context) SLOState {
	details := m.currentAuthDetails(ctx)
	if details == nil {
		return 0
	}
	state, _ := details[kDetailsSLOState].(SLOState)
	return state
}

func (m *SPLogoutMiddleware) updateSLOState(ctx context.Context, updater func(current SLOState) SLOState) {
	details := m.currentAuthDetails(ctx)
	if details == nil {
		return
	}
	state, _ := details[kDetailsSLOState].(SLOState)
	details[kDetailsSLOState] = updater(state)
}

/***********************
	Workaround
 ***********************/

type FixedLogoutRequest struct {
	saml.LogoutRequest
}

func MakeFixedLogoutRequest(sp *saml.ServiceProvider, idpURL, nameID string) (*FixedLogoutRequest, error) {
	req, e := sp.MakeLogoutRequest(idpURL, nameID)
	if e != nil {
		return nil, e
	}
	return &FixedLogoutRequest{*req}, nil
}

// Redirect this is copied from saml.AuthnRequest.Redirect.
// As of crewjam/saml 0.4.8, AuthnRequest's Redirect is fixed for properly setting Signature in redirect URL:
// 	https://github.com/crewjam/saml/pull/339
// However, saml.LogoutRequest.Redirect is not fixed. We need to do that by ourselves
// TODO revisit this part later when newer crewjam/saml library become available
func (req *FixedLogoutRequest) Redirect(relayState string, sp *saml.ServiceProvider) (*url.URL, error) {
	w := &bytes.Buffer{}
	w1 := base64.NewEncoder(base64.StdEncoding, w)
	defer func() {}()
	w2, _ := flate.NewWriter(w1, 9)
	doc := etree.NewDocument()
	doc.SetRoot(req.Element())
	if _, err := doc.WriteTo(w2); err != nil {
		return nil, err
	}
	_ = w2.Close()
	_ = w1.Close()

	rv, _ := url.Parse(req.Destination)
	// We can't depend on Query().set() as order matters for signing
	query := rv.RawQuery
	if len(query) > 0 {
		query += "&SAMLRequest=" + url.QueryEscape(string(w.Bytes()))
	} else {
		query += "SAMLRequest=" + url.QueryEscape(string(w.Bytes()))
	}

	if relayState != "" {
		query += "&RelayState=" + relayState
	}
	if len(sp.SignatureMethod) > 0 {
		query += "&SigAlg=" + url.QueryEscape(sp.SignatureMethod)
		signingContext, err := saml.GetSigningContext(sp)

		if err != nil {
			return nil, err
		}

		sig, err := signingContext.SignString(query)
		if err != nil {
			return nil, err
		}
		query += "&Signature=" + url.QueryEscape(base64.StdEncoding.EncodeToString(sig))
	}

	rv.RawQuery = query

	return rv, nil
}
