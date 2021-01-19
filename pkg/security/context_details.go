package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type ContextDetailsStore interface {
	ReadContextDetails(ctx context.Context, key interface{}) (ContextDetails, error)
	SaveContextDetails(ctx context.Context, details ContextDetails) (key interface{}, err error)
	RemoveContextDetails(ctx context.Context, key interface{}) error
}

type ContextDetails interface {
	ProviderDetails
	TenantDetails
	UserDetails
	CredentialDetails
}

type ProviderDetails interface {
	ProviderId() string
	ProviderName() string
	ProviderDisplayName() string
}

type TenantDetails interface {
	TenantId() string
	TenantName() string
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

type CredentialDetails interface {
	ExpiryTime() time.Time
	Roles() utils.StringSet
	IssueTime() time.Time
	AuthenticationTime() time.Time
	OritinalUsername() string
	Masqueraded() bool
}