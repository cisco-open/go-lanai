// Package v3 Generated by lanai-cli codegen.
// Derived from contents in openapi contract, path: /my/api/v3/anotherTest
package v3

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"github.com/cisco-open/test-service/pkg/api"
	"go.uber.org/fx"
)

type AnotherTestController struct{}

type anotherTestControllerDI struct {
	fx.In
}

func NewAnotherTestController(di anotherTestControllerDI) web.Controller {
	return &AnotherTestController{}
}

func (c *AnotherTestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.
			New("anothertest-get").
			Get("/api/v3/anotherTest").
			EndpointFunc(c.GetAnotherTest).
			Build(),
		rest.
			New("anothertest-post").
			Post("/api/v3/anotherTest").
			EndpointFunc(c.TestWithNoRequestBody).
			Build(),
	}
}

func (c *AnotherTestController) GetAnotherTest(ctx context.Context, req api.GenericResponse) (int, interface{}, error) {
	return 501, nil, nil
}

func (c *AnotherTestController) TestWithNoRequestBody(ctx context.Context) (int, interface{}, error) {
	return 501, nil, nil
}
