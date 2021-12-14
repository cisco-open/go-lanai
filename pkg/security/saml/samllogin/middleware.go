package samllogin

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	netutil "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/net"
	"encoding/base64"
	"encoding/xml"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

// ServiceProviderMiddleware
/**
	A SAML service provider should be able to work with multiple identity providers.
	Because the saml package assumes a service provider is configured with one idp only,
	we use the internal field to store information about this service provider,
	and we will create new saml.ServiceProvider struct for each new idp connection when its needed.
 */
type ServiceProviderMiddleware struct {
	//using value instead of pointer here because we need to copy it when connecting to specific idps.
	// the methods on saml.ServiceProvider are actually pointer receivers. golang will implicitly use
	// the pointers to these value as receivers
	internal   saml.ServiceProvider
	idpManager idp.IdentityProviderManager

	// list of bindings, can be saml.HTTPPostBinding or saml.HTTPRedirectBinding
	// order indicates preference
	bindings       []string
	requestTracker samlsp.RequestTracker

	authenticator security.Authenticator
	successHandler security.AuthenticationSuccessHandler

	fallbackEntryPoint security.AuthenticationEntryPoint

	clientManager *CacheableIdpClientManager
}

type Options struct {
	URL               url.URL
	Key               *rsa.PrivateKey
	Certificate       *x509.Certificate
	Intermediates     []*x509.Certificate
	ACSPath	string
	MetadataPath string
	SLOPath string
	AllowIDPInitiated bool
	SignRequest       bool
	ForceAuthn        bool
}


func NewMiddleware(sp saml.ServiceProvider, tracker samlsp.RequestTracker,
	idpManager idp.IdentityProviderManager,
	clientManager *CacheableIdpClientManager,
	handler security.AuthenticationSuccessHandler, authenticator security.Authenticator,
	errorPath string) *ServiceProviderMiddleware {

	return &ServiceProviderMiddleware{
		internal:           sp,
		bindings:           []string{saml.HTTPPostBinding, saml.HTTPRedirectBinding},
		idpManager:         idpManager,
		clientManager:      clientManager,
		requestTracker:     tracker,
		successHandler:     handler,
		authenticator:      authenticator,
		fallbackEntryPoint: redirect.NewRedirectWithURL(errorPath),
	}
}

func (sp *ServiceProviderMiddleware) MetadataHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		//do this because the refresh metadata middleware is conditional,
		//but the metadata endpoint is not conditional
		sp.refreshMetadata(c)

		index := 0
		descriptor := sp.internal.Metadata()
		var mergedAcs []saml.IndexedEndpoint

		//we don't support single logout yet, so don't include this in metadata
		descriptor.SPSSODescriptors[0].SingleLogoutServices = nil

		//we only provide ACS for the domains we configured
		for _, delegate := range sp.clientManager.GetAllClients() {
			delegateDescriptor := delegate.Metadata().SPSSODescriptors[0]
			delegateAcs := delegateDescriptor.AssertionConsumerServices[0]
			delegateAcs.Index = index
			mergedAcs = append(mergedAcs, delegateAcs)
			index++
		}

		descriptor.SPSSODescriptors[0].AssertionConsumerServices = mergedAcs

		w := c.Writer
		buf, _ := xml.MarshalIndent(descriptor, "", "  ")
		w.Header().Set("Content-LoggerType", "application/samlmetadata+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=metadata.xml")
		_, _ = w.Write(buf)
	}
}

// MakeAuthenticationRequest Since we support multiple domains each with different IDP, the auth request specify which matching ACS should be
// used for IDP to call back.
func (sp *ServiceProviderMiddleware) MakeAuthenticationRequest(r *http.Request, w http.ResponseWriter) error {
	host := netutil.GetForwardedHostName(r)
	client, ok := sp.clientManager.GetClientByDomain(host)

	if !ok {
		logger.Debugf("cannot find idp for domain %s", host)
		return security.NewExternalSamlAuthenticationError("cannot find idp for this domain")
	}

	var bindingLocation string
	var binding string
	for _, b := range sp.bindings {
		bindingLocation = client.GetSSOBindingLocation(b)
		if bindingLocation != "" {
			binding = b
			break
		}
	}

	if bindingLocation == "" {
		return security.NewExternalSamlAuthenticationError("idp does not have supported bindings.")
	}

	authReq, err := client.MakeAuthenticationRequest(bindingLocation)

	if err != nil {
		return security.NewExternalSamlAuthenticationError("cannot make auth request to binding location", err)
	}

	relayState, err := sp.requestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		return security.NewExternalSamlAuthenticationError("cannot track saml auth request", err)
	}

	if binding == saml.HTTPRedirectBinding {
		redirectURL := authReq.Redirect(relayState)
		w.Header().Add("Location", redirectURL.String())
		w.WriteHeader(http.StatusFound)
	} else if binding == saml.HTTPPostBinding {
		//add a hash for the inline script generated by authReq.Post so that we know only
		//this inline script is executed.
		//this is to prevent the case of the html is injected by bad actors, although it's unlikely in our case
		w.Header().Add("Content-Security-Policy", ""+
			"default-src; "+
			"script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; "+ //this hash matches the inline script generated by authReq.Post
			"reflected-xss block; referrer no-referrer;")
		w.Header().Add("Content-type", "text/html")

		body := append([]byte(`<!DOCTYPE html><html><body>`), authReq.Post(relayState)...)
		body = append(body, []byte(`</body></html>`)...)
		_, err = w.Write(body)
		if err != nil {
			return security.NewExternalSamlAuthenticationError("cannot post auth request", err)
		}
	}
	return nil
}

