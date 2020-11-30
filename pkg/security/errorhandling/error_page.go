package errorhandling

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"errors"
	"fmt"
	"net/http"
)

func ErrorWithStatus(ctx context.Context, _ web.EmptyRequest) (int, *template.ModelView, error) {
	s := session.Get(ctx)
	if s == nil {
		err := fmt.Errorf("error message not available")
		return http.StatusInternalServerError, nil, err
	}

	code, codeOk := s.Flash(FlashKeyPreviousStatusCode).(int)
	if !codeOk {
		code = 500
	}

	err, errOk := s.Flash(FlashKeyPreviousError).(error)
	if !errOk {
		err = errors.New("unknown error")
	}

	return 0, nil, web.NewHttpError(code, err)
}
