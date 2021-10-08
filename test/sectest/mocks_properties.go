package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	PropertiesPrefix = "mocking"
)

type mockingProperties struct {
	Accounts      map[string]*mockedAccountProperties `json:"accounts"`
	Tenants       map[string]*mockedTenantProperties  `json:"tenants"`
	TokenValidity utils.Duration                      `json:"token-validity"`
}

type mockedAccountProperties struct {
	UserId        string   `json:"id"` // optional field
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	DefaultTenant string   `json:"default-tenant"`
	Tenants       []string `json:"tenants"`
	Perms         []string `json:"permissions"`
}

type mockedTenantProperties struct {
	ID   string `json:"id"` // optional field
	ExternalId string `json:"external-id"`
}

func bindMockingProperties(ctx *bootstrap.ApplicationContext) *mockingProperties {
	props := mockingProperties{
		Accounts:      map[string]*mockedAccountProperties{},
		Tenants:       map[string]*mockedTenantProperties{},
		TokenValidity: utils.Duration(120 * time.Second),
	}
	if err := ctx.Config().Bind(&props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind mocking properties"))
	}
	return &props
}
