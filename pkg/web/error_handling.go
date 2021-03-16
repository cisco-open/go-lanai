package web

import (
	"context"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-playground/validator/v10"
	"net/http"
)

/*************************
	error translation
 *************************/
func newErrorEncoder(encoder httptransport.ErrorEncoder, translators ...ErrorTranslator) httptransport.ErrorEncoder {
	return func(ctx context.Context, err error, rw http.ResponseWriter) {
		for _, t := range translators {
			err = t.Translate(ctx, err)
		}
		encoder(ctx, err, rw)
	}
}

type defaultErrorTranslator struct {}

func (i defaultErrorTranslator) Translate(_ context.Context, err error) error {
	switch err.(type) {
	case validator.ValidationErrors:
		return ValidationErrors{err.(validator.ValidationErrors)}
	case StatusCoder, HttpError:
		return err
	default:
		return HttpError{error: err, SC: http.StatusInternalServerError}
	}
}

func newDefaultErrorTranslator() defaultErrorTranslator {
	return defaultErrorTranslator{}
}

/*****************************
	Error Encoder
******************************/
func JsonErrorEncoder() httptransport.ErrorEncoder {
	return jsonErrorEncoder
}

func jsonErrorEncoder(c context.Context, err error, w http.ResponseWriter) {
	if _,ok := err.(json.Marshaler); !ok {
		err = NewHttpError(0, err)
	}
	httptransport.DefaultErrorEncoder(c, err, w)
}