func (sp *ServiceProviderMiddleware) ACSHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		err := r.ParseForm()
		if err != nil {
			sp.handleError(c, security.NewExternalSamlAuthenticationError("Can't parse request body", err))
			return
		}

		//Parse the response and get entityId
		rawResponseBuf, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLResponse"))
		if err != nil {
			sp.handleError(c, security.NewExternalSamlAuthenticationError("Error decoding (base64) SAMLResponse", err))
			return
		}

		// do some validation first before we decrypt
		resp := saml.Response{}
		if err := xml.Unmarshal(rawResponseBuf, &resp); err != nil {
			sp.handleError(c, security.NewExternalSamlAuthenticationError("Error unmarshalling SAMLResponse as xml", err))
			return
		}

		client, ok := sp.clientManager.GetClientByEntityId(resp.Issuer.Value)
		if !ok {
			sp.handleError(c, security.NewExternalSamlAuthenticationError("cannot find idp metadata corresponding for assertion"))
			return
		}

		var possibleRequestIDs []string
		if sp.internal.AllowIDPInitiated {
			possibleRequestIDs = append(possibleRequestIDs, "")
		}

		trackedRequests := sp.requestTracker.GetTrackedRequests(r)
		for _, tr := range trackedRequests {
			possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
		}

		assertion, err := client.ParseResponse(r, possibleRequestIDs)
		if err != nil {
			logger.Error("error processing assertion", "err", err)
			sp.handleError(c, security.NewExternalSamlAuthenticationError(err.Error(), err))
			return
		}

		candidate := &AssertionCandidate{
			Assertion: assertion,
		}
		auth, err := sp.authenticator.Authenticate(c, candidate)

		if err != nil {
			sp.handleError(c, security.NewExternalSamlAuthenticationError(err))
			return
		}

		before := security.Get(c)
		sp.handleSuccess(c, before, auth)
	}
}

//cache that are populated by the refresh metadata middleware instead of populated dynamically on commence
// because in a multi-instance micro service deployment, the auth request and auth response can occur on
// different instance

func (sp *ServiceProviderMiddleware) RefreshMetadataHandler() gin.HandlerFunc {
	return sp.refreshMetadata
}

func (sp *ServiceProviderMiddleware) refreshMetadata(c *gin.Context) {
	idpDetails := sp.idpManager.GetIdentityProvidersWithFlow(c.Request.Context(), idp.ExternalIdpSAML)
	var samlIdpDetails []SamlIdentityProvider
	for _, i := range idpDetails {
		if s, ok := i.(SamlIdentityProvider); ok {
			samlIdpDetails = append(samlIdpDetails, s)
		}
	}
	sp.clientManager.RefreshCache(samlIdpDetails)
}

func (sp *ServiceProviderMiddleware) Commence(c context.Context, r *http.Request, w http.ResponseWriter, _ error) {
	err := sp.MakeAuthenticationRequest(r, w)
	if err != nil {
		sp.fallbackEntryPoint.Commence(c, r, w, err)
	}
}

func (sp *ServiceProviderMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
	}
	sp.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	if c.Writer.Written() {
		c.Abort()
	}
}

func (sp *ServiceProviderMiddleware) handleError(c *gin.Context, err error) {
	if trackedRequestIndex := c.Request.Form.Get("RelayState"); trackedRequestIndex != "" {
		_ = sp.requestTracker.StopTrackingRequest(c.Writer, c.Request, trackedRequestIndex)
	}
	security.Clear(c)
	_ = c.Error(err)
	c.Abort()
}