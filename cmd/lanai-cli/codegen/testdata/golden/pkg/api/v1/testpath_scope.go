// Package v1 Generated by lanai_cli codegen. DO NOT EDIT
// Derived from path: /my/api/v1/testpath/{scope}
package v1

import (
	"cto-github.cisco.com/NFV-BU/test-service/pkg/api"
)

type DeleteTestPathRequest struct {
	Scope     string `uri:"scope" binding:"required,regexA79C5"`
	TestParam string `form:"testParam" binding:"omitempty,max=128"`
}

type DeleteTestPathResponse struct {
	Id string `json:"id"`
	api.GenericResponse
}

type TestpathScopeGetRequest struct {
	Scope string `uri:"scope" binding:"required,regexA79C5"`
}

type PostTestPathRequest struct {
	Scope string `uri:"scope" binding:"required,regexA397E"`
}
