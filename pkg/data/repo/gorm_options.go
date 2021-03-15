package repo

import "gorm.io/gorm"

type gormOptions func(*gorm.DB) *gorm.DB

// WhereCondition generic condition using gorm.DB.Where()
func WhereCondition(query interface{}, args ...interface{}) CrudCondition {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
}

// JoinsOption used for Read
func JoinsOption(query string, args ...interface{}) CrudOption {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, args...)
	})
}

// JoinsOption used for Read
func PreloadOption(query string, args ...interface{}) CrudOption {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args...)
	})
}

// OmitOption specify fields that you want to ignore when creating, updating and querying
// mostly used for Write
func OmitOption(fields ...string) CrudOption {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Omit(fields...)
	})
}

// SelectOption specify fields that you want when querying, creating, updating
// used for Read and Write
func SelectOption(query string, args ...interface{}) CrudOption {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Select(query, args...)
	})
}

// SortOption specify order when retrieve records from database
// e.g. SortOption("name DESC")
//      SortOption(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func SortOption(value interface{}) CrudOption {
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
	})
}

// PageOption specify order when retrieve records from database
// page = page number started with 0
// size = page size (# of records per page)
func PageOption(page, size int) CrudOption {
	offset := page * size
	return gormOptions(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(size)
	})
}

