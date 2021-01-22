package samllogin

//TODO: merge this with form login
type IdentityProviderDetails struct {
	EntityId         string
	Domain           string //internal Domain
	MetadataLocation string
	ExternalIdName   string
	ExternalIdpName  string
	//TODO: option to require metadata to have signature, option to verify metadata signature
	// this is optional because both Spring and Okta's metadata are not signed
}

type IdentityProviderManager interface {
	GetAllIdentityProvider() []IdentityProviderDetails
	GetIdentityProviderByEntityId(entityId string) IdentityProviderDetails
}