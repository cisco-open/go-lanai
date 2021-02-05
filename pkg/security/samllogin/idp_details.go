package samllogin

type SamlIdpDetails struct {
	EntityId         string
	Domain           string
	MetadataLocation string
	ExternalIdName   string
	ExternalIdpName  string
	//TODO: option to require metadata to have signature, option to verify metadata signature
	// this is optional because both Spring and Okta's metadata are not signed
}

func (s SamlIdpDetails) GetDomain() string {
	return s.Domain
}
