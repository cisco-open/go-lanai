package cockroach

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"net/url"
)

const (

)

type gormInitDI struct {
	fx.In
	AppContext *bootstrap.ApplicationContext
	Properties CockroachProperties
}

func NewGorm(di gormInitDI) *gorm.DB {
	var user *url.Userinfo
	if di.Properties.Username != "" {
		user = url.UserPassword(di.Properties.Username, di.Properties.Password)
	}
	uri := &url.URL{
		Scheme:      "postgres",
		User:        user,
		Host:        di.Properties.Host,
		Path:        di.Properties.Host,
	}

	cfg, e := pgx.ParseConfig(uri.String())
	logger.WithContext(di.AppContext).Infof("pgx config: %v, %v", cfg, e)
	//postgres.Config{
	//	DriverName:           "",
	//	DSN:                  "",
	//	PreferSimpleProtocol: false,
	//	WithoutReturning:     false,
	//	Conn:                 nil,
	//}
	//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	dsn := "host=localhost user=root password=root dbname=idm port=26257 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}