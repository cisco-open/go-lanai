package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Passwd")

func init() {
	bootstrap.AddOptions(
		fx.Invoke(NewPasswordIdpSecurityConfigurer),
	)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func configureSecurity(init security.Registrar, store security.AccountStore, manager idp.AuthFlowManager) {
	// TODO this might not be needed
}
