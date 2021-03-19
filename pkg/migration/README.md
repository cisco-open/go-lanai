#Data Migration

##Setup Migration App
To use data migration, first create a main method in your project.

For example

```go
package main

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	consul "cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/cassandra"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/cockroach"
	data "cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/migration"
	vault "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	"cto-github.cisco.com/livdu/europa/internal/migrations/v4_0"
	"go.uber.org/fx"
	"time"
)

func init() {
	// choose modules
	appconfig.Use()
	consul.Use()
	vault.Use()
	data.Use()
	cockroach.Use()
	cassandra.Use()
	migration.Use()

	v4_0.Use()
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"migrate",
		nil,
		[]fx.Option{fx.StartTimeout(1 * time.Minute)},
	)
	bootstrap.Execute()
}
```

1. In the init() function, declare all the packages that you want to use for your migration step. Including the package where the migration
steps are defined. In this case, this is v4_0.Use()
   
2. In the main() method, bootstrap the migration application.

##Add Migration Step

Create a package and add the migration steps

For example

```go
package v4_0

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/migration"
	"github.com/gocql/gocql"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var logger = log.New("migration_v4_0_0_1")

func init() {
	bootstrap.AddOptions(
		fx.Invoke(registerMigrations),
	)
}

func Use() {}

func registerMigrations(r *migration.Registrar,  cassandraSession *gocql.Session, db *gorm.DB) {
	r.AddMigrations(
		migration.WithVersion("4.0.0.1").WithTag(migration.TagPreUpgrade).WithFile("internal/migrations/v4_0/create_tenant_table.sql", db).WithDesc("create table"),
		migration.WithVersion("4.0.0.2").WithTag(migration.TagPreUpgrade).WithFunc(MoveTenantData(cassandraSession, db)).WithDesc("move data from cassandra to cockroach"),
	)
}
```

1. Use the standard go-lanai provider mechanism to register the migration steps
2. migration step can be either from a sql file, or a go function.
3. if your migration is a go function, you can inject any component that your migration needs as long as they are available through
the declaration in the init() method of your main function.
   
