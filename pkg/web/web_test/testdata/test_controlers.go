package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
)

type Controller struct{}

func (c Controller) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/basic/:var").EndpointFunc(StructPtr200).Build(),
	}
}

