// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package httpclient

import (
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"fmt"
	"net/http"
)

const (
	// Reserved http client reserved error range
	Reserved = 0xcc << ReservedOffset
)

// All "Type" values are used as mask
const (
	_                     = iota
	ErrorTypeCodeInternal = Reserved + iota<<ErrorTypeOffset
	ErrorTypeCodeTransport
	ErrorTypeCodeResponse
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeInternal
const (
	_                        = iota
	ErrorSubTypeCodeInternal = ErrorTypeCodeInternal + iota<<ErrorSubTypeOffset
	ErrorSubTypeCodeDiscovery
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeTransport
const (
	_                       = iota
	ErrorSubTypeCodeTimeout = ErrorTypeCodeTransport + iota<<ErrorSubTypeOffset
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeResponse
const (
	_                          = iota
	ErrorSubTypeCodeServerSide = ErrorTypeCodeResponse + iota<<ErrorSubTypeOffset
	ErrorSubTypeCodeClientSide
	ErrorSubTypeCodeMedia
)

// ErrorSubTypeCodeInternal
const (
	_                 = iota
	ErrorCodeInternal = ErrorSubTypeCodeInternal + iota
)

// ErrorSubTypeCodeDiscovery
const (
	_                      = iota
	ErrorCodeDiscoveryDown = ErrorSubTypeCodeDiscovery + iota
	ErrorCodeNoEndpointFound
)

// ErrorSubTypeCodeTimeout
const (
	_                      = iota
	ErrorCodeServerTimeout = ErrorSubTypeCodeTimeout + iota
)

// ErrorSubTypeCodeMedia
const (
	_                  = iota
	ErrorCodeMediaType = ErrorSubTypeCodeMedia + iota
	ErrorCodeSerialization
)

// ErrorSubTypeCodeClientSide
const (
	_                          = iota
	ErrorCodeGenericClientSide = ErrorSubTypeCodeClientSide + iota
	ErrorCodeUnauthorized
	ErrorCodeForbidden
)

// ErrorSubTypeCodeServerSide
const (
	_                          = iota
	ErrorCodeGenericServerSide = ErrorSubTypeCodeServerSide + iota
)

// ErrorTypes, can be used in errors.Is
var (
	ErrorCategoryHttpClient = NewErrorCategory(Reserved, errors.New("error type: http client"))
	ErrorTypeInternal       = NewErrorType(ErrorTypeCodeInternal, errors.New("error type: internal"))
	ErrorTypeTransport      = NewErrorType(ErrorTypeCodeTransport, errors.New("error type: http transport"))
	ErrorTypeResponse       = NewErrorType(ErrorTypeCodeResponse, errors.New("error type: error status code"))

	ErrorSubTypeInternalError = NewErrorSubType(ErrorSubTypeCodeInternal, errors.New("error sub-type: internal"))
	ErrorSubTypeDiscovery     = NewErrorSubType(ErrorSubTypeCodeDiscovery, errors.New("error sub-type: discover"))
	ErrorSubTypeTimeout       = NewErrorSubType(ErrorSubTypeCodeTimeout, errors.New("error sub-type: server timeout"))
	ErrorSubTypeServerSide    = NewErrorSubType(ErrorSubTypeCodeServerSide, errors.New("error sub-type: server side"))
	ErrorSubTypeClientSide    = NewErrorSubType(ErrorSubTypeCodeClientSide, errors.New("error sub-type: client side"))
	ErrorSubTypeMedia         = NewErrorSubType(ErrorSubTypeCodeMedia, errors.New("error sub-type: server timeout"))
)

// Concrete error, can be used in errors.Is for exact match

var (
	ErrorDiscoveryDown = NewError(ErrorCodeDiscoveryDown, "service discovery is not available")
)

func init() {
	Reserve(ErrorCategoryHttpClient)
}

type ErrorResponseBody interface {
	Error() string
	Message() string
	Details() map[string]string
}

type ErrorResponse struct {
	http.Response
	RawBody []byte
	Body    ErrorResponseBody
}

func (er ErrorResponse) Error() string {
	if er.Body == nil {
		return er.Status
	}
	return er.Body.Error()
}

func (er ErrorResponse) Message() string {
	if er.Body == nil {
		return er.Status
	}
	return er.Body.Message()
}

// Error can optionally store *http.Response's status code, headers and body
type Error struct {
	CodedError
	Response *ErrorResponse
}

func (e Error) Error() string {
	return e.String()
}

func (e Error) String() string {
	switch {
	case e.Response == nil:
		return e.CodedError.Error()
	case e.Response.Body == nil:
		return fmt.Sprintf("error HTTP response [%s]", e.Response.Status)
	default:
		return fmt.Sprintf("error HTTP response [%s]: %s", e.Response.Status, e.Response.Message())
	}
}

func (e Error) WithMessage(msg string, args ...interface{}) *Error {
	return newError(NewCodedError(e.CodedError.Code(), fmt.Errorf(msg, args...)), e.Response)
}

/**********************
	Constructors
 **********************/
func newError(codedErr *CodedError, errResp *ErrorResponse) *Error {
	err := &Error{
		CodedError: *codedErr,
		Response:   errResp,
	}
	return err
}

func NewError(code int64, e interface{}, causes ...interface{}) *Error {
	return newError(NewCodedError(code, e, causes...), nil)
}

// NewErrorWithResponse create a Error with ErrorResponse.
// if given "e" is an ErrorResponseBody, it saved into ErrorResponse
func NewErrorWithResponse(code int64, e interface{}, resp *http.Response, rawBody []byte, causes ...interface{}) *Error {
	body, _ := e.(ErrorResponseBody)
	coded := NewCodedError(code, e, causes...)
	errResp := &ErrorResponse{
		Response: *resp,
		RawBody:  rawBody,
		Body:     body,
	}
	return newError(coded, errResp)
}

// NewErrorWithStatusCode create a Error with ErrorResponse, and choose error code based on status code
// if given "e" is an ErrorResponseBody, it saved into ErrorResponse
func NewErrorWithStatusCode(e interface{}, resp *http.Response, rawBody []byte, causes ...interface{}) *Error {
	var code int64
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		code = ErrorCodeUnauthorized
	case resp.StatusCode == http.StatusForbidden:
		code = ErrorCodeForbidden
	case resp.StatusCode >= 400 && resp.StatusCode <= 499:
		code = ErrorCodeGenericClientSide
	case resp.StatusCode >= 500 && resp.StatusCode <= 599:
		code = ErrorCodeGenericServerSide
	default:
		return NewError(ErrorCodeInternal, "attempt to create response error with non error status code %d", resp.StatusCode)
	}
	return NewErrorWithResponse(code, e, resp, rawBody, causes...)
}

func NewInternalError(value interface{}, causes ...interface{}) *Error {
	return NewError(ErrorSubTypeCodeInternal, value, causes...)
}

func NewDiscoveryDownError(value interface{}, causes ...interface{}) *Error {
	return NewError(ErrorCodeDiscoveryDown, value, causes...)
}

func NewNoEndpointFoundError(value interface{}, causes ...interface{}) *Error {
	return NewError(ErrorCodeNoEndpointFound, value, causes...)
}

func NewServerTimeoutError(value interface{}, causes ...interface{}) *Error {
	return NewError(ErrorCodeServerTimeout, value, causes...)
}

func NewMediaTypeError(value interface{}, resp *http.Response, rawBody []byte, causes ...interface{}) *Error {
	return NewErrorWithResponse(ErrorCodeMediaType, value, resp, rawBody, causes...)
}

func NewSerializationError(value interface{}, resp *http.Response, rawBody []byte, causes ...interface{}) *Error {
	return NewErrorWithResponse(ErrorCodeSerialization, value, resp, rawBody, causes...)
}

func NewRequestSerializationError(value interface{}, causes ...interface{}) *Error {
	return NewError(ErrorCodeSerialization, value, causes...)
}
