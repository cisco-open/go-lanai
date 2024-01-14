package service

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const InMemoryPropertiesPrefix = "security.in-memory"

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
	FullName          string   `json:"full-name"`
	Email             string   `json:"email"`
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
	return &AccountsProperties{
		Accounts: map[string]PropertiesBasedAccount{},
	}
}

func BindAccountsProperties(ctx *bootstrap.ApplicationContext) AccountsProperties {
	props := NewAccountsProperties()
	if err := ctx.Config().Bind(props, InMemoryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AccountsProperties"))
	}
	return *props
}

func NewAccountPoliciesProperties() *AccountPoliciesProperties {
	return &AccountPoliciesProperties{
		Policies: map[string]PropertiesBasedAccountPolicy{},
	}
}

func BindAccountPoliciesProperties(ctx *bootstrap.ApplicationContext) AccountPoliciesProperties {
	props := NewAccountPoliciesProperties()
	if err := ctx.Config().Bind(props, InMemoryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AccountPoliciesProperties"))
	}
	return *props
}

type TenantProperties struct {
	Tenants map[string]PropertiesBasedTenant `json:"tenants"`
}

func newTenantProperties() *TenantProperties {
	return &TenantProperties{
		Tenants: map[string]PropertiesBasedTenant{},
	}
}

type PropertiesBasedTenant struct {
	ID         string `json:"id"` // optional field
	ExternalId string `json:"external-id"`
	Name       string `json:"name"`
}

func BindTenantProperties(ctx *bootstrap.ApplicationContext) TenantProperties {
	props := newTenantProperties()
	if err := ctx.Config().Bind(props, InMemoryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind TenantProperties"))
	}
	return *props
}

type PropertiesBasedClient struct {
	ClientId             string   `json:"client-id"`
	Secret               string   `json:"secret"`
	GrantTypes           []string `json:"grant-types"`
	RedirectUris         []string `json:"redirect-uris"`
	Scopes               []string `json:"scopes"`
	AutoApproveScopes    []string `json:"auto-approve-scopes"`
	AccessTokenValidity  string   `json:"access-token-validity"`
	RefreshTokenValidity string   `json:"refresh-token-validity"`
	UseSessionTimeout    bool     `json:"use-session-timeout"`
	TenantRestrictions   []string `json:"tenant-restrictions"`
}

type ClientsProperties struct {
	Clients map[string]PropertiesBasedClient `json:"clients"`
}

func NewClientsProperties() *ClientsProperties {
	return &ClientsProperties{
		Clients: map[string]PropertiesBasedClient{},
	}
}

func BindClientsProperties(ctx *bootstrap.ApplicationContext) ClientsProperties {
	props := NewClientsProperties()
	if err := ctx.Config().Bind(props, InMemoryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ClientsProperties"))
	}
	return *props
}
