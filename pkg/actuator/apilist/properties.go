package apilist

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "management.endpoint.apilist"
)

type FSType int

type Properties struct {
	StaticPath string `json:"static-path"`
}


//NewProperties create a Properties with default values
func NewProperties() *Properties {
	return &Properties{
		StaticPath: "configs/api-list.json",
	}
}

//BindProperties create and bind SessionProperties, with a optional prefix
func BindProperties(ctx *bootstrap.ApplicationContext) Properties {
	props := NewProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind Properties"))
	}
	return *props
}
