package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml_util"
	"github.com/crewjam/saml"
	"net"
	"net/http"
)

type CacheableIdpClientManager struct {
	template saml.ServiceProvider
	//for fetching idp metadata
	httpClient *http.Client

	cache map[SamlIdpDetails]*saml.ServiceProvider
}

func NewCacheableIdpClientManager(template saml.ServiceProvider) *CacheableIdpClientManager {
	return &CacheableIdpClientManager{
		template: template,
		httpClient:          http.DefaultClient,
		cache: make(map[SamlIdpDetails]*saml.ServiceProvider),
	}
}

func (m *CacheableIdpClientManager) RefreshCache(identityProviders []SamlIdpDetails) {
	keep := make(map[SamlIdpDetails]bool)
	var add []SamlIdpDetails

	for _, details := range identityProviders {
		if _, ok := m.cache[details]; !ok {
			add = append(add, details)
		} else {
			keep[details] = true
		}
	}

	for details := range m.cache {
		if _, ok := keep[details]; !ok {
			delete(m.cache, details)
		}
	}

	for _, details := range add {
		idpDescriptor, err := saml_util.ResolveMetadata(details.MetadataLocation, m.httpClient)
		if err == nil {
			//make a copy
			client := m.template
			client.IDPMetadata = idpDescriptor

			_, port, err := net.SplitHostPort(client.AcsURL.Host)
			if err == nil {
				client.AcsURL.Host = net.JoinHostPort(details.Domain, port)
			} else {
				client.AcsURL.Host = details.Domain
			}

			_, port, err = net.SplitHostPort(client.SloURL.Host)
			if err == nil {
				client.SloURL.Host = net.JoinHostPort(details.Domain, port)
			} else {
				client.SloURL.Host = details.Domain
			}
			m.cache[details] = &client
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

func (m *CacheableIdpClientManager) GetClientByComparator(comparator func(details SamlIdpDetails) bool) (client *saml.ServiceProvider, ok bool) {
	for details, client := range m.cache {
		if comparator(details) {
			return client, true
		}
	}
	return nil, false
}

func (m *CacheableIdpClientManager) GetClientByDomain(domain string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details SamlIdpDetails) bool {
		return details.Domain == domain
	})
}
func (m *CacheableIdpClientManager) GetClientByEntityId(entityId string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details SamlIdpDetails) bool {
		return details.EntityId == entityId
	})
}