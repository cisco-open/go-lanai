package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

const (
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

const (
	errTmplInvalidCrudValue = `%T is not a valid value for %s, requires %s`
	errTmplInvalidCrudModel = "%T is not a valid model for %s, requires %s"
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
	meta, e := newGormMetadata(api.DB(context.Background()), model)
	if e != nil {
		return nil, e
	}
	ret := &GormCrud{
		GormApi:      api,
		GormMetadata: meta,
	}
	return ret, nil
}

func (g GormCrud) FindById(ctx context.Context, dest interface{}, id interface{}, options ...Option) error {
	if !g.isSupportedValue(dest, singleModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage(errTmplInvalidCrudValue, dest, "FindById", "*Struct")
	}

	// TODO verify this using index key
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

	return execute(ctx, g.GormApi.DB(ctx), nil, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Take(dest, id)
	})
}

func (g GormCrud) FindAll(ctx context.Context, dest interface{}, options ...Option) error {
	if !g.isSupportedValue(dest, multiModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage(errTmplInvalidCrudValue, dest, "FindAll", "*[]Struct or *[]*Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), nil, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Find(dest)
	})
}

func (g GormCrud) FindOneBy(ctx context.Context, dest interface{}, condition Condition, options...Option) error {
	if !g.isSupportedValue(dest, singleModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage(errTmplInvalidCrudValue, dest, "FindOneBy", "*Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), condition, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Take(dest)
	})
}

func (g GormCrud) FindAllBy(ctx context.Context, dest interface{}, condition Condition, options ...Option) error {
	if !g.isSupportedValue(dest, multiModelRead) {
		return ErrorInvalidCrudParam.
			WithMessage(errTmplInvalidCrudValue, dest, "FindAllBy", "*[]Struct or *[]*Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), condition, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Find(dest)
	})
}

func (g GormCrud) CountAll(ctx context.Context, options...Option) (int, error) {
	var ret int64
	e := execute(ctx, g.GormApi.DB(ctx), nil, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Count(&ret)
	})
	if e != nil {
		return -1, e
	}
	return int(ret), nil
}

func (g GormCrud) CountBy(ctx context.Context, condition Condition, options...Option) (int, error) {
	var ret int64
	e := execute(ctx, g.GormApi.DB(ctx), condition, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Count(&ret)
	})
	if e != nil {
		return -1, e
	}
	return int(ret), nil
}

func (g GormCrud) Save(ctx context.Context, v interface{}, options...Option) error {
	if !g.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidCrudParam.WithMessage(errTmplInvalidCrudValue, v, "Save", "*Struct or []*Struct or []Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), nil, options, nil, func(db *gorm.DB) *gorm.DB {
		return db.Save(v)
	})
}

func (g GormCrud) Create(ctx context.Context, v interface{}, options...Option) error {
	if !g.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidCrudParam.WithMessage(errTmplInvalidCrudValue, v, "Create", "*Struct, []*Struct or []Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), nil, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Create(v)
	})
}

func (g GormCrud) Update(ctx context.Context, model interface{}, v interface{}, options...Option) error {
	if !g.isSupportedValue(model, singleModelWrite) {
		return ErrorInvalidCrudParam.
			WithMessage(errTmplInvalidCrudModel, v, "Update", "*Struct or Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), nil, options, modelFunc(model), func(db *gorm.DB) *gorm.DB {
		// note we use the actual model instead of template g.model
		return db.Updates(v)
	})
}

func (g GormCrud) Delete(ctx context.Context, v interface{}, options...Option) error {
	if !g.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidCrudParam.WithMessage(errTmplInvalidCrudValue, v, "Delete", "*Struct, []Struct or []*Struct")
	}

	return execute(ctx, g.GormApi.DB(ctx), nil, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Delete(v)
	})
}

func (g GormCrud) DeleteBy(ctx context.Context, condition Condition, options...Option) error {
	return execute(ctx, g.GormApi.DB(ctx), condition, options, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		return db.Delete(g.model)
	})
}

func (g GormCrud) Truncate(ctx context.Context) error {
	return execute(ctx, g.GormApi.DB(ctx), nil, nil, modelFunc(g.model), func(db *gorm.DB) *gorm.DB {
		if e := db.Statement.Parse(g.model); e != nil {
			_ = db.AddError(ErrorInvalidCrudModel.WithMessage("unable to parse table name for model %T", g.model))
			return db
		}
		table := interface{}(db.Statement.TableExpr)
		if db.Statement.TableExpr == nil {
			table = db.Statement.Table
		}
		return db.Exec(fmt.Sprintf(`TRUNCATE TABLE %s CASCADE`,  db.Statement.Quote(table)))
	})
}

