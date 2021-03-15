package repo

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

// GormCrud implements CrudRepository and can be embedded into any repositories using gorm as ORM
type GormCrud struct {
	GormApi
	model interface{}
}

func newGormCrud(api GormApi, model interface{}) (*GormCrud, error) {
	if e := validateGormModel(model); e != nil {
		return nil, e
	}
	return &GormCrud{
		GormApi: api,
		model: model,
	}, nil
}

func (g GormCrud) FindById(ctx context.Context, dest interface{}, id interface{}, options ...CrudOption) error {
	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Find(dest, id)
	})
}

func (g GormCrud) FindAll(ctx context.Context, dest interface{}, options ...CrudOption) error {
	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Find(dest)
	})
}

func (g GormCrud) FindBy(ctx context.Context, dest interface{}, condition CrudCondition, options ...CrudOption) error {
	return g.execute(ctx, condition, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Find(dest)
	})
}

func (g GormCrud) CountAll(ctx context.Context) (int, error) {
	var ret int64
	e := g.execute(ctx, nil, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Count(&ret)
	})
	if e != nil {
		return -1, e
	}
	return int(ret), nil
}

func (g GormCrud) CountBy(ctx context.Context, condition CrudCondition) (int, error) {
	var ret int64
	e := g.execute(ctx, condition, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Count(&ret)
	})
	if e != nil {
		return -1, e
	}
	return int(ret), nil
}

func (g GormCrud) Save(ctx context.Context, v interface{}, options...CrudOption) error {
	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Save(v)
	})
}

func (g GormCrud) Create(ctx context.Context, v interface{}, options...CrudOption) error {
	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Create(v)
	})
}

func (g GormCrud) Update(ctx context.Context, v interface{}, options...CrudOption) error {
	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Updates(v)
	})
}

func (g GormCrud) Delete(ctx context.Context, v interface{}) error {
	return g.execute(ctx, nil, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Delete(v)
	})
}

func (g GormCrud) DeleteBy(ctx context.Context, condition CrudCondition) error {
	return g.execute(ctx, condition, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Delete(g.model)
	})
}

func (g GormCrud) Truncate(ctx context.Context) error {
	return g.execute(ctx, nil, nil, func(db *gorm.DB) *gorm.DB {
		// FIXME this is not proper implementation, we need to know table name in this case
		return db.Model(g.model).Exec("TRUNCATE TABLE ? RESTRICT", reflect.TypeOf(g.model).Name())
	})
}

func (g GormCrud) execute(ctx context.Context, condition CrudCondition, options []CrudOption, f func(*gorm.DB) *gorm.DB) error {
	var e error
	db := g.GormApi.DB(ctx)
	if db, e = g.applyOptions(db, options); e != nil {
		return e
	}

	if db, e = g.applyCondition(db, condition); e != nil {
		return e
	}

	if r := f(db); r.Error != nil {
		return r.Error
	}
	return nil
}

func (g GormCrud) applyOptions(db *gorm.DB, opts []CrudOption) (*gorm.DB, error) {
	if opts == nil {
		return db, nil
	}
	for _, v := range opts {
		switch opt := v.(type) {
		case gormOptions:
			db = opt(db)
		default:
			return nil, ErrorUnsupportedOptions.WithMessage("unsupported CrudOption %T", v)
		}
	}
	return db, nil
}

func (g GormCrud) applyCondition(db *gorm.DB, condition CrudCondition) (*gorm.DB, error) {
	if condition == nil {
		return db, nil
	}

	switch where := condition.(type) {
	case gormOptions:
		db = where(db)
	case clause.Where:
		db = db.Clauses(where)
	default:
		db = db.Where(where)
	}
	return db, nil
}

func validateGormModel(model interface{}) error {
	if model == nil {
		return ErrorInvalidCrudModel.WithMessage("%T is not a valid model for gorm CRUD repository", model)
	}

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return nil
	}
	return ErrorInvalidCrudModel.WithMessage("%T is not a valid model for gorm CRUD repository", model)
}