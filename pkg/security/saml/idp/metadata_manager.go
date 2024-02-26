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

package samlidp

import (
    "context"
    "crypto/x509"
    "errors"
    "fmt"
    samlctx "github.com/cisco-open/go-lanai/pkg/security/saml"
    samlutils "github.com/cisco-open/go-lanai/pkg/security/saml/utils"
    "github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
    "github.com/crewjam/saml"
    "net/http"
    "reflect"
    "sync"
)

type SpMetadataManager struct {
	//for fetching idp metadata
	httpClient *http.Client
	//entityId to descriptor
	cache map[string]*saml.EntityDescriptor
	processed map[string]SamlSpDetails
	cacheMutex sync.RWMutex
}

func (m *SpMetadataManager) GetServiceProvider(serviceProviderID string) (SamlSpDetails, *saml.EntityDescriptor, error) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	for k, v := range m.cache {
		if k == serviceProviderID {
			return m.processed[k], v, nil
		}
	}
	return SamlSpDetails{}, nil, errors.New(fmt.Sprintf("service provider metadata for %s not found", serviceProviderID))
}

func (m *SpMetadataManager) RefreshCache(ctx context.Context, clients []samlctx.SamlClient) {
	m.cacheMutex.RLock()
	remove, refresh := m.compareWithCache(clients)
	m.cacheMutex.RUnlock()

	//nothing changed, just return
	if len(refresh) == 0 && len(remove) == 0{
		return
	}

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	//check again because another process may have got into the write lock first and updated cache
	remove, refresh = m.compareWithCache(clients)
	//if the cache was updated by another process, then just return
	if len(refresh) == 0 && len(remove) == 0 {
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
		if spDescriptor, ok := resolved[details.EntityId]; ok {
			m.cache[details.EntityId] = spDescriptor
			m.processed[details.EntityId] = details
		}
	}
}

func (m *SpMetadataManager) compareWithCache(clients []samlctx.SamlClient) (remove map[string]bool, refresh []SamlSpDetails) {
	keep := make(map[string]bool)
	remove = make(map[string]bool)

	for _, c := range clients {
		var details SamlSpDetails
		if defaultClient, ok := c.(DefaultSamlClient); ok {
			details = defaultClient.SamlSpDetails
		} else {
			details = SamlSpDetails{
				EntityId:c.GetEntityId(),
				MetadataSource: c.GetMetadataSource(),
				SkipAssertionEncryption: c.ShouldSkipAssertionEncryption(),
				SkipAuthRequestSignatureVerification: c.ShouldSkipAuthRequestSignatureVerification(),
				MetadataRequireSignature: c.ShouldMetadataRequireSignature(),
				MetadataTrustCheck: c.ShouldMetadataTrustCheck(),
				MetadataTrustedKeys: c.GetMetadataTrustedKeys(),
			}
		}

		if _, ok := m.cache[details.EntityId]; !ok {
			refresh = append(refresh, details)
		} else {
			processed := m.processed[details.EntityId]
			if !reflect.DeepEqual(processed, details) {
				refresh = append(refresh, details)
			} else {
				keep[details.EntityId] = true
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

func (m *SpMetadataManager) resolveMetadata(ctx context.Context, refresh []SamlSpDetails) (resolved map[string]*saml.EntityDescriptor) {
	resolved = make(map[string]*saml.EntityDescriptor)
	for _, details := range refresh {
		spDescriptor, data, err := samlutils.ResolveMetadata(ctx, details.MetadataSource, samlutils.WithHttpClient(m.httpClient))
		if err == nil {
			if details.MetadataRequireSignature && spDescriptor.Signature == nil{
				logger.WithContext(ctx).Warnf("sp metadata rejected because it is not signed")
				continue
			}

			if details.MetadataTrustCheck {
				var allCerts []*x509.Certificate
				for _, keyLoc := range details.MetadataTrustedKeys {
					certs, err := cryptoutils.LoadCert(keyLoc)
					if err == nil {
						allCerts = append(allCerts, certs...)
					}
				}

				err = samlutils.VerifySignature(samlutils.MetadataSignature(data, allCerts...))
				if err != nil {
					logger.WithContext(ctx).Warnf("sp metadata rejected because it's signature cannot be verified")
					continue
				}
			}
			resolved[details.EntityId] = spDescriptor
		} else {
			logger.WithContext(ctx).Warnf("could not resolve idp metadata", "details", details)
		}
	}
	return resolved
}