package security

import (
	"context"
)

/*********************************
	Abstraction - DTO
 *********************************/

type Provider struct {
	Id               string
	Name             string
	DisplayName      string
	Description      string
	LocaleCode       string
	NotificationType string
	Email            string
	//CurrencyCode string
}

type Tenant struct {
	Id           string
	ExternalId   string
	DisplayName  string
	Description  string
	ProviderId   string
	Suspended    bool
	CurrencyCode string
	LocaleCode   string
}

/*********************************
	Abstraction - Stores
 *********************************/

type ProviderStore interface {
	LoadProviderById(ctx context.Context, id string) (*Provider, error)
}

// TODO: probably need to add another interface where the method is LoadUserTenantById(userId, tenantId)
// but make implementation of this interface optional, so that it doesn't affect existing implementation.
type TenantStore interface {
	LoadTenantById(ctx context.Context, id string) (*Tenant, error)
	LoadTenantByExternalId(ctx context.Context, name string) (*Tenant, error)
}
