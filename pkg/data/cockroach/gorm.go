package cockroach

import (
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
	dsKeySslKeyPass  = "sslpassword "
)

type initDI struct {
	fx.In
	AppContext  *bootstrap.ApplicationContext
	Properties  CockroachProperties
	CertsManager tlsconfig.Manager `optional:"true"`
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
	if di.Properties.Tls.Enable && di.CertsManager != nil {
		source, e := di.CertsManager.Source(di.AppContext, tlsconfig.WithSourceProperties(&di.Properties.Tls.Config))
		if e == nil {
			certFiles, e := source.Files(di.AppContext)
			if e == nil {
				options[dsKeySslRootCert] = strings.Join(certFiles.RootCAPaths, " ")
				options[dsKeySslCert] = certFiles.CertificatePath
				options[dsKeySslKey] = certFiles.PrivateKeyPath
				if len(certFiles.PrivateKeyPassphrase) != 0 {
					options[dsKeySslKeyPass] = certFiles.PrivateKeyPassphrase
				}
			}
		} else {
			logger.Errorf("Failed to provision TLS certificates: %v", e)
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
