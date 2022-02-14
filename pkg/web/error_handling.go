package web

import (
	"context"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-playground/validator/v10"
	"net/http"
)

/*************************
	ErrorHandlerMapping
 *************************/

type errorTranslateMapping struct {
	name          string
	order         int
	matcher       RouteMatcher
	condition     RequestMatcher
	translateFunc ErrorTranslateFunc
}

func NewErrorTranslateMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, translateFunc ErrorTranslateFunc) ErrorTranslateMapping {
	return &errorTranslateMapping{
		name:          name,
		matcher:       matcher,
		order:         order,
		condition:     cond,
		translateFunc: translateFunc,
	}
}

func (m errorTranslateMapping) Name() string {
	return m.name
}

func (m errorTranslateMapping) Matcher() RouteMatcher {
	return m.matcher
}

func (m errorTranslateMapping) Order() int {
	return m.order
}

func (m errorTranslateMapping) Condition() RequestMatcher {
	return m.condition
}

func (m errorTranslateMapping) TranslateFunc() ErrorTranslateFunc {
	return m.translateFunc
}

/*************************
	Error Translation
 *************************/

func newErrorEncoder(encoder httptransport.ErrorEncoder, translators ...ErrorTranslator) httptransport.ErrorEncoder {
	return func(ctx context.Context, err error, rw http.ResponseWriter) {
		for _, t := range translators {
			err = t.Translate(ctx, err)
		}
		encoder(ctx, err, rw)
	}
}

type mappedErrorTranslator struct {
	order         int
	condition     RequestMatcher
	translateFunc ErrorTranslateFunc
}

func (t mappedErrorTranslator) Order() int {
	return t.order
}

func (t mappedErrorTranslator) Translate(ctx context.Context, err error) error {
	if t.condition != nil {
		if ginCtx := GinContext(ctx); ginCtx != nil {
			if matched, e := t.condition.MatchesWithContext(ctx, ginCtx.Request); e != nil || !matched {
				return err
			}
		}
	}
	return t.translateFunc(ctx, err)
}

func newMappedErrorTranslator(m ErrorTranslateMapping) *mappedErrorTranslator {
	return &mappedErrorTranslator{
		order:         m.Order(),
		condition:     m.Condition(),
		translateFunc: m.TranslateFunc(),
	}
}

type defaultErrorTranslator struct{}

func (i defaultErrorTranslator) Translate(_ context.Context, err error) error {
	//nolint:errorlint
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
	//nolint:errorlint
	if _, ok := err.(json.Marshaler); !ok {
		err = NewHttpError(0, err)
	}
	httptransport.DefaultErrorEncoder(c, err, w)
}
