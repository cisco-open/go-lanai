package samllogin

import (
	"crypto/rsa"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"encoding/xml"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"net/url"
)

type SPOptions struct {
	URL               url.URL
	Key               *rsa.PrivateKey
	Certificate       *x509.Certificate
	Intermediates     []*x509.Certificate
	ACSPath           string
	MetadataPath      string
	SLOPath           string
	AllowIDPInitiated bool
	SignRequest       bool
	ForceAuthn        bool
	NameIdFormat      string
}

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

func (sp *SPMetadataMiddleware) MetadataHandlerFunc() gin.HandlerFunc {
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

// RefreshMetadataHandler
func (sp *SPMetadataMiddleware) RefreshMetadataHandler() gin.HandlerFunc {
	return sp.refreshMetadata
}

// cache that are populated by the refresh metadata middleware instead of populated dynamically on commence
// because in a multi-instance micro service deployment, the auth request and auth response can occur on
// different instance
func (sp *SPMetadataMiddleware) refreshMetadata(c *gin.Context) {
	idpDetails := sp.idpManager.GetIdentityProvidersWithFlow(c.Request.Context(), idp.ExternalIdpSAML)
	var samlIdpDetails []SamlIdentityProvider
	for _, i := range idpDetails {
		if s, ok := i.(SamlIdentityProvider); ok {
			samlIdpDetails = append(samlIdpDetails, s)
		}
	}
	sp.clientManager.RefreshCache(samlIdpDetails)
}
