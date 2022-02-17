package data

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

var (
	GormErrorMapping = map[error]DataError{
		gorm.ErrRecordNotFound:        NewDataError(ErrorCodeRecordNotFound, gorm.ErrRecordNotFound),
		gorm.ErrInvalidTransaction:    NewDataError(ErrorCodeInvalidTransaction, gorm.ErrInvalidTransaction),
		gorm.ErrNotImplemented:        NewDataError(ErrorCodeInvalidApiUsage, gorm.ErrNotImplemented),
		gorm.ErrMissingWhereClause:    NewDataError(ErrorCodeInvalidSQL, gorm.ErrMissingWhereClause),
		gorm.ErrUnsupportedRelation:   NewDataError(ErrorCodeInvalidSchema, gorm.ErrUnsupportedRelation),
		gorm.ErrPrimaryKeyRequired:    NewDataError(ErrorCodeInvalidSQL, gorm.ErrPrimaryKeyRequired),
		gorm.ErrModelValueRequired:    NewDataError(ErrorCodeOrmMapping, gorm.ErrModelValueRequired),
		gorm.ErrInvalidData:           NewDataError(ErrorCodeOrmMapping, gorm.ErrInvalidData),
		gorm.ErrUnsupportedDriver:     NewDataError(ErrorCodeInternal, gorm.ErrUnsupportedDriver),
		gorm.ErrRegistered:            NewDataError(ErrorCodeInternal, gorm.ErrRegistered), // TODO ??
		gorm.ErrInvalidField:          NewDataError(ErrorCodeInvalidSQL, gorm.ErrInvalidField),
		gorm.ErrEmptySlice:            NewDataError(ErrorCodeIncorrectRecordCount, gorm.ErrEmptySlice),
		gorm.ErrDryRunModeUnsupported: NewDataError(ErrorCodeInvalidApiUsage, gorm.ErrDryRunModeUnsupported),
		gorm.ErrInvalidDB:             NewDataError(ErrorCodeInvalidApiUsage, gorm.ErrInvalidDB),
		gorm.ErrInvalidValue:          NewDataError(ErrorCodeInvalidSQL, gorm.ErrInvalidValue),
		gorm.ErrInvalidValueOfLength:  NewDataError(ErrorCodeInvalidSQL, gorm.ErrInvalidValueOfLength),
	}
)

type DefaultGormErrorTranslator struct {
	ErrorTranslator
}

func (t DefaultGormErrorTranslator) TranslateWithDB(db *gorm.DB) error {
	if db.Error == nil {
		return nil
	}
	err := t.Translate(db.Statement.Context, db.Error)
	//nolint:errorlint
	switch e := err.(type) {
	case DataError:
		switch {
		case db.Statement != nil:
			return e.WithDetails(db.Statement)
		}
	}
	return err
}

// gormErrorTranslator implements GormErrorTranslator and ErrorTranslator
type gormErrorTranslator struct{}

func NewGormErrorTranslator() ErrorTranslator {
	return DefaultGormErrorTranslator{
		ErrorTranslator: gormErrorTranslator{},
	}
}

func (gormErrorTranslator) Order() int {
	return ErrorTranslatorOrderGorm
}

func (gormErrorTranslator) Translate(_ context.Context, err error) error {
	for k, v := range GormErrorMapping {
		if errors.Is(err, k) {
			return v
		}
	}
	return err
}


