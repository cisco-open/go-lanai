package dbtest

import (
	"context"
	"gorm.io/gorm"
)

/*************************
	Enums
 *************************/

const (
	modeAuto mode = iota
	modePlayback
	modeRecord
)

type mode int

/*************************
	DBOptions
 *************************/

type DBOptions func(opt *DBOption)
type DBOption struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
	SSL      bool
}

func DBName(db string) DBOptions {
	return func(opt *DBOption) {
		opt.DBName = db
	}
}

func DBCredentials(user, password string) DBOptions {
	return func(opt *DBOption) {
		opt.Username = user
		opt.Password = password
	}
}

func DBPort(port int) DBOptions {
	return func(opt *DBOption) {
		opt.Port = port
	}
}

func DBHost(host string) DBOptions {
	return func(opt *DBOption) {
		opt.Host = host
	}
}

/*************************
	TX context
 *************************/

type mockedTxContext struct {
	context.Context
}

func (c mockedTxContext) Parent() context.Context {
	return c.Context
}

type mockedGormContext struct {
	mockedTxContext
	db *gorm.DB
}

func (c mockedGormContext) DB() *gorm.DB {
	return c.db
}

