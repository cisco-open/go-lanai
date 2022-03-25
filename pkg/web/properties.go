package web

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/pkg/errors"
)

/***********************
	Server
************************/

const (
	ServerPropertiesPrefix = "server"
)

type ServerProperties struct {
	Port        int               `json:"port"`
	ContextPath string            `json:"context-path"`
	Logging     LoggingProperties `json:"logging"`
}

type LoggingProperties struct {
	Enabled      bool                              `json:"enabled"`
	DefaultLevel log.LoggingLevel                  `json:"default-level"`
	Levels       map[string]LoggingLevelProperties `json:"levels"`
}

// LoggingLevelProperties is used to override logging level on particular set of paths
// the LoggingProperties.Pattern support wildcard and should not include "context-path"
// the LoggingProperties.Method is space separated values. If left blank or contains "*", it matches all methods
type LoggingLevelProperties struct {
	Method  string           `json:"method"`
	Pattern string           `json:"pattern"`
	Level   log.LoggingLevel `json:"level"`
}

// NewServerProperties create a ServerProperties with default values
func NewServerProperties() *ServerProperties {
	return &ServerProperties{
		Port:        -1,
		ContextPath: "/",
		Logging: LoggingProperties{
			Enabled:      true,
			DefaultLevel: log.LevelDebug,
			Levels:       map[string]LoggingLevelProperties{},
		},
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
