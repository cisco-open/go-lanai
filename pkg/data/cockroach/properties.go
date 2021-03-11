package cockroach

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	ManagementPropertiesPrefix = "data.cockroach"
)

type CockroachProperties struct {
	//Enabled       bool                               `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

//NewCockroachProperties create a CockroachProperties with default values
func NewCockroachProperties() *CockroachProperties {
	return &CockroachProperties{
		Host:     "localhost",
		Port:     26257,
		Username: "root",
		Password: "root",
	}
}

//BindCockroachProperties create and bind SessionProperties, with a optional prefix
func BindCockroachProperties(ctx *bootstrap.ApplicationContext) CockroachProperties {
	props := NewCockroachProperties()
	if err := ctx.Config().Bind(props, ManagementPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CockroachProperties"))
	}
	return *props
}

