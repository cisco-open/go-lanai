package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml_util"
	"errors"
	"fmt"
	"github.com/crewjam/saml"
	"net/http"
)

type SpMetadataManager struct {
	//for fetching idp metadata
	httpClient *http.Client
	cache map[SamlSpDetails]*saml.EntityDescriptor
}

func (m *SpMetadataManager) GetServiceProvider(serviceProviderID string) (SamlSpDetails, *saml.EntityDescriptor, error) {
	for k, v := range m.cache {
		if k.EntityId == serviceProviderID {
			return k, v, nil
		}
	}
	return SamlSpDetails{}, nil, errors.New(fmt.Sprintf("service provider metadata for %s not found", serviceProviderID))
}

func (m *SpMetadataManager) RefreshCache(clients []SamlClient) {
	keep := make(map[SamlSpDetails]bool)
	var add []SamlSpDetails

	for _, c := range clients {
		var details SamlSpDetails
		if defaultClient, ok := c.(DefaultSamlClient); ok {
			details = defaultClient.SamlSpDetails
		} else {
			details = SamlSpDetails{
				c.GetEntityId(),
				c.GetMetadataSource(),
				c.ShouldSkipAssertionEncryption(),
				c.ShouldSkipAuthRequestSignatureVerification(),
			}
		}

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
		spDescriptor, err := saml_util.ResolveMetadata(details.MetadataSource, m.httpClient)
		if err == nil {
			m.cache[details] = spDescriptor
		} else {
			logger.Error("could not resolve idp metadata", "details", details)
		}
	}
}