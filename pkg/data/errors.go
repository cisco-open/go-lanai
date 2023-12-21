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

package data

import (
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"gorm.io/gorm"
)

const (
	// Reserved data reserved reserved error range
	Reserved = 0xdb << ReservedOffset
)

// All "Type" values are used as mask
const (
	_                     = iota
	ErrorTypeCodeInternal = Reserved + iota<<ErrorTypeOffset
	ErrorTypeCodeNonTransient
	ErrorTypeCodeTransient
	ErrorTypeCodeUncategorizedServerSide
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeInternal
const (
	_                        = iota
	ErrorSubTypeCodeInternal = ErrorTypeCodeInternal + iota<<ErrorSubTypeOffset
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeNonTransient
const (
	_                     = iota
	ErrorSubTypeCodeQuery = ErrorTypeCodeNonTransient + iota<<ErrorSubTypeOffset
	ErrorSubTypeCodeApi
	ErrorSubTypeCodeDataRetrieval
	ErrorSubTypeCodeDataIntegrity
	ErrorSubTypeCodeTransaction
	ErrorSubTypeCodeSecurity
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeTransient
const (
	_                           = iota
	ErrorSubTypeCodeConcurrency = ErrorTypeCodeTransient + iota<<ErrorSubTypeOffset
	ErrorSubTypeCodeTimeout
	ErrorSubTypeCodeReplica
)

// ErrorSubTypeCodeInternal
const (
	_                 = iota
	ErrorCodeInternal = ErrorSubTypeCodeInternal + iota
)

// ErrorSubTypeCodeQuery
const (
	_                   = iota
	ErrorCodeInvalidSQL = ErrorSubTypeCodeQuery + iota
	ErrorCodeInvalidPagination
)

// ErrorSubTypeCodeApi
const (
	_                        = iota
	ErrorCodeInvalidApiUsage = ErrorSubTypeCodeApi + iota
	ErrorCodeUnsupportedCondition
	ErrorCodeUnsupportedOptions
	ErrorCodeInvalidCrudModel
	ErrorCodeInvalidCrudParam
)

// ErrorSubTypeCodeDataRetrieval
const (
	_                       = iota
	ErrorCodeRecordNotFound = ErrorSubTypeCodeDataRetrieval + iota
	ErrorCodeOrmMapping
	ErrorCodeIncorrectRecordCount
)

// ErrorSubTypeCodeDataIntegrity
const (
	_                     = iota
	ErrorCodeDuplicateKey = ErrorSubTypeCodeDataIntegrity + iota
	ErrorCodeConstraintViolation
	ErrorCodeInvalidSchema
)

// ErrorSubTypeCodeTransaction
const (
	_                           = iota
	ErrorCodeInvalidTransaction = ErrorSubTypeCodeTransaction + iota
)

// ErrorSubTypeCodeSecurity
const (
	_                             = iota
	ErrorCodeAuthenticationFailed = ErrorSubTypeCodeSecurity + iota
	ErrorCodeFieldOperationDenied
)

// ErrorSubTypeCodeConcurrency
const (
	_                           = iota
	ErrorCodePessimisticLocking = ErrorSubTypeCodeConcurrency + iota
	ErrorCodeOptimisticLocking
)

// ErrorSubTypeCodeTimeout
const (
	_                     = iota
	ErrorCodeQueryTimeout = ErrorSubTypeCodeTimeout + iota
)

// ErrorSubTypeCodeApi
const (
	_                           = iota
	ErrorCodeReplicaUnavailable = ErrorSubTypeCodeReplica + iota
)

// ErrorTypes, can be used in errors.Is
var (
	ErrorCategoryData                = NewErrorCategory(Reserved, errors.New("error type: data"))
	ErrorTypeInternal                = NewErrorType(ErrorTypeCodeInternal, errors.New("error type: internal"))
	ErrorTypeNonTransient            = NewErrorType(ErrorTypeCodeNonTransient, errors.New("error type: non-transient"))
	ErrorTypeTransient               = NewErrorType(ErrorTypeCodeTransient, errors.New("error type: transient"))
	ErrorTypeUnCategorizedServerSide = NewErrorType(ErrorTypeCodeUncategorizedServerSide, errors.New("error type: uncategorized server-side"))

	ErrorSubTypeInternalError = NewErrorSubType(ErrorSubTypeCodeInternal, errors.New("error sub-type: internal"))

	ErrorSubTypeQuery         = NewErrorSubType(ErrorSubTypeCodeQuery, errors.New("error sub-type: query"))
	ErrorSubTypeApi           = NewErrorSubType(ErrorSubTypeCodeApi, errors.New("error sub-type: api"))
	ErrorSubTypeDataRetrieval = NewErrorSubType(ErrorSubTypeCodeDataRetrieval, errors.New("error sub-type: retrieval"))
	ErrorSubTypeDataIntegrity = NewErrorSubType(ErrorSubTypeCodeDataIntegrity, errors.New("error sub-type: integrity"))
	ErrorSubTypeTransaction   = NewErrorSubType(ErrorSubTypeCodeTransaction, errors.New("error sub-type: transaction"))
	ErrorSubTypeSecurity      = NewErrorSubType(ErrorSubTypeCodeSecurity, errors.New("error sub-type: security"))

	ErrorSubTypeConcurrency = NewErrorSubType(ErrorSubTypeCodeConcurrency, errors.New("error sub-type: concurency"))
	ErrorSubTypeTimeout     = NewErrorSubType(ErrorSubTypeCodeTimeout, errors.New("error sub-type: timeout"))
	ErrorSubTypeReplica     = NewErrorSubType(ErrorSubTypeCodeReplica, errors.New("error sub-type: replica"))
)

// Concrete error, can be used in errors.Is for exact match
var (
	ErrorSortByUnknownColumn  = NewDataError(ErrorCodeOrmMapping, "SortBy column unknown")
	ErrorRecordNotFound       = NewDataError(ErrorCodeRecordNotFound, gorm.ErrRecordNotFound)
	ErrorIncorrectRecordCount = NewDataError(ErrorCodeIncorrectRecordCount, "incorrect record count")
	ErrorDuplicateKey         = NewDataError(ErrorCodeDuplicateKey, "duplicate key")
)

func init() {
	Reserve(ErrorCategoryData)
}

//goland:noinspection GoNameStartsWithPackageName
type DataError interface {
	error
	NestedError
	Details() interface{}
	WithDetails(interface{}) DataError
	WithMessage(msg string, args ...interface{}) DataError
}

// dataError implements DataError and errorutils.Unwrapper
//goland:noinspection GoNameStartsWithPackageName
type dataError struct {
	*CodedError
	details interface{}
}

func (e dataError) Details() interface{} {
	return e.details
}

func (e dataError) WithDetails(details interface{}) DataError {
	return dataError{
		CodedError: e.CodedError,
		details:    details,
	}
}

func (e dataError) WithMessage(msg string, args ...interface{}) DataError {
	return dataError{
		CodedError: e.CodedError.WithMessage(msg, args...),
		details:    e.details,
	}
}

func (e dataError) Unwrap() error {
	cause := e.Cause()
	//nolint:errorlint
	switch cause.(type) {
	case NestedError:
		return e.RootCause()
	default:
		return cause
	}
}

// webDataError also implements web.StatusCoder
//goland:noinspection GoNameStartsWithPackageName
type webDataError struct {
	dataError
	SC int
}

func (e webDataError) StatusCode() int {
	return e.SC
}

func (e webDataError) WithStatusCode(sc int) DataError {
	return webDataError{dataError: e.dataError, SC: sc}
}

func (e webDataError) WithMessage(msg string, args ...interface{}) DataError {
	return webDataError{dataError: e.dataError.WithMessage(msg, args...).(dataError), SC: e.SC}
}

/**********************
	Constructors
 **********************/

func NewDataError(code int64, e interface{}, causes ...interface{}) DataError {
	return &dataError{
		CodedError: NewCodedError(code, e, causes...),
	}
}

func NewErrorWithStatusCode(err error, sc int) DataError {
	//nolint:errorlint // we don't consider wrapped error here.
	switch e := err.(type) {
	case dataError:
		return &webDataError{dataError: e, SC: sc}
	case CodedError:
		return &webDataError{dataError: dataError{CodedError: &e}, SC: sc}
	case *CodedError:
		return &webDataError{dataError: dataError{CodedError: e}, SC: sc}
	case ErrorCoder:
		return &webDataError{dataError: *NewDataError(e.Code(), e).(*dataError), SC: sc}
	default:
		return &webDataError{dataError: *NewDataError(ErrorSubTypeCodeInternal, e).(*dataError), SC: sc}
	}
}

func NewInternalError(value interface{}, causes ...interface{}) DataError {
	return NewDataError(ErrorSubTypeCodeInternal, value, causes...)
}

func NewRecordNotFoundError(value interface{}, causes ...interface{}) DataError {
	return NewDataError(ErrorCodeRecordNotFound, value, causes...)
}

func NewConstraintViolationError(value interface{}, causes ...interface{}) DataError {
	return NewDataError(ErrorCodeConstraintViolation, value, causes...)
}

func NewDuplicateKeyError(value interface{}, causes ...interface{}) DataError {
	return NewDataError(ErrorCodeDuplicateKey, value, causes...)
}
