// Package v2 Generated by lanai_cli codegen. DO NOT EDIT
// Derived from path: /my/api/v2/testArrayUUID
package v2

import (
	"cto-github.cisco.com/NFV-BU/test-service/pkg/api"
)

type TestUUIDInArrayRequest struct {
	Id []string `form:"id" binding:"omitempty,dive,uuid"`
}

type TestRequestBodyWithAllOfRequest struct {
	Id *string `uri:"id" binding:"required"`
	api.RequestBodyWithAllOf
}
