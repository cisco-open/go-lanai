package web

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
)

const (
	templateInvalidMvcHandlerFunc = "invalid MVC handler function signature: %v, but got <%v>"
	templateValidationFieldError = "validation failed on '%s' with criteria '%s'"
)

/*****************************
	Error definitions
******************************/

// mapping related
type errorInvalidMvcHandlerFunc struct {
	reason error
	target *reflect.Value
}

func (e *errorInvalidMvcHandlerFunc) Error() string {
	return fmt.Sprintf(templateInvalidMvcHandlerFunc, e.reason.Error(), e.target.Type())
}

/**************************
	Generic Http Error
***************************/
type HttpErrorResponse struct {
	StatusText string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Details map[string]string`json:"details,omitempty"`
}

type HttpError struct {
	error
	SC int
	H  http.Header
}

// TODO consider implement Unwrap
// json.Marshaler
func (e HttpError) MarshalJSON() ([]byte, error) {
	if original,ok := e.error.(json.Marshaler); ok {
		return original.MarshalJSON()
	}
	err := &HttpErrorResponse{
		StatusText: http.StatusText(e.StatusCode()),
		Message: e.Error(),
	}
	return json.Marshal(err)
}

// httptransport.StatusCoder
func (e HttpError) StatusCode() int {
	if original,ok := e.error.(StatusCoder); ok {
		return original.StatusCode()
	} else if e.SC == 0 {
		return http.StatusInternalServerError
	} else {
		return e.SC
	}
}

// httptransport.Headerer
func (e HttpError) Headers() http.Header {
	if original,ok := e.error.(Headerer); ok {
		return original.Headers()
	}
	return e.H
}

/**************************
	BadRequest Errors
***************************/
type BadRequestError struct {
	error
}

// httptransport.StatusCoder
func (_ BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}

type ValidationErrors struct {
	validator.ValidationErrors
}

// json.Marshaler
func (e ValidationErrors) MarshalJSON() ([]byte, error) {
	err := &HttpErrorResponse{
		StatusText: http.StatusText(e.StatusCode()),
		Message: "validation failed",
		Details: make(map[string]string, len(e.ValidationErrors)),
	}

	for _, obj := range e.ValidationErrors {
		fe := obj.(validator.FieldError)
		err.Details[fe.Namespace()] = fmt.Sprintf(templateValidationFieldError, fe.Field(), fe.Tag())
	}

	return json.Marshal(err)
}

// httptransport.StatusCoder
func (_ ValidationErrors) StatusCode() int {
	return http.StatusBadRequest
}

/*****************************
	Constructor Functions
******************************/
func NewHttpError(sc int, err error, headers ...http.Header) error {
	var h http.Header
	if len(headers) != 0 {
		h = make(http.Header)
		for _,toMerge := range headers {
			mergeHeaders(h, toMerge)
		}
	}
	return HttpError{error: err, SC: sc, H: h}
}

func NewBadRequestError(err error) error {
	return BadRequestError{error: err}
}

func mergeHeaders(src http.Header, toMerge http.Header) {
	for k, values := range toMerge {
		for _, v := range values {
			src.Add(k, v)
		}
	}
}

/*****************************
	Privates
******************************/