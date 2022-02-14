package cassandra

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/gocql/gocql"
	"go.uber.org/fx"
	"time"
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
	cluster.Consistency = gocql.ParseConsistency(p.Consistency)
	cluster.Timeout = time.Duration(p.Timeout)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: p.Username,
		Password: p.Password,
	}

	session, _ := cluster.CreateSession()
	return session
}

func BindCassandraProperties(ctx *bootstrap.ApplicationContext) CassandraProperties {
	p := NewCassandraProperties()
	_ = ctx.Config().Bind(p, CassandraPropertiesPrefix) // Note, we don't panic if this bind is missing
	return *p
}