package scope

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("SEC.Scope")

var Module = &bootstrap.Module{
	Name:       "security scope",
	Precedence: bootstrap.SecurityIntegrationPrecedence,
	Options: []fx.Option{
		fx.Provide(provideDefaultScopeManager),
		fx.Invoke(configureScopeManagers),
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

type defaultDI struct {
	fx.In
	AuthClient       seclient.AuthenticationClient             `optional:"true"`
	Properties       securityint.SecurityIntegrationProperties `optional:"true"`
	TokenStoreReader oauth2.TokenStoreReader                   `optional:"true"`
	Customizers      []ManagerCustomizer                       `group:"security-scope"`
}

func provideDefaultScopeManager(di defaultDI) ScopeManager {
	if di.TokenStoreReader == nil || di.AuthClient == nil {
		return nil
	}

	// default options
	opts := []ManagerOptions{
		func(opt *managerOption) {
			opt.Client = di.AuthClient
			opt.TokenStoreReader = di.TokenStoreReader
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
			// TBD, this is consistent behavior from java impl. Such configuration allows dev-ops to give
			// special treatment on certain accounts. Since we don't know any use case of this feature at
			// the time of writing this code, we temporarily disabled it, but keep the code for reference.
			//for _, acct := range di.Properties.Accounts.Additional {
			//	if acct.UName == "" || acct.Password == "" {
			//		continue
			//	}
			//	credentials[acct.UName] = acct.Password
			//	if acct.SystemAccount {
			//		sysAccts.Add(acct.UName)
			//	}
			//}
			opt.KnownCredentials = credentials
			opt.SystemAccounts = sysAccts
		},
	}

	// customizers
	for _, c := range di.Customizers {
		opts = append(opts, c.Customize()...)
	}

	return newDefaultScopeManager(opts...)
}

func configureScopeManagers(EffectiveScopeManager ScopeManager) {
	if EffectiveScopeManager == nil {
		msg := fmt.Sprintf(`Security Scope managers requires "resserver" and "seclient", but not configured`)
		logger.Warnf(msg)
		panic(msg)
	}
	scopeManager = EffectiveScopeManager
}
