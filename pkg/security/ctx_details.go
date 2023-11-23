package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type ContextDetailsStore interface {
	ReadContextDetails(ctx context.Context, key interface{}) (ContextDetails, error)
	SaveContextDetails(ctx context.Context, key interface{}, details ContextDetails) error
	RemoveContextDetails(ctx context.Context, key interface{}) error
	ContextDetailsExists(ctx context.Context, key interface{}) bool
}

type ContextDetails interface {
	AuthenticationDetails
	KeyValueDetails
}

// ProviderDetails is available if tenant is selected (tenant dictates provider)
type ProviderDetails interface {
	ProviderId() string
	ProviderName() string
	ProviderDisplayName() string
	ProviderDescription() string
	ProviderEmail() string
	ProviderNotificationType() string
}

// TenantDetails is available in the following scenarios:
//
//	user auth, tenant can be determined (either selected tenant, or there is a default tenant)
//	client auth, tenant can be determined (either selected tenant, or there is a default tenant)
type TenantDetails interface {
	TenantId() string
	TenantExternalId() string
	TenantSuspended() bool
}

// TenantAccessDetails This is available if authenticated entity is supposed to have access to tenants.
type TenantAccessDetails interface {
	EffectiveAssignedTenantIds() utils.StringSet
}

// UserDetails is available for user authentication
type UserDetails interface {
	UserId() string
	Username() string
	AccountType() AccountType
	AssignedTenantIds() utils.StringSet
	LocaleCode() string
	CurrencyCode() string
	FirstName() string
	LastName() string
	Email() string
}

type AuthenticationDetails interface {
	ExpiryTime() time.Time
	IssueTime() time.Time
	Roles() utils.StringSet
	Permissions() utils.StringSet
	AuthenticationTime() time.Time
}

type ProxiedUserDetails interface {
	OriginalUsername() string
	Proxied() bool
}

type KeyValueDetails interface {
	Value(string) (interface{}, bool)
	Values() map[string]interface{}
}
