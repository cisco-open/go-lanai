package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_util"
	"encoding/xml"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

// SPMetadataMiddleware
// A SAML service provider should be able to work with multiple identity providers.
// Because the saml package assumes a service provider is configured with one idp only,
// we use the internal field to store information about this service provider,
// and we will create new saml.ServiceProvider struct for each new idp connection when its needed.
type SPMetadataMiddleware struct {
	// using value instead of pointer here because we need to copy it when connecting to specific idps.
	// the methods on saml.ServiceProvider are actually pointer receivers. golang will implicitly use
	// the pointers to these value as receivers
	internal      saml.ServiceProvider
	idpManager    idp.IdentityProviderManager
	clientManager *CacheableIdpClientManager
}

// MetadataHandlerFunc endpoint that provide SP's metadata
func (m *SPMetadataMiddleware) MetadataHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		//do this because the refresh metadata middleware is conditional,
		//but the metadata endpoint is not conditional
		m.refreshMetadata(c)

		descriptor := m.internal.Metadata()
		var mergedAcs []saml.IndexedEndpoint
		var mergedSlo []saml.Endpoint
		//we only provide ACS and SLO for the domains we configured
		for i, delegate := range m.clientManager.GetAllClients() {
			// ACS
			delegateDescriptor := delegate.Metadata().SPSSODescriptors[0]
			delegateAcs := delegateDescriptor.AssertionConsumerServices[0]
			delegateAcs.Index = i
			mergedAcs = append(mergedAcs, delegateAcs)

			// SLO
			delegateSlo := delegateDescriptor.SingleLogoutServices
			mergedSlo = append(mergedSlo, delegateSlo...)
		}
		descriptor.SPSSODescriptors[0].AssertionConsumerServices = mergedAcs
		descriptor.SPSSODescriptors[0].SingleLogoutServices = mergedSlo

		w := c.Writer
		buf, _ := xml.MarshalIndent(descriptor, "", "  ")
		w.Header().Set("Content-LoggerType", "application/samlmetadata+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=metadata.xml")
		_, _ = w.Write(buf)
	}
}

// RefreshMetadataHandler MW that responsible to refresh IDP's metadata whenever SAML Login/Logout related endpoint is called
func (m *SPMetadataMiddleware) RefreshMetadataHandler() gin.HandlerFunc {
	return m.refreshMetadata
}

// cache that are populated by the refresh metadata middleware instead of populated dynamically on commence
// because in a multi-instance micro service deployment, the auth request and auth response can occur on
// different instance
func (m *SPMetadataMiddleware) refreshMetadata(c *gin.Context) {
	idpDetails := m.idpManager.GetIdentityProvidersWithFlow(c.Request.Context(), idp.ExternalIdpSAML)
	var samlIdpDetails []SamlIdentityProvider
	for _, i := range idpDetails {
		if s, ok := i.(SamlIdentityProvider); ok {
			samlIdpDetails = append(samlIdpDetails, s)
		}
	}
	m.clientManager.RefreshCache(c, samlIdpDetails)
}

// resolveBinding find first supported binding using given binding location extractor
func (m *SPMetadataMiddleware) resolveBinding(supported []string, extractor func(string) string) (location, binding string) {
	for _, b := range supported {
		location = extractor(b)
		if location != "" {
			binding = b
			return
		}
	}
	return "", ""
}

// bindableSamlRequest abstracted interface that both saml.AuthnRequest and FixedLogoutRequest implements
type bindableSamlRequest interface {
	Redirect(relayState string, sp *saml.ServiceProvider) (*url.URL, error)
	Post(relayState string) []byte
}

func (m *SPMetadataMiddleware) redirectBindingExecutor(req bindableSamlRequest, relayState string, sp *saml.ServiceProvider) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		redirectURL, e := req.Redirect(relayState, sp)
		if e != nil {
			return e
		}
		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
		_, _ = w.Write(nil)
		return nil
	}
}

func (m *SPMetadataMiddleware) postBindingExecutor(req bindableSamlRequest, relayState string) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		data := req.Post(relayState)
		return saml_util.WritePostBindingForm(data, w)
	}
}


