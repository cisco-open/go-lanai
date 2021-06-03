package scope

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
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
		fx.Invoke(configureSecurityScopeManagers),
	},
}

func Use() {
	seclient.Use()
	bootstrap.Register(Module)
}

// FxManagerCustomizers takes providers of ManagerCustomizer and wrap them with FxGroup
func FxManagerCustomizers(providers ...interface{}) []fx.Annotated {
	annotated := make([]fx.Annotated, len(providers))
	for i, t := range providers {
		annotated[i] = fx.Annotated{
			Group:  FxGroup,
			Target: t,
		}
	}
	return annotated
}

type secScopeDI struct {
	fx.In
	AuthClient       seclient.AuthenticationClient
	Properties       securityint.SecurityIntegrationProperties
	AuthServerConfig *authserver.Configuration `optional:"true"`
	ResServerConfig  *resserver.Configuration  `optional:"true"`
	Customizers      []ManagerCustomizer       `group:"security-scope"`
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

	// default options
	opts := []ManagerOptions{
		func(opt *managerOption) {
			opt.Client = di.AuthClient
			opt.Authenticator = authenticator
			opt.BackOffPeriod = time.Duration(di.Properties.FailureBackOff)
			opt.GuaranteedValidity = time.Duration(di.Properties.GuaranteedValidity)

			// parse accounts
			credentials := map[string]string{}
			sysAccts := utils.NewStringSet()
			if di.Properties.Accounts.Default.Username != "" {
				opt.DefaultSystemAccount = di.Properties.Accounts.Default.Username
				credentials[di.Properties.Accounts.Default.Username] = di.Properties.Accounts.Default.Password
				sysAccts.Add(di.Properties.Accounts.Default.Username)
			}
			for _, acct := range di.Properties.Accounts.Additional {
				if acct.Username == "" || acct.Password == "" {
					continue
				}
				credentials[acct.Username] = acct.Password
				if acct.SystemAccount {
					sysAccts.Add(acct.Username)
				}
			}
			opt.KnownCredentials = credentials
			opt.SystemAccounts = sysAccts
		},
	}

	// customizers
	for _, c := range di.Customizers {
		opts = append(opts, c.Customize()...)
	}

	scopeManager = newDefaultScopeManager(opts...)
}
