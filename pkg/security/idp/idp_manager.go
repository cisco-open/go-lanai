package idp


type IdentityProviderDetails interface {
	GetDomain() string
}

type IdentityProviderManager interface {
	GetAllIdentityProvider() []IdentityProviderDetails
	GetIdentityProviderByEntityId(entityId string) (IdentityProviderDetails, error)
}