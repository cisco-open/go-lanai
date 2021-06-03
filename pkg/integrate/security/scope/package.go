package scope

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("SEC.Scope")

var Module = &bootstrap.Module{
	Name:       "security-Scope",
	Precedence: bootstrap.SecurityIntegrationPrecedence,
	Options: []fx.Option{
		//fx.Provide(bindHttpClientProperties),
		fx.Invoke(configureSecurityScopeManagers),
	},
}

func Use() {
	seclient.Use()
	bootstrap.Register(Module)
}

type secScopeDI struct {
	fx.In
	AuthClient       seclient.AuthenticationClient
	AuthServerConfig *authserver.Configuration `optional:"true"`
	ResServerConfig  *resserver.Configuration  `optional:"true"`
}

func configureSecurityScopeManagers(di secScopeDI) {
	var authenticator security.Authenticator
	switch {
	case di.AuthServerConfig != nil:
		authenticator = di.AuthServerConfig.SharedTokenAuthenticator()
	case di.ResServerConfig != nil:
		authenticator = di.ResServerConfig.SharedTokenAuthenticator()
	default:
		msg := fmt.Sprintf(`Security Scope managers requires "resserver" or "authserver", but none is configured`)
		logger.Warnf(msg)
		panic(msg)
	}

	scopeManager = newDefaultScopeManager(func(opt *managerOption) {
		// TODO user properties
		opt.Client = di.AuthClient
		opt.Authenticator = authenticator
		opt.BackOffPeriod = 5 * time.Second
		opt.GuaranteedValidity = 30 * time.Second
		opt.KnownCredentials = map[string]string{
			"livan": "password",
			"tim": "password",
		}
		opt.SystemAccounts = utils.NewStringSet("livan")
		opt.DefaultSystemAccount = "livan"
	})
}
