package repo

import "C"
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
	case []Condition, clause.Where:
		funcs, e = conditionToDBFuncs(Condition(i))
	case gormOptions:
		funcs, e = optsToDBFuncs([]Option{v})
	default:
		funcs, e = conditionToDBFuncs(Condition(i))
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

// Where is a Condition that directly bridge parameters to (*gorm.DB).Where()
func Where(query interface{}, args ...interface{}) Condition {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
}

// Joins is an Option for Find* operations, typically used to populate "ToOne" relationship using JOIN clause
// e.g. CrudRepository.FindById(ctx, &user, Joins("Status"))
//
// When used on "ToMany", JOIN query is usually used instead of field
// e.g.	CrudRepository.FindById(ctx, &user, Joins("JOIN address ON address.user_id = users.id AND address.country = ?", "Canada"))
func Joins(query string, args ...interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, args...)
	})
}

// Preload is an Option for Find* operations, typically used to populate relationship fields using separate queries
// e.g.
//		CrudRepository.FindAll(ctx, &user, Preload("Roles.Permissions"))
// 		CrudRepository.FindAll(ctx, &user, Preload("Roles", "role_name NOT IN (?)", "excluded"))
func Preload(query string, args ...interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args...)
	})
}

// Omit is an Option specifying fields that you want to ignore when creating, updating and querying.
// When supported by gorm.io, this Option is a direct bridge to (*gorm.DB).Omit().
// Please see https://gorm.io/docs/ for detailed usage
func Omit(fields ...string) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Omit(fields...)
	})
}

// Select is an Option specify fields that you want when querying, creating, updating.
// This Option has different meaning when used for different operations (query vs create vs update vs save vs delete)
// When supported by gorm.io, this Option is a direct bridge to (*gorm.DB).Select().
// // Please see https://gorm.io/docs/ for detailed usage
func Select(query interface{}, args ...interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Select(query, args...)
	})
}

// Page is an Option specifying pagination when retrieve records from database
// page: page number started with 0
// size: page size (# of records per page)
// e.g.
//		CrudRepository.FindAll(ctx, &user, Page(2, 10))
//		CrudRepository.FindAllBy(ctx, &user, Where(...), Page(2, 10))
func Page(page, size int) Option {
	offset := page * size
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(size)
	})
}

// Sort is an Option specifying order when retrieve records from database by using column.
// This Option is typically used together with Page option
// When supported by gorm.io, this Option is a direct bridge to (*gorm.DB).Order()
// e.g.
//		CrudRepository.FindAll(ctx, &user, Page(2, 10), Sort("name DESC"))
//		CrudRepository.FindAllBy(ctx, &user, Where(...), Page(2, 10), Sort(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true}))
func Sort(value interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
	})
}

// SortBy an Option similar to Sort, but specifying model's field name
// This Option also support order by direct "ToOne" relation's field when used together with Joins.
// e.g.
//		CrudRepository.FindAll(ctx, &user, Joins("Profile"), Page(2, 10), SortBy("Profile.FirstName", false))
//		CrudRepository.FindAllBy(ctx, &user, Where(...), Page(2, 10), SortBy("Username", true))
func SortBy(fieldName string, desc bool) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		if e := requireSchema(db); e != nil {
			_ = db.AddError(ErrorUnsupportedOptions.WithMessage("SortBy not supported in this usage: %v", e))
			return db
		}
		col, e := toColumn(db.Statement.Schema, fieldName)
		if e != nil {
			_ = db.AddError(ErrorUnsupportedOptions.
				WithMessage("SortBy error: %v", e))
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
