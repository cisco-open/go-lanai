package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

func newJsonResponse(req *JsonRequest) *JsonResponse {
	return &JsonResponse{
		UriVar:     req.UriVar,
		QueryVar:   req.QueryVar,
		HeaderVar:  req.HeaderVar,
		JsonString: req.JsonString,
		JsonInt:    req.JsonInt,
	}
}

/*********************
	Supported
 *********************/

func StructPtr200(_ context.Context, req *JsonRequest) (*JsonResponse, error) {
	return newJsonResponse(req), nil
}

func Struct200(_ context.Context, req JsonRequest) (JsonResponse, error) {
	return *newJsonResponse(&req), nil
}

func StructPtr201(_ context.Context, req *JsonRequest) (int, *JsonResponse, error) {
	return http.StatusCreated, newJsonResponse(req), nil
}

func Struct201(_ context.Context, req JsonRequest) (int, JsonResponse, error) {
	return http.StatusCreated, *newJsonResponse(&req), nil
}

func StructPtr201WithHeader(_ context.Context, req *JsonRequest) (http.Header, int, *JsonResponse, error) {
	header := http.Header{}
	header.Set("X-VAR", req.HeaderVar)
	return header, http.StatusCreated, newJsonResponse(req), nil
}

func Struct201WithHeader(_ context.Context, req JsonRequest) (http.Header, int, JsonResponse, error) {
	header := http.Header{}
	header.Set("X-VAR", req.HeaderVar)
	return header, http.StatusCreated, *newJsonResponse(&req), nil
}

func Raw(ctx context.Context, req *http.Request) (interface{}, error) {
	gc := web.GinContext(ctx)
	var jsonReq JsonRequest
	_ = gc.BindUri(&jsonReq)
	_ = binding.Query.Bind(req, &jsonReq)
	_ = binding.Header.Bind(req, &jsonReq)
	_ = binding.JSON.Bind(req, &jsonReq)
	return newJsonResponse(&jsonReq), nil
}

func NoRequest(_ context.Context) (*JsonResponse, error) {
	return &JsonResponse{}, nil
}

/*********************
	Not Supported
 *********************/

func MissingResponse(_ context.Context, _ *JsonRequest) error {
	return nil
}

func MissingError(_ context.Context, _ *JsonRequest) *JsonResponse {
	return nil
}

func MissingContext(_ *JsonRequest) (*JsonResponse, error) {
	return nil, nil
}

func WrongErrorPosition(_ context.Context, _ *JsonRequest) (error, *JsonResponse, int) {
	return nil, nil, 0
}

func WrongContextPosition( _ *JsonRequest, _ context.Context) (*JsonResponse, int, error) {
	return nil, 0, nil
}

func ExtraInput(_ context.Context, _ *JsonRequest, _ string) (*JsonResponse, error) {
	return nil, nil
}
