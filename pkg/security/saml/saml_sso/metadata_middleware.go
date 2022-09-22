package saml_auth

import (
	"crypto"
	"crypto/x509"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/xml"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"sort"
)

type Options struct {
	Key                    crypto.PrivateKey
	Cert                   *x509.Certificate
	EntityIdUrl            url.URL
	SsoUrl                 url.URL
	SloUrl                 url.URL
	SigningMethod          string
	serviceProviderManager saml_auth_ctx.SamlClientStore
}

type MetadataMiddleware struct {
	samlClientStore   saml_auth_ctx.SamlClientStore // used to load the saml clients
	spMetadataManager *SpMetadataManager            // manages the resolved service provider metadata
	idp               *saml.IdentityProvider
}

func NewMetadataMiddleware(opts *Options, samlClientStore saml_auth_ctx.SamlClientStore) *MetadataMiddleware {

	spDescriptorManager := &SpMetadataManager{
		cache:      make(map[string]*saml.EntityDescriptor),
		processed:  make(map[string]SamlSpDetails),
		httpClient: http.DefaultClient,
	}

	idp := &saml.IdentityProvider{
		Key:         opts.Key,
		Logger:      newLoggerAdaptor(logger),
		Certificate: opts.Cert,
		//since we have our own middleware implementation, this value here only serves the purpose of defining the entity id.
		MetadataURL:     opts.EntityIdUrl,
		SSOURL:          opts.SsoUrl,
		LogoutURL:       opts.SloUrl,
		SignatureMethod: opts.SigningMethod,
	}

	mw := &MetadataMiddleware{
		idp:               idp,
		samlClientStore:   samlClientStore,
		spMetadataManager: spDescriptorManager,
	}
	return mw
}

func (mw *MetadataMiddleware) RefreshMetadataHandler(condition web.RequestMatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		if matches, err := condition.MatchesWithContext(c.Request.Context(), c.Request); !matches || err != nil {
			return
		}

		if clients, e := mw.samlClientStore.GetAllSamlClient(c.Request.Context()); e == nil {
			mw.spMetadataManager.RefreshCache(c, clients)
		}
	}
}

func (mw *MetadataMiddleware) MetadataHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		metadata := mw.idp.Metadata()
		sort.SliceStable(metadata.IDPSSODescriptors[0].SingleSignOnServices, func(i, j int) bool {
			return metadata.IDPSSODescriptors[0].SingleSignOnServices[i].Binding < metadata.IDPSSODescriptors[0].SingleSignOnServices[j].Binding
		})

		var t = true
		//We always want the authentication request to be signed
		//But because this is not supported by the saml package, we set it here explicitly
		metadata.IDPSSODescriptors[0].WantAuthnRequestsSigned = &t
		w := c.Writer
		buf, _ := xml.MarshalIndent(metadata, "", "  ")
		w.Header().Set("Content-Type", "application/samlmetadata+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=metadata.xml")
		_, _ = w.Write(buf)
	}
}
