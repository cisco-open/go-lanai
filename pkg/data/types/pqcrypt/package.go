package pqcrypt

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"embed"
	"fmt"
	"go.uber.org/fx"
)

//var logger = log.New("Data.Enc")

//go:embed defaults-data-enc.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "data-encryption",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindDataEncryptionProperties, provideEncryptor),
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

type encDI struct {
	fx.In
	Properties DataEncryptionProperties `optional:"true"`
	Client     *vault.Client            `optional:"true"`
	UnnamedEnc Encryptor                `optional:"true"`
}

type encOut struct {
	fx.Out
	Enc Encryptor `name:"data/Encryptor"`
}

func provideEncryptor(di encDI) encOut {
	if di.UnnamedEnc != nil {
		return encOut{
			Enc: di.UnnamedEnc,
		}
	}

	var enc Encryptor
	switch {
	case di.Properties.Enabled:
		if di.Client == nil {
			panic(fmt.Errorf("data encryption enabled but vault client is not initialized"))
		}
		venc := newVaultEncryptor(di.Client, &di.Properties.Key)
		enc = compositeEncryptor{venc, plainTextEncryptor{}}
	default:
		enc = plainTextEncryptor{}
	}
	return encOut{
		Enc: enc,
	}
}

/**************************
	Initialize
***************************/
type initDI struct {
	fx.In
	Enc Encryptor `name:"data/Encryptor"`
}

func initialize(di initDI) {
	encryptor = di.Enc
}
