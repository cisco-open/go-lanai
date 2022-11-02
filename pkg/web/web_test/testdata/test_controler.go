package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
)

type BasicRequest struct {
	UriVar     string `uri:"var"`
	QueryVar   string `form:"q"`
	HeaderVar  string `header:"X-VAR"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

type BasicResponse struct {
	UriVar     string `json:"uri"`
	QueryVar   string `form:"q"`
	HeaderVar  string `header:"header"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

type Controller struct{}

func (c Controller) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/basic/:var").EndpointFunc(c.Basic).Build(),
	}
}

func (c Controller) Basic(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
	return &BasicResponse{
		UriVar:     req.UriVar,
		QueryVar:   req.QueryVar,
		HeaderVar:  req.HeaderVar,
		JsonString: req.JsonString,
		JsonInt:    req.JsonInt,
	}, nil
}
