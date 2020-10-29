package web

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

/***********************
	General
************************/
type bindingDependencies struct {
	fx.In
	Config *appconfig.Config `name:"bootstrap_config"`
}

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
func BindServerProperties(d bindingDependencies) ServerProperties {
	props := NewServerProperties()
	if err := d.Config.Bind(props, ServerPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ServerProperties"))
	}
	return *props
}