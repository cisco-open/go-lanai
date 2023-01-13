// Generated by lanai_cli codegen. DO NOT EDIT
package v2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	v2 "cto-github.cisco.com/NFV-BU/test-service/pkg/service/v2"
	"go.uber.org/fx"
)

type TestpathController struct {
	testpathService v2.TestpathService
}

type testpathControllerDI struct {
	fx.In
	TestpathService v2.TestpathService
}

func NewTestpathController(di testpathControllerDI) web.Controller {
	return &TestpathController{
		testpathService: di.TestpathService,
	}
}

func (c *TestpathController) Mappings() []web.Mapping {
	return []web.Mapping{}
}