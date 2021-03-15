package repo

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"gorm.io/gorm"
)

type GormFactory struct {
	db *gorm.DB
	txManager tx.GormTxManager
	api GormApi
}

func newGormFactory(db *gorm.DB, txManager tx.GormTxManager) Factory {
	return &GormFactory{
		db: db,
		txManager: txManager,
		api: newGormApi(db, txManager),
	}
}

func (f GormFactory) NewCRUD(model interface{}, options...interface{}) CrudRepository {
	api := f.NewGormApi(options...)
	crud, e := newGormCrud(api, model)
	if e != nil {
		panic(e)
	}

	return crud
}

func (f GormFactory) NewGormApi(options...interface{}) GormApi {
	api := f.api
	for _, v := range options {
		switch opt := v.(type) {
		case gorm.Session:
			api = api.WithSession(&opt)
		case *gorm.Session:
			api = api.WithSession(opt)
		default:
			continue
		}
		break
	}
	return api
}