/*******************
	Helpers
 *******************/

func modelFunc(m interface{}) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Model(m)
	}
}

func execute(_ context.Context, db *gorm.DB, condition Condition, options []Option, preOptsFn, fn func(*gorm.DB) *gorm.DB) error {
	if preOptsFn != nil {
		options = append(options, priorityOption{
			order:   order.Highest,
			wrapped: preOptsFn,
		})
	}

	var e error
	if db, e = applyOptions(db, options); e != nil {
		return e
	}

	if db, e = applyCondition(db, condition); e != nil {
		return e
	}

	if r := fn(db); r.Error != nil {
		return r.Error
	}
	return nil
}

func optsToDBFuncs(opts []Option) ([]func(*gorm.DB)*gorm.DB, error) {
	scopes := make([]func(*gorm.DB)*gorm.DB, 0, len(opts))
	for _, v := range opts {
		switch rv := reflect.ValueOf(v); rv.Kind() {
		case reflect.Slice, reflect.Array:
			size := rv.Len()
			slice := make([]Option, size)
			for i := 0; i < size; i++ {
				slice[i] = rv.Index(i).Interface()
			}
			sub, e := optsToDBFuncs(slice)
			if e != nil {
				return nil, e
			}
			scopes = append(scopes, sub...)
		default:
			switch opt := v.(type) {
			case priorityOption:
				sub, e := optsToDBFuncs([]Option{opt.wrapped})
				if e != nil {
					return nil, e
				}
				scopes = append(scopes, sub...)
			case delayedOption:
				//SuppressWarnings go:S1871 we can use "opt.wrapped" here, but SONAR doesn't understand type switching
				sub, e := optsToDBFuncs([]Option{opt.wrapped})
				if e != nil {
					return nil, e
				}
				scopes = append(scopes, sub...)
			case gormOptions:
				scopes = append(scopes, opt)
			case func(*gorm.DB) *gorm.DB:
				scopes = append(scopes, opt)
			default:
				return nil, ErrorUnsupportedOptions.WithMessage("unsupported Option %T", v)
			}
		}
	}
	return scopes, nil
}

func applyOptions(db *gorm.DB, opts []Option) (*gorm.DB, error) {
	if len(opts) == 0 {
		return db, nil
	}

	order.SortStable(opts, order.UnorderedMiddleCompare)

	funcs, e := optsToDBFuncs(opts)
	if e != nil {
		return nil, e
	}
	// Note, we choose to apply funcs by our self instead of using db.Scopes(...),
	// because we don't want to confuse GORM with other scopes added else where
	for _, fn := range funcs {
		db = fn(db)
	}
	return db, db.Error
}

func conditionToDBFuncs(condition Condition) ([]func(*gorm.DB)*gorm.DB, error) {
	var scopes []func(*gorm.DB)*gorm.DB
	switch cv := reflect.ValueOf(condition); cv.Kind() {
	case reflect.Slice, reflect.Array:
		size := cv.Len()
		scopes = make([]func(*gorm.DB)*gorm.DB, 0, size)
		for i := 0; i < size; i++ {
			sub, e := conditionToDBFuncs(cv.Index(i).Interface())
			if e != nil {
				return nil, e
			}
			scopes = append(scopes, sub...)
		}
	default:
		var scope func(*gorm.DB)*gorm.DB
		switch where := condition.(type) {
		case gormOptions:
			scope = where
		case func(*gorm.DB) *gorm.DB:
			scope = where
		case clause.Where:
			scope = func(db *gorm.DB) *gorm.DB {
				return db.Clauses(where)
			}
		default:
			scope = func(db *gorm.DB) *gorm.DB {
				return db.Where(where)
			}
		}
		scopes = []func(*gorm.DB)*gorm.DB{scope}
	}

	return scopes, nil
}

func applyCondition(db *gorm.DB, condition Condition) (*gorm.DB, error) {
	if condition == nil {
		return db, nil
	}

	funcs, e := conditionToDBFuncs(condition)
	if e != nil {
		return nil, e
	}
	// Note, we choose to apply funcs by our self instead of using db.Scopes(...),
	// because we don't want to confuse GORM with other scopes added else where
	for _, fn := range funcs {
		db = fn(db)
	}
	return db, db.Error
}
