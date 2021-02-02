package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"go.uber.org/fx"
)

/**
	1. Generate metadata from configuration (i.e. issuer)
	2. Add idp metadata via API
	3. compare saml library's code for checking assertion against that of the java implementation

Implementation details:
	1. add entry point
	2. generate metadata and add metadata API
	3. add idp metadata via API
 */


func init() {
	bootstrap.AddOptions(
		fx.Invoke(configureSecurity),
	)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func configureSecurity(init security.Registrar, manager idp.AuthFlowManager) {
	init.Register(&SamlConfigurer{
		authFlowManager: manager,
	})

}