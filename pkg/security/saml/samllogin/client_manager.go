package samllogin

import (
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_util"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"github.com/crewjam/saml"
	"net"
	"net/http"
	"reflect"
)

type CacheableIdpClientManager struct {
	template saml.ServiceProvider
	//for fetching idp metadata
	httpClient *http.Client

	cache map[string]*saml.ServiceProvider
	processed map[string]SamlIdentityProvider
}

func NewCacheableIdpClientManager(template saml.ServiceProvider) *CacheableIdpClientManager {
	return &CacheableIdpClientManager{
		template: template,
		httpClient:          http.DefaultClient,
		cache: make(map[string]*saml.ServiceProvider),
		processed: make(map[string]SamlIdentityProvider),
	}
}

func (m *CacheableIdpClientManager) RefreshCache(identityProviders []SamlIdentityProvider) {
	keep := make(map[string]bool)
	var refresh []SamlIdentityProvider

	for _, details := range identityProviders {
		if _, ok := m.cache[details.EntityId()]; !ok {
			refresh = append(refresh, details)
			m.processed[details.EntityId()] = details
		} else {
			processed := m.processed[details.EntityId()]
			if !reflect.DeepEqual(processed, details) {
				refresh = append(refresh, details)
			} else {
				keep[details.EntityId()] = true
			}
		}
	}

	for entityId := range m.cache {
		if _, ok := keep[entityId]; !ok {
			delete(m.cache, entityId)
			delete(m.processed, entityId)
		}
	}

	for _, details := range refresh {
		idpDescriptor, data, err := saml_util.ResolveMetadata(details.MetadataLocation(), m.httpClient)
		if err == nil {
			if details.ShouldMetadataRequireSignature() && idpDescriptor.Signature == nil{
				logger.Error("idp metadata rejected because it is not signed")
				continue
			}

			if details.ShouldMetadataTrustCheck() {
				var allCerts []*x509.Certificate
				for _, keyLoc := range details.GetMetadataTrustedKeys() {
					certs, err := cryptoutils.LoadCert(keyLoc)
					if err == nil {
						allCerts = append(allCerts, certs...)
					}
				}

				err = saml_util.VerifySignature(data, allCerts...)
				if err != nil {
					logger.Error("idp metadata rejected because it's signature cannot be verified")
					continue
				}
			}

			//make a copy
			client := m.template
			client.IDPMetadata = idpDescriptor

			_, port, err := net.SplitHostPort(client.AcsURL.Host)
			if err == nil {
				client.AcsURL.Host = net.JoinHostPort(details.Domain(), port)
			} else {
				client.AcsURL.Host = details.Domain()
			}

			_, port, err = net.SplitHostPort(client.SloURL.Host)
			if err == nil {
				client.SloURL.Host = net.JoinHostPort(details.Domain(), port)
			} else {
				client.SloURL.Host = details.Domain()
			}
			m.cache[details.EntityId()] = &client
		} else {
			logger.Error("could not resolve idp metadata", "details", details)
		}
	}
}

func (m *CacheableIdpClientManager) GetAllClients() []*saml.ServiceProvider {
	clients := make([]*saml.ServiceProvider, len(m.cache))
	idx := 0
	for  _, client := range m.cache {
		clients[idx] = client
		idx++
	}
	return clients
}

func (m *CacheableIdpClientManager) GetClientByComparator(comparator func(details SamlIdentityProvider) bool) (client *saml.ServiceProvider, ok bool) {
	for entityId, details := range m.processed {
		if comparator(details) {
			return m.cache[entityId], true
		}
	}
	return nil, false
}

func (m *CacheableIdpClientManager) GetClientByDomain(domain string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details SamlIdentityProvider) bool {
		return details.Domain() == domain
	})
}
func (m *CacheableIdpClientManager) GetClientByEntityId(entityId string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details SamlIdentityProvider) bool {
		return details.EntityId() == entityId
	})
}