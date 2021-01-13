package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const InmemoryAccountsPropertiesPrefix = "security.in-memory"

type PropertiesBasedAccount struct {
	ID                string   `json:"id"`
	Type              string   `json:"type"`
	Username          string   `json:"username"`
	Password          string   `json:"password"`
	Permissions       []string `json:"permissions"`
	Disabled          bool     `json:"disabled"`
	Locked            bool     `json:"locked"`
	UseMFA            bool     `json:"mfa-enabled"`
	DefaultTenantId   string   `json:"default-tenant-id"`
	Tenants           []string `json:"tenants"`
	AccountPolicyName string   `json:"policy-name"`
}

type PropertiesBasedAccountPolicy struct {
	Name                string `json:"name"`
	LockingEnabled      bool   `json:"lock-enabled"`
	LockoutDuration     string `json:"lock-duration"`
	FailuresLimit       int    `json:"failure-limit"`
	FailuresInterval    string `json:"failure-interval"`
	AgingEnabled        bool   `json:"aging-enabled"`
	MaxAge              string `json:"max-age"`
	ExpiryWarningPeriod string `json:"warning-period"`
	GracefulAuthLimit   int    `json:"graceful-auth-limit"`
}

type AccountsProperties struct {
	Accounts map[string]PropertiesBasedAccount `json:"accounts"`
}

type AccountPoliciesProperties struct {
	Policies map[string]PropertiesBasedAccountPolicy `json:"policies"`
}

func NewAccountsProperties() *AccountsProperties {
	return &AccountsProperties {
		Accounts: map[string]PropertiesBasedAccount{},
	}
}

func BindAccountsProperties(ctx *bootstrap.ApplicationContext) AccountsProperties {
	props := NewAccountsProperties()
	if err := ctx.Config().Bind(props, InmemoryAccountsPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AccountsProperties"))
	}
	return *props
}

func NewAccountPoliciesProperties() *AccountPoliciesProperties {
	return &AccountPoliciesProperties {
		Policies: map[string]PropertiesBasedAccountPolicy{},
	}
}

func BindAccountPoliciesProperties(ctx *bootstrap.ApplicationContext) AccountPoliciesProperties {
	props := NewAccountPoliciesProperties()
	if err := ctx.Config().Bind(props, InmemoryAccountsPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AccountPoliciesProperties"))
	}
	return *props
}