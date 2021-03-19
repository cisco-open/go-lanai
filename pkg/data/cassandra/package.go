package cassandra

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"github.com/gocql/gocql"
)

var Module = &bootstrap.Module{
	Name: "cockroach",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(BindCassandraProperties, NewSession),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func NewSession(p CassandraProperties) *gocql.Session {
	cluster := gocql.NewCluster(p.Hosts()...)
	cluster.Keyspace = p.KeySpaceName
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: p.Username,
		Password: p.Password,
	}

	session, _ := cluster.CreateSession()
	return session
}

func BindCassandraProperties(ctx *bootstrap.ApplicationContext) CassandraProperties {
	p := NewCassandraProperties()
	ctx.Config().Bind(p, CassandraPropertiesPrefix)
	return *p
}