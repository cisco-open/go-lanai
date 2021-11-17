package data

import (
	"context"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"net/http"
)

//goland:noinspection GoNameStartsWithPackageName
// WebDataErrorTranslator implements web.ErrorTranslator
type WebDataErrorTranslator struct{}

//goland:noinspection GoNameStartsWithPackageName
func NewWebDataErrorTranslator() ErrorTranslator {
	return WebDataErrorTranslator{}
}

func (WebDataErrorTranslator) Order() int {
	return ErrorTranslatorOrderData
}

func (t WebDataErrorTranslator) Translate(ctx context.Context, err error) error {
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

func (t WebDataErrorTranslator) errorWithStatusCode(_ context.Context, err error, sc int) error {
	return NewErrorWithStatusCode(err.(DataError), sc)
}
