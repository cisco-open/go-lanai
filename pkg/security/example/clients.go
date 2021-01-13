package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const InmemoryClientsPropertiesPrefix = "security.in-memory"

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
	return &ClientsProperties {
		Clients: map[string]PropertiesBasedClient{},
	}
}

func BindClientsProperties(ctx *bootstrap.ApplicationContext) ClientsProperties {
	props := NewClientsProperties()
	if err := ctx.Config().Bind(props, InmemoryClientsPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ClientsProperties"))
	}
	return *props
}
