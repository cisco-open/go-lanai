package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"go.uber.org/fx"
)

type secDI struct {
	fx.In
	SecRegistrar security.Registrar
}

func NewResServerConfigurer(_ secDI) resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		//do nothing
	}
}
