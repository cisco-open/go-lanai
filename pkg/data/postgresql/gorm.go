// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package postgresql

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/certs"
	"github.com/cisco-open/go-lanai/pkg/data"
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
	AppContext   *bootstrap.ApplicationContext
	Properties   data.DatabaseProperties
	CertsManager certs.Manager `optional:"true"`
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
		source, e := di.CertsManager.Source(di.AppContext, certs.WithSourceProperties(&di.Properties.Tls.Certs))
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
