package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

const (
	ErrorMessage     = "this endpoint always fail"
	ErrorDescription = "This is a generated error"
)

type EchoResponse struct {
	Headers map[string]string `json:"headers"`
	Form    map[string]string `json:"form"`
	ReqBody json.RawMessage   `json:"body"`
}

type ErrorResponse struct {
	EchoResponse
	SC      int `json:"-"`
	Message string
}

func (r ErrorResponse) Error() string {
	return r.Message
}

func (r ErrorResponse) StatusCode() int {
	return r.SC
}

func (r ErrorResponse) MarshalJSON() ([]byte, error) {
	v := struct {
		EchoResponse
		Error   string       `json:"error"`
		Message string       `json:"message"`
		Desc    string       `json:"error_description"`
		Details EchoResponse `json:"details"`
	}{
		EchoResponse: r.EchoResponse,
		Error:        ErrorMessage,
		Message:      r.Message,
		Desc:         ErrorDescription,
		Details:      EchoResponse{},
	}
	return json.Marshal(v)
}

type MockedController struct{}

func NewMockedController() web.Controller {
	return MockedController{}
}

func (c MockedController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/echo").EndpointFunc(c.Echo).Build(),
		rest.Put("/fail").EndpointFunc(c.Fail).Build(),
	}
}

func (c MockedController) Echo(_ context.Context, req *http.Request) (interface{}, error) {
	return c.echoResponse(req)
}

func (c MockedController) Fail(_ context.Context, req *http.Request) (*EchoResponse, error) {
	echo, e := c.echoResponse(req)
	if e != nil {
		return nil, e
	}
	sc, e := strconv.Atoi(req.Form.Get("sc"))
	if e != nil {
		sc = http.StatusInternalServerError
	}
	return nil, &ErrorResponse{
		EchoResponse: *echo,
		SC:           sc,
		Message:      ErrorMessage,
	}
}

func (c MockedController) echoResponse(req *http.Request) (*EchoResponse, error) {
	ret := EchoResponse{
		Headers: map[string]string{},
		Form:    map[string]string{},
	}

	for k := range req.Header {
		ret.Headers[k] = req.Header.Get(k)
	}

	if e := req.ParseForm(); e != nil {
		return nil, e
	}

	for k := range req.Form {
		ret.Form[k] = req.Form.Get(k)
	}

	data, e := io.ReadAll(req.Body)
	if e != nil {
		return nil, e
	}
	ret.ReqBody = data

	return &ret, nil
}
