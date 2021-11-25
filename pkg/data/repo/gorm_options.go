package repo

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormOptions func(*gorm.DB) *gorm.DB

/********************
	Util Functions
 ********************/

// MustApplyOptions takes a slice of Option and apply it to the given gorm.DB.
// This function is intended for custom repository implementations.
// The function panic if any Option is not supported type
func MustApplyOptions(db *gorm.DB, opts ...Option) *gorm.DB {
	return AsGormScope(opts)(db)
}

// MustApplyConditions takes a slice of Condition and apply it to the given gorm.DB.
// This function is intended for custom repository implementations.
// The function panic if any Condition is not supported type
func MustApplyConditions(db *gorm.DB, conds ...Condition) *gorm.DB {
	return AsGormScope(conds)(db)
}

// AsGormScope convert following types to a func(*gorm.DB)*gorm.DB:
// - Option or slice of Option
// - Condition or slice of Condition
// - func(*gorm.DB)*gorm.DB (noop)
// - slice of func(*gorm.DB)*gorm.DB
//
// This function is intended for custom repository implementations. The result can be used as "db.Scopes(result...)"
// The function panic on any type not listed above
func AsGormScope(i interface{}) func(*gorm.DB)*gorm.DB {
	var funcs []func(*gorm.DB)*gorm.DB
	var e error
	switch v := i.(type) {
	case func(*gorm.DB)*gorm.DB:
		return v
	case []func(*gorm.DB)*gorm.DB:
		funcs = v
	case []Option:
		funcs, e = optsToDBFuncs(v)
	case Option:
		funcs, e = optsToDBFuncs([]Option{v})
	case Condition:
		funcs, e = conditionToDBFuncs(v)
	case []Condition:
		funcs, e = conditionToDBFuncs(Condition(v))
	default:
		e = fmt.Errorf("unsupported interface [%T] to be converted to GORM scope", i)
	}

	// wrap up
	switch {
	case e != nil:
		panic(e)
	case len(funcs) == 0:
		return func(db *gorm.DB)*gorm.DB {return db}
	case len(funcs) == 1:
		return funcs[0]
	default:
		return func(db *gorm.DB)*gorm.DB {
			for _, fn := range funcs {
				db = fn(db)
			}
			return db
		}
	}
}

/**************************
	Options & Conditions
 **************************/

// WhereCondition generic condition using gorm.DB.Where()
func WhereCondition(query interface{}, args ...interface{}) Condition {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
}

// JoinsOption used for Read
func JoinsOption(query string, args ...interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, args...)
	})
}

// PreloadOption used for Read
func PreloadOption(query string, args ...interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args...)
	})
}

// OmitOption specify fields that you want to ignore when creating, updating and querying
// mostly used for Write
func OmitOption(fields ...string) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Omit(fields...)
	})
}

// SelectOption specify fields that you want when querying, creating, updating
// used for Read and Write
func SelectOption(query interface{}, args ...interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Select(query, args...)
	})
}

// PageOption specify order when retrieve records from database
// page = page number started with 0
// size = page size (# of records per page)
func PageOption(page, size int) Option {
	offset := page * size
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(size)
	})
}

// SortOption specify order when retrieve records from database
// e.g. SortOption("name DESC")
//      SortOption(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func SortOption(value interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
	})
}

// SortByField an Option to sort by given model field
// e.g. SortByField("FieldName", false)
// 		SortByField("OneToOne.FieldName", false)
func SortByField(fieldName string, desc bool) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		if e := requireSchema(db); e != nil {
			_ = db.AddError(ErrorUnsupportedOptions.WithMessage("SortByField not supported in this usage: %v", e))
			return db
		}
		col, e := toColumn(db.Statement.Schema, fieldName)
		if e != nil {
			_ = db.AddError(ErrorUnsupportedOptions.
				WithMessage("SortByField error: %v", e))
			return db
		}
		return db.Order(clause.OrderByColumn{
			Column: *col,
			Desc: desc,
		})
	})
}

/***********************
	Helpers
 ***********************/

func requireSchema(db *gorm.DB) error {
	switch {
	case db.Statement.Schema == nil && db.Statement.Model == nil:
		return fmt.Errorf("schema/model is not available")
	case db.Statement.Schema == nil:
		if e := db.Statement.Parse(db.Statement.Model); e != nil {
			return fmt.Errorf("failed to parse schema - %v", e)
		}
	}
	return nil
}
