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

type GormErrorTranslator struct{}

func NewGormErrorTranslator() ErrorTranslator {
	return GormErrorTranslator{}
}

func (GormErrorTranslator) Order() int {
	return ErrorTranslatorOrderGorm
}

func (GormErrorTranslator) Translate(_ context.Context, err error) error {
	for k, v := range GormErrorMapping {
		if errors.Is(err, k) {
			return v
		}
	}
	return err
}

