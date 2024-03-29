// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package repo

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/data"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

const (
	maxUInt32 = int(^uint32(0))
)

type gormOptions func(*gorm.DB) *gorm.DB

// priorityOption is an option wrapper that guarantee to run before regular options
// priorityOption implements order.PriorityOrdered
type priorityOption struct {
	order   int
	wrapped interface{}
}

func (o priorityOption) PriorityOrder() int {
	return o.order
}

// delayedOption is an option wrapper that guarantee to run after regular options
// delayedOption implements order.Ordered
type delayedOption struct {
	order   int
	wrapped interface{}
}

func (o delayedOption) Order() int {
	return o.order
}

// postExecOptions is applied after SQL is executed. Mostly useful to assert/update result or update error
type postExecOptions func(*gorm.DB) *gorm.DB

/********************
	Util Functions
 ********************/

// MustApplyOptions takes a slice of Option and apply it to the given gorm.DB.
// This function is intended for custom repository implementations.
// The function panic if any Option is not supported type
func MustApplyOptions(db *gorm.DB, opts ...Option) *gorm.DB {
	order.SortStable(opts, order.UnorderedMiddleCompare)
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
func AsGormScope(i interface{}) func(*gorm.DB) *gorm.DB {
	var funcs []func(*gorm.DB) *gorm.DB
	var e error
	switch v := i.(type) {
	case func(*gorm.DB) *gorm.DB:
		return v
	case []func(*gorm.DB) *gorm.DB:
		funcs = v
	case []Option:
		funcs, e = optsToDBFuncs(v)
	case []Condition, clause.Where:
		funcs, e = conditionToDBFuncs(Condition(i))
	case gormOptions, priorityOption, delayedOption:
		funcs, e = optsToDBFuncs([]Option{i})
	case postExecOptions:
		e = ErrorUnsupportedOptions.WithMessage("unsupported Option %T", v)
	default:
		funcs, e = conditionToDBFuncs(Condition(i))
	}

	// wrap up
	switch {
	case e != nil:
		panic(e)
	case len(funcs) == 0:
		return func(db *gorm.DB) *gorm.DB { return db }
	case len(funcs) == 1:
		return funcs[0]
	default:
		return func(db *gorm.DB) *gorm.DB {
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

// Or is a Condition that directly bridge parameters to (*gorm.DB).Or()
func Or(query interface{}, args ...interface{}) Condition {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Or(query, args...)
	})
}

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
	opt := gormOptions(func(db *gorm.DB) *gorm.DB {
		offset := page * size
		if offset < 0 || size <= 0 || offset+size >= maxUInt32 {
			_ = db.AddError(ErrorInvalidPagination)
			return db
		}
		db = db.Offset(offset).Limit(size)

		// add default sorting to ensure fixed order
		sort := clause.OrderByColumn{Column: clause.Column{Name: clause.PrimaryKey}}
		db.Statement.AddClauseIfNotExists(clause.OrderBy{
			Columns: []clause.OrderByColumn{sort},
		})
		return db
	})
	// we want to run this option AFTER any Sort or SortBy
	return delayedOption{
		order:   order.Lowest,
		wrapped: opt,
	}
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
			_ = db.AddError(data.ErrorSortByUnknownColumn.
				WithMessage("SortBy error: %v", e))
			return db
		}
		return db.Order(clause.OrderByColumn{
			Column: *col,
			Desc:   desc,
		})
	})
}

// ErrorOnZeroRows a post-exec option that force repository returns error in case of db.AffectedRows == 0
// This option is useful on certain operations such as CrudRepository.Delete, or CrudRepository.Update,
// which doesn't return error if there is no row get affected/deleted.
func ErrorOnZeroRows() Option {
	// Implementation Note:
	//	Alternative way (probably the proper way) to implement this is to add "after *" callback that reads
	// 	statement's settings and process result accordingly, and ErrorOnZeroRows() can be a regular gormOptions that put a flag
	// 	in statement's settings.
	// 	The callback approach above would allow our ErrorTranslator to intercept the set error. But for this particular
	// 	use case, it doesn't matter because we don't need to translate data.ErrorRecordNotFound error
	return postExecOptions(func(db *gorm.DB) *gorm.DB {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = data.ErrorRecordNotFound
		}
		return db
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
