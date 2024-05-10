package data

import (
	"context"
	"errors"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/webtest"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
	"time"
)

func TestTranslateDataErrorForHttpResponse(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithTimeout(time.Minute),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewTestTranslateDateErrorController),
			web.FxErrorTranslatorProviders(webDataErrorTranslator),
		),
		test.GomegaSubTest(SubTestDataErrorTranslation(), "TestDataErrorTranslation"),
	)
}

func SubTestDataErrorTranslation() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		expect := map[string]int{
			"ErrorCodeRecordNotFound":      http.StatusNotFound,
			"ErrorIncorrectRecordCount":    http.StatusNotFound,
			"ErrorCodeConstraintViolation": http.StatusConflict,
			"ErrorCodeInvalidSQL":          http.StatusBadRequest,
			"ErrorCodeQueryTimeout":        http.StatusRequestTimeout,
			"ErrorCodePessimisticLocking":  http.StatusServiceUnavailable,
		}
		for code, status := range expect {
			req := webtest.NewRequest(ctx, http.MethodGet, "/translate", nil,
				webtest.Queries("error_code", code),
			)
			resp := webtest.MustExec(ctx, req).Response
			g.Expect(resp.StatusCode).To(Equal(status))
		}
	}
}

func webDataErrorTranslator() web.ErrorTranslator {
	return NewWebDataErrorTranslator()
}

type TranslateRequest struct {
	ErrorCode string `form:"error_code" binding:"required"`
}

type TestTranslateDateErrorController struct {
	errorLookUp map[string]error
}

func NewTestTranslateDateErrorController() *TestTranslateDateErrorController {
	return &TestTranslateDateErrorController{
		errorLookUp: map[string]error{
			"ErrorCodeRecordNotFound":      NewDataError(ErrorCodeRecordNotFound, "record not found"),
			"ErrorIncorrectRecordCount":    NewDataError(ErrorCodeIncorrectRecordCount, "incorrect record count"),
			"ErrorCodeConstraintViolation": NewDataError(ErrorCodeConstraintViolation, "constraint violation"),
			"ErrorCodeInvalidSQL":          NewDataError(ErrorCodeInvalidSQL, "invalid sql"),
			"ErrorCodeQueryTimeout":        NewDataError(ErrorCodeQueryTimeout, "query timeout"),
			"ErrorCodePessimisticLocking":  NewDataError(ErrorCodePessimisticLocking, "pessimistic locking"),
		},
	}
}

func (c *TestTranslateDateErrorController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("error-get").Get("/translate").
			EndpointFunc(c.Translate).Build(),
	}
}

func (c *TestTranslateDateErrorController) Translate(ctx context.Context, r *TranslateRequest) (interface{}, error) {
	err, ok := c.errorLookUp[r.ErrorCode]
	if !ok {
		return nil, errors.New("unknown error code")
	}
	return nil, err
}
