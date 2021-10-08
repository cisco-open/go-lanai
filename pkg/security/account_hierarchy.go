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

type TenantStore interface {
	LoadTenantById(ctx context.Context, id string) (*Tenant, error)
	LoadTenantByExternalId(ctx context.Context, name string) (*Tenant, error)
}
