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

type ProviderDetails interface {
	ProviderId() string
	ProviderName() string
	ProviderDisplayName() string
	ProviderDescription() string
	ProviderEmail() string
	ProviderNotificationType() string
}

type TenantDetails interface {
	TenantId() string
	TenantExternalId() string
	TenantSuspended() bool
}

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

// TODO: review if this is a suitable package for this interface
type ClientDetails interface {
	ClientId() string
	AssignedTenantIds() utils.StringSet
	Scopes() utils.StringSet
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
