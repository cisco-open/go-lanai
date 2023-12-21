// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package sp

import (
	"context"
	"crypto/x509"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	samlutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"github.com/crewjam/saml"
	"net"
	"net/http"
	"reflect"
	"sync"
)

type CacheableIdpClientManager struct {
	template saml.ServiceProvider
	//for fetching idp metadata
	httpClient *http.Client

	cache map[string]*saml.ServiceProvider
	processed map[string]samlctx.SamlIdentityProvider
	cacheMutex sync.RWMutex
}

func NewCacheableIdpClientManager(template saml.ServiceProvider) *CacheableIdpClientManager {
	return &CacheableIdpClientManager{
		template: template,
		httpClient:          http.DefaultClient,
		cache: make(map[string]*saml.ServiceProvider),
		processed: make(map[string]samlctx.SamlIdentityProvider),
	}
}

func (m *CacheableIdpClientManager) RefreshCache(ctx context.Context, identityProviders []samlctx.SamlIdentityProvider) {
	m.cacheMutex.RLock()
	remove, refresh := m.compareWithCache(identityProviders)
	m.cacheMutex.RUnlock()

	//nothing changed, just return
	if len(refresh) == 0 && len(remove) == 0{
		return
	}

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	//check again in case another process has already made the update
	remove, refresh = m.compareWithCache(identityProviders)
	if len(refresh) == 0 && len(remove) == 0{
		return
	}

	resolved := m.resolveMetadata(ctx, refresh)

	for entityId, doRemove := range remove {
		if doRemove {
			delete(m.cache, entityId)
			delete(m.processed, entityId)
		}
	}

	for _, details := range refresh {
		if client, ok := resolved[details.EntityId()]; ok {
			m.cache[details.EntityId()] = client
			m.processed[details.EntityId()] = details
		}
	}
}

func (m *CacheableIdpClientManager) compareWithCache(identityProviders []samlctx.SamlIdentityProvider) (remove map[string]bool, refresh []samlctx.SamlIdentityProvider) {
	keep := make(map[string]bool)
	remove = make(map[string]bool)

	for _, details := range identityProviders {
		if _, ok := m.cache[details.EntityId()]; !ok {
			refresh = append(refresh, details)
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
			remove[entityId] = true
		}
	}
	return remove, refresh
}

func (m *CacheableIdpClientManager) resolveMetadata(ctx context.Context, refresh []samlctx.SamlIdentityProvider) (resolved map[string]*saml.ServiceProvider){
	resolved = make(map[string]*saml.ServiceProvider)
	for _, details := range refresh {
		idpDescriptor, data, err := samlutils.ResolveMetadata(ctx, details.MetadataLocation(), samlutils.WithHttpClient(m.httpClient))
		if err == nil {
			if details.ShouldMetadataRequireSignature() && idpDescriptor.Signature == nil{
				logger.WithContext(ctx).Errorf("idp metadata rejected because it is not signed")
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

				err = samlutils.VerifySignature(samlutils.MetadataSignature(data, allCerts...))
				if err != nil {
					logger.WithContext(ctx).Errorf("idp metadata rejected because it's signature cannot be verified")
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
			resolved[details.EntityId()] = &client
		}
	}
	return resolved
}

func (m *CacheableIdpClientManager) GetAllClients() []*saml.ServiceProvider {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	clients := make([]*saml.ServiceProvider, len(m.cache))
	idx := 0
	for  _, client := range m.cache {
		clients[idx] = client
		idx++
	}
	return clients
}

func (m *CacheableIdpClientManager) GetClientByComparator(comparator func(details samlctx.SamlIdentityProvider) bool) (client *saml.ServiceProvider, ok bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	for entityId, details := range m.processed {
		if comparator(details) {
			return m.cache[entityId], true
		}
	}
	return nil, false
}

func (m *CacheableIdpClientManager) GetClientByDomain(domain string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details samlctx.SamlIdentityProvider) bool {
		return details.Domain() == domain
	})
}
func (m *CacheableIdpClientManager) GetClientByEntityId(entityId string) (client *saml.ServiceProvider, ok bool) {
	return m.GetClientByComparator(func(details samlctx.SamlIdentityProvider) bool {
		return details.EntityId() == entityId
	})
}