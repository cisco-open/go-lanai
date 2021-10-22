package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

const(
	// e.g. *Model
	typeModelPtr typeKey = iota
	// e.g. Model
	typeModel
	// e.g. *[]Model
	typeModelSlicePtr
	// e.g. *[]*Model{}
	typeModelPtrSlicePtr
	// e.g. []Model
	typeModelSlice
	// e.g. []*Model
	typeModelPtrSlice
	// map[string]interface{}
	typeGenericMap
)

type typeKey int

var (
	singleModelRead   = utils.NewSet(typeModelPtr)
	multiModelRead    = utils.NewSet(typeModelPtrSlicePtr, typeModelSlicePtr)
	singleModelWrite  = utils.NewSet(typeModelPtr, typeModel)
	//multiModelWrite   = utils.NewSet(typeModelPtrSlice, typeModelSlice, typeModelPtrSlicePtr, typeModelSlicePtr)
	genericModelWrite = utils.NewSet(
		typeModelPtr,
		typeModelPtrSlice,
		typeGenericMap,
		typeModelPtrSlicePtr,
		typeModelSlice,
		typeModelSlicePtr,
		typeModel,
	)
)

// GormCrud implements CrudRepository and can be embedded into any repositories using gorm as ORM
type GormCrud struct {
	GormApi
	GormMetadata
}

func newGormCrud(api GormApi, model interface{}) (*GormCrud, error) {
	// Note we uses raw db here to leverage internal schema cache
	meta, e := newModelMetadata(api.DB(context.Background()), model)
	if e != nil {
		return nil, e
	}
	return &GormCrud{
		GormApi:      api,
		GormMetadata: meta,
	}, nil
}

func (g GormCrud) FindById(ctx context.Context, dest interface{}, id interface{}, options ...Option) error {
	if !g.isSupportedValue(dest, singleModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", dest, "FindById", "*Struct")
	}

	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		// TODO verify this using composite key
		switch v := id.(type) {
		case string:
			if uid, e := uuid.Parse(v); e == nil {
				id = uid
			}
		case *string:
			if uid, e := uuid.Parse(*v); e == nil {
				id = uid
			}
		}
		return db.Model(g.model).Take(dest, id)
	})
}

func (g GormCrud) FindAll(ctx context.Context, dest interface{}, options ...Option) error {
	if !g.isSupportedValue(dest, multiModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", dest, "FindAll", "*[]Struct or *[]*Struct")
	}

	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Find(dest)
	})
}

func (g GormCrud) FindOneBy(ctx context.Context, dest interface{}, condition Condition, options...Option) error {
	if !g.isSupportedValue(dest, singleModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", dest, "FindOneBy", "*Struct")
	}

	return g.execute(ctx, condition, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Take(dest)
	})
}

func (g GormCrud) FindAllBy(ctx context.Context, dest interface{}, condition Condition, options ...Option) error {
	if !g.isSupportedValue(dest, multiModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", dest, "FindAllBy", "*[]Struct or *[]*Struct")
	}

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

func (g GormCrud) CountBy(ctx context.Context, condition Condition) (int, error) {
	var ret int64
	e := g.execute(ctx, condition, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Count(&ret)
	})
	if e != nil {
		return -1, e
	}
	return int(ret), nil
}

func (g GormCrud) Save(ctx context.Context, v interface{}, options...Option) error {
	if !g.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", v, "Save", "*Struct, []*Struct or []Struct")
	}

	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Save(v)
	})
}

func (g GormCrud) Create(ctx context.Context, v interface{}, options...Option) error {
	if !g.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", v, "Create", "*Struct, []*Struct or []Struct")
	}

	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Create(v)
	})
}

func (g GormCrud) Update(ctx context.Context, model interface{}, v interface{}, options...Option) error {
	if !g.isSupportedValue(model, singleModelWrite) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid model for %s, requires %s", v, "Update", "*Struct or Struct")
	}

	return g.execute(ctx, nil, options, func(db *gorm.DB) *gorm.DB {
		// note we use the actual model instead of template g.model
		return db.Model(model).Updates(v)
	})
}

func (g GormCrud) Delete(ctx context.Context, v interface{}) error {
	if !g.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidCrudParam.
			WithMessage("%T is not a valid value for %s, requires %s", v, "Delete", "*Struct, []*Struct or []Struct")
	}

	return g.execute(ctx, nil, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Delete(v)
	})
}

func (g GormCrud) DeleteBy(ctx context.Context, condition Condition) error {
	return g.execute(ctx, condition, nil, func(db *gorm.DB) *gorm.DB {
		return db.Model(g.model).Delete(g.model)
	})
}

func (g GormCrud) Truncate(ctx context.Context) error {
	return g.execute(ctx, nil, nil, func(db *gorm.DB) *gorm.DB {
		db = db.Model(g.model)
		if e := db.Statement.Parse(g.model); e != nil {
			_ = db.AddError(ErrorInvalidCrudModel.WithMessage("unable to parse table name for model %T", g.model))
			return db
		}
		table := interface{}(db.Statement.TableExpr)
		if db.Statement.TableExpr == nil {
			table = db.Statement.Table
		}
		return db.Exec("TRUNCATE TABLE ? RESTRICT", table)
	})
}

func (g GormCrud) execute(ctx context.Context, condition Condition, options []Option, f func(*gorm.DB) *gorm.DB) error {
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

func (g GormCrud) applyOptions(db *gorm.DB, opts []Option) (*gorm.DB, error) {
	if opts == nil {
		return db, nil
	}
	for _, v := range opts {
		switch opt := v.(type) {
		case gormOptions:
			db = opt(db)
		case func(*gorm.DB) *gorm.DB:
			db = opt(db)
		default:
			return nil, ErrorUnsupportedOptions.WithMessage("unsupported Option %T", v)
		}
	}
	return db, nil
}

func (g GormCrud) applyCondition(db *gorm.DB, condition Condition) (*gorm.DB, error) {
	if condition == nil {
		return db, nil
	}
	var e error
	switch cv := reflect.ValueOf(condition); cv.Kind() {
	case reflect.Slice, reflect.Array:
		size := cv.Len()
		for i := 0; i < size; i++ {
			if db, e = g.applyCondition(db, cv.Index(i).Interface()); e != nil {
				return nil, e
			}
		}
	default:
		switch where := condition.(type) {
		case gormOptions:
			db = where(db)
		case func(*gorm.DB) *gorm.DB:
			db = where(db)
		case clause.Where:
			db = db.Clauses(where)
		default:
			db = db.Where(where)
		}
	}

	return db, nil
}

func (g GormCrud) isSupportedValue(value interface{}, types utils.Set) bool {
	t := reflect.TypeOf(value)
	typ, ok := g.types[t]
	return ok && types.Has(typ)
}