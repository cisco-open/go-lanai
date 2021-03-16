package data

import (
	"context"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"net/http"
)

type DataErrorTranslator struct {}

func NewDataErrorTranslator() *DataErrorTranslator {
	return &DataErrorTranslator{}
}

func (t DataErrorTranslator) Order() int {
	return ErrorTranslatorOrderData
}

func (t DataErrorTranslator) Translate(ctx context.Context, err error) error {
	if _, ok := err.(errorutils.ErrorCoder); !ok || !errors.Is(err, ErrorCategoryData) {
		return err
	}

	switch {
	case errors.Is(err, ErrorRecordNotFound), errors.Is(err, ErrorIncorrectRecordCount):
		return t.errorWithStatusCode(ctx, err, http.StatusNotFound)
	case errors.Is(err, ErrorSubTypeDataIntegrity):
		return t.errorWithStatusCode(ctx, err, http.StatusConflict)
	case errors.Is(err, ErrorSubTypeQuery):
		return t.errorWithStatusCode(ctx, err, http.StatusBadRequest)
	case errors.Is(err, ErrorSubTypeTimeout):
		return t.errorWithStatusCode(ctx, err, http.StatusRequestTimeout)
	case errors.Is(err, ErrorTypeTransient):
		return t.errorWithStatusCode(ctx, err, http.StatusServiceUnavailable)
	default:
		return t.errorWithStatusCode(ctx, err, http.StatusInternalServerError)
	}
}

func (t DataErrorTranslator) errorWithStatusCode(ctx context.Context, err error, sc int) error {
	if _, ok := err.(*DataError); !ok {
		return NewDataError(err.(errorutils.ErrorCoder).Code(), err.Error(), err).WithStatusCode(sc)
	}

	return err.(*DataError).WithStatusCode(sc)
}
