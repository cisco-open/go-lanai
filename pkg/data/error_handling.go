package data

import (
	"context"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"net/http"
)

// ErrorTranslator redefines web.ErrorTranslator and order.Ordered
// having this redefinition is to break dependency between data and web package
type ErrorTranslator interface {
	order.Ordered
	Translate(ctx context.Context, err error) error
}

// DriverErrorTranslator is a composite error translator that invoke driver specific translators provided in other packages
// e.g. cockroach.ErrorTranslator
type DriverErrorTranslator []ErrorTranslator

func NewDriverErrorTranslator(translators ...ErrorTranslator) ErrorTranslator {
	order.SortStable(translators, order.OrderedFirstCompare)
	return DriverErrorTranslator(translators)
}

func (DriverErrorTranslator) Order() int {
	return ErrorTranslatorOrderDriver
}

func (t DriverErrorTranslator) Translate(ctx context.Context, err error) error {
	for _, translator := range t {
		err = translator.Translate(ctx, err)
	}
	return err
}

//goland:noinspection GoNameStartsWithPackageName
type DataErrorTranslator struct{}

//goland:noinspection GoNameStartsWithPackageName
func NewDataErrorTranslator() ErrorTranslator {
	return DataErrorTranslator{}
}

func (DataErrorTranslator) Order() int {
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

func (t DataErrorTranslator) errorWithStatusCode(_ context.Context, err error, sc int) error {
	if _, ok := err.(*DataError); !ok {
		return NewDataError(err.(errorutils.ErrorCoder).Code(), err).WithStatusCode(sc)
	}

	return err.(*DataError).WithStatusCode(sc)
}
