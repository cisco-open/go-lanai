package data

import (
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"fmt"
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
	ErrorRecordNotFound       = NewDataError(ErrorCodeRecordNotFound, "record not found")
	ErrorIncorrectRecordCount = NewDataError(ErrorCodeIncorrectRecordCount, "incorrect record count")
)

func init() {
	Reserve(ErrorCategoryData)
}

// DataError also implements web.StatusCoder
type DataError struct {
	CodedError
	SC int
}

func (e DataError) StatusCode() int {
	return e.SC
}

func (e *DataError) WithStatusCode(sc int) *DataError {
	return &DataError{CodedError: e.CodedError, SC: sc}
}

func (e DataError) WithMessage(msg string, args ...interface{}) *DataError {
	return newDataError(NewCodedError(e.CodedError.Code(), fmt.Errorf(msg, args...)))
}

/**********************
	Constructors
 **********************/
func newDataError(codedErr *CodedError) *DataError {
	return &DataError{
		CodedError: *codedErr,
	}
}

func NewDataError(code int64, e interface{}, causes ...interface{}) *DataError {
	return newDataError(NewCodedError(code, e, causes...))
}

func NewInternalError(value interface{}, causes ...interface{}) *DataError {
	return NewDataError(ErrorSubTypeCodeInternal, value, causes...)
}

func NewRecordNotFoundError(value interface{}, causes ...interface{}) *DataError {
	return NewDataError(ErrorCodeRecordNotFound, value, causes...)
}

func NewConstraintViolationError(value interface{}, causes ...interface{}) *DataError {
	return NewDataError(ErrorCodeConstraintViolation, value, causes...)
}

func NewDuplicateKeyError(value interface{}, causes ...interface{}) *DataError {
	return NewDataError(ErrorCodeDuplicateKey, value, causes...)
}
