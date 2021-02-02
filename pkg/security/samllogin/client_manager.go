package samllogin

import (
	"context"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type CacheableIdpClientManager struct {
	template saml.ServiceProvider
	//for fetching idp metadata
	httpClient *http.Client

	cache map[IdentityProviderDetails]*saml.ServiceProvider
}

func NewCacheableIdpClientManager(template saml.ServiceProvider) *CacheableIdpClientManager {
	return &CacheableIdpClientManager{
		template: template,
		httpClient:          http.DefaultClient,
		cache: make(map[IdentityProviderDetails]*saml.ServiceProvider),
	}
}

func (m *CacheableIdpClientManager) RefreshCache(identityProviders []IdentityProviderDetails) {
	keep := make(map[IdentityProviderDetails]bool)
	var add []IdentityProviderDetails

	for _, details := range identityProviders {
		if _, ok := m.cache[details]; !ok {
			add = append(add, details)
		} else {
			keep[details] = true
		}
	}

	for details, _ := range m.cache {
		if _, ok := keep[details]; !ok {
			delete(m.cache, details)
		}
	}

	for _, details := range add {
		idpDescriptor, err := m.resolveIdpMetadata(details.MetadataLocation)
		if err == nil {
			//make a copy
			client := m.template
			client.IDPMetadata = idpDescriptor
			client.AcsURL.Host = details.Domain
			client.SloURL.Host = details.Domain

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

func (m *CacheableIdpClientManager) GetClientByComparator(comparator func(details IdentityProviderDetails) bool) (client *saml.ServiceProvider, ok bool) {
	for details, client := range m.cache {
		if comparator(details) {
			return client, true
		}
	}
	return nil, false
}

func (m *CacheableIdpClientManager) GetClientByDomain(domain string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details IdentityProviderDetails) bool {
		return details.Domain == domain
	})
}
func (m *CacheableIdpClientManager) GetClientByEntityId(entityId string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details IdentityProviderDetails) bool {
		return details.EntityId == entityId
	})
}

func (m *CacheableIdpClientManager) resolveIdpMetadata(metadataLocation string) (*saml.EntityDescriptor, error) {
	metadataUrl, err := url.Parse(metadataLocation)
	if err != nil {
		return nil, err
	}

	if metadataUrl.Scheme == "file" {
		file, err := os.Open(metadataUrl.Path)
		if err != nil {
			return nil, err
		}
		raw, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		metadata, err := samlsp.ParseMetadata(raw)
		return metadata, err
	} else {
		metadata, err := samlsp.FetchMetadata(context.TODO(), m.httpClient, *metadataUrl)
		return metadata, err
	}
}