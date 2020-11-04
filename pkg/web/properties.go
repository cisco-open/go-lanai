package web

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

/***********************
	Server
************************/
const ServerPropertiesPrefix = "server"

type ServerProperties struct {
	Port        int    `json:"port"`
	ContextPath string `json:"context-path"`
}

// NewServerProperties create a ServerProperties with default values
func NewServerProperties() *ServerProperties {
	return &ServerProperties{
		Port: -1,
		ContextPath: "/",
	}
}

//BindServerProperties create and bind a ServerProperties using default prefix
func BindServerProperties(ctx *bootstrap.ApplicationContext) ServerProperties {
	props := NewServerProperties()
	if err := ctx.Config().Bind(props, ServerPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ServerProperties"))
	}
	return *props
}