package repo

import "gorm.io/gorm"

type gormOptions func(*gorm.DB) *gorm.DB

/********************
	Util Functions
 ********************/

// MustApplyOptions takes a slice of Option and apply it to the given gorm.DB.
// This function is intended for custom repository implementations.
// The function panic if any Option is not supported type
func MustApplyOptions(db *gorm.DB, opts []Option) *gorm.DB {
	for _, fn := range GormScopes(opts) {
		db = fn(db)
	}
	return db
}

// GormScopes takes a slice of Option and convert them to GORM scopes (func(*gorm.DB)*gorm.DB)
// The result can be used as "db.Scopes(result...)"
// This function is intended for custom repository implementations.
// The function panic if any Option is not supported type
func GormScopes(opts []Option) (scopes []func(*gorm.DB)*gorm.DB) {
	var e error
	if scopes, e = toScopes(opts); e != nil {
		panic(e)
	}
	return
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

// SortOption specify order when retrieve records from database
// e.g. SortOption("name DESC")
//      SortOption(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func SortOption(value interface{}) Option {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
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

