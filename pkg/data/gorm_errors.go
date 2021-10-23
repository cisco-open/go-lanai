package data

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

var (
	GormErrorMapping = map[error]*DataError{
		gorm.ErrRecordNotFound:        NewDataError(ErrorCodeRecordNotFound, gorm.ErrRecordNotFound.Error(), gorm.ErrRecordNotFound),
		gorm.ErrInvalidTransaction:    NewDataError(ErrorCodeInvalidTransaction, gorm.ErrInvalidTransaction.Error(), gorm.ErrInvalidTransaction),
		gorm.ErrNotImplemented:        NewDataError(ErrorCodeInvalidApiUsage, gorm.ErrNotImplemented.Error(), gorm.ErrNotImplemented),
		gorm.ErrMissingWhereClause:    NewDataError(ErrorCodeInvalidSQL, gorm.ErrMissingWhereClause.Error(), gorm.ErrMissingWhereClause),
		gorm.ErrUnsupportedRelation:   NewDataError(ErrorCodeInvalidSchema, gorm.ErrUnsupportedRelation.Error(), gorm.ErrUnsupportedRelation),
		gorm.ErrPrimaryKeyRequired:    NewDataError(ErrorCodeInvalidSQL, gorm.ErrPrimaryKeyRequired.Error(), gorm.ErrPrimaryKeyRequired),
		gorm.ErrModelValueRequired:    NewDataError(ErrorCodeOrmMapping, gorm.ErrModelValueRequired.Error(), gorm.ErrModelValueRequired),
		gorm.ErrInvalidData:           NewDataError(ErrorCodeOrmMapping, gorm.ErrInvalidData.Error(), gorm.ErrInvalidData),
		gorm.ErrUnsupportedDriver:     NewDataError(ErrorCodeInternal, gorm.ErrUnsupportedDriver.Error(), gorm.ErrUnsupportedDriver),
		gorm.ErrRegistered:            NewDataError(ErrorCodeInternal, gorm.ErrRegistered.Error(), gorm.ErrRegistered), // TODO ??
		gorm.ErrInvalidField:          NewDataError(ErrorCodeInvalidSQL, gorm.ErrInvalidField.Error(), gorm.ErrInvalidField),
		gorm.ErrEmptySlice:            NewDataError(ErrorCodeIncorrectRecordCount, gorm.ErrEmptySlice.Error(), gorm.ErrEmptySlice),
		gorm.ErrDryRunModeUnsupported: NewDataError(ErrorCodeInvalidApiUsage, gorm.ErrDryRunModeUnsupported.Error(), gorm.ErrDryRunModeUnsupported),
		gorm.ErrInvalidDB:             NewDataError(ErrorCodeInvalidApiUsage, gorm.ErrInvalidDB.Error(), gorm.ErrInvalidDB),
		gorm.ErrInvalidValue:          NewDataError(ErrorCodeInvalidSQL, gorm.ErrInvalidValue.Error(), gorm.ErrInvalidValue),
		gorm.ErrInvalidValueOfLength:  NewDataError(ErrorCodeInvalidSQL, gorm.ErrInvalidValueOfLength.Error(), gorm.ErrInvalidValueOfLength),
	}
)

type GormErrorTranslator struct{}

func NewGormErrorTranslator() ErrorTranslator {
	return GormErrorTranslator{}
}

func (GormErrorTranslator) Order() int {
	return ErrorTranslatorOrderGorm
}

func (GormErrorTranslator) Translate(ctx context.Context, err error) error {
	ret := convertGormError(ctx, err)
	if ret == nil {
		return err
	}
	return ret
}

func convertGormError(ctx context.Context, err error) *DataError {
	for k, v := range GormErrorMapping {
		if errors.Is(err, k) {
			return v
		}
	}
	return nil
}
