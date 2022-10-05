package samltest

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

type ProviderProperties struct {
	EntityID         string `json:"entity-id"`
	MetadataSource   string `json:"metadata-source"`
	CertsSource      string `json:"certs"`
	PrivateKeySource string `json:"private-key"`
}

type ExtIDPProperties struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	IdName string `json:"id-name"`
}

type IDPProperties struct {
	ProviderProperties
	ExtIDPProperties
	SSOPath string `json:"sso"`
	SLOPath string `json:"slo"`
}

type SPProperties struct {
	ProviderProperties
	ACSPath string         `json:"acs"`
	SLOPath string         `json:"slo"`
	IDP     *IDPProperties `json:"idp"`
}

type MockedClientProperties struct {
	SPProperties
	SkipEncryption            bool                      `json:"skip-encryption"`
	SkipSignatureVerification bool                      `json:"skip-signature-verification"`
	TenantRestriction         utils.CommaSeparatedSlice `json:"tenant-restriction"`
	TenantRestrictionType     string                    `json:"tenant-restriction-type"`
}
