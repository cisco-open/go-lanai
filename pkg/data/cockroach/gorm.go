package cockroach

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

const (
	dsKeyHost        = "host"
	dsKeyPort        = "port"
	dsKeyDB          = "dbname"
	dsKeySslMode     = "sslmode"
	dsKeyUsername    = "user"
	dsKeyPassword    = "password"
	dsKeySslRootCert = "sslrootcert"
	dsKeySslCert     = "sslcert"
	dsKeySslKey      = "sslkey"
)

type initDI struct {
	fx.In
	AppContext  *bootstrap.ApplicationContext
	Properties  CockroachProperties
	TLSProvider *tlsconfig.ProviderFactory
}

func NewGormDialetor(di initDI) gorm.Dialector {
	//"host=localhost user=root password=root dbname=idm port=26257 sslmode=disable"
	options := map[string]interface{}{
		dsKeyHost:    di.Properties.Host,
		dsKeyPort:    di.Properties.Port,
		dsKeyDB:      di.Properties.Database,
		dsKeySslMode: di.Properties.SslMode,
	}
	// Setup TLS properties
	if di.Properties.Tls.Enable {
		if !di.Properties.Tls.Config.FileCache.Enabled && di.Properties.Tls.Config.Type == "vault" {
			logger.Error("Can't enable tls for postgres driver with vault tls provider without enabling FileCache.")
		} else {
			provider, err := di.TLSProvider.GetProvider(di.Properties.Tls.Config)
			if err != nil {
				logger.Error("Failed to provision tls provider.")
			}
			ctx := context.Background()
			_, err = provider.GetClientCertificate(ctx)
			if err != nil {
				logger.Error("Failed to fetch tls certificate.")
			}
			_, err = provider.RootCAs(ctx)
			if err != nil {
				logger.Error("Failed to fetch ca certificate.")
			}
			basePath := di.Properties.Tls.Config.FileCache.Path + di.Properties.Tls.Config.FileCache.Prefix
			options[dsKeySslRootCert] = basePath + tlsconfig.CaSuffix
			options[dsKeySslCert] = basePath + tlsconfig.CertSuffix
			options[dsKeySslKey] = basePath + tlsconfig.KeySuffix
		}
	}

	if di.Properties.Username != "" {
		options[dsKeyUsername] = di.Properties.Username
		options[dsKeyPassword] = di.Properties.Password
	}

	config := postgres.Config{
		//DriverName:           "postgres",
		DSN: toDSN(options),
	}
	return NewGormDialectorWithConfig(config)
}

func toDSN(options map[string]interface{}) string {
	opts := []string{}
	for k, v := range options {
		opt := fmt.Sprintf("%s=%v", k, v)
		opts = append(opts, opt)
	}
	return strings.Join(opts, " ")
}
