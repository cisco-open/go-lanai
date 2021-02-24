package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"go.uber.org/fx"
)

func init() {
	bootstrap.AddOptions(
		fx.Invoke(configureSecurity),
	)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func configureSecurity(init security.Registrar, manager idp.AuthFlowManager) {
	// TODO might not needed
}