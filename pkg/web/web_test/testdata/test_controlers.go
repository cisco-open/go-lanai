package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
)

type JsonRequest struct {
	UriVar     string `uri:"var"`
	QueryVar   string `form:"q"`
	HeaderVar  string `header:"X-VAR"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

type JsonResponse struct {
	UriVar     string `json:"uri"`
	QueryVar   string `json:"q"`
	HeaderVar  string `json:"header"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

type Controller struct{}

func (c Controller) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/basic/:var").EndpointFunc(StructPtr200).Build(),
	}
}

