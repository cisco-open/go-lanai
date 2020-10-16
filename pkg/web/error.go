package web

import (
	"encoding/json"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
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

type httpError struct {
	error
	sc int
	header http.Header
}

// TODO consider implement Unwrap

func (e httpError) MarshalJSON() ([]byte, error) {
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
func (e httpError) StatusCode() int {
	if original,ok := e.error.(httptransport.StatusCoder); ok {
		return original.StatusCode()
	} else if e.sc == 0 {
		return http.StatusInternalServerError
	} else {
		return e.sc
	}
}

// httptransport.Headerer
func (e httpError) Headers() http.Header {
	if original,ok := e.error.(httptransport.Headerer); ok {
		return original.Headers()
	}
	return e.header
}

/**************************
	Other HTTP Errors
***************************/
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
	Helper Functions
******************************/
func HttpError(err error) error {
	switch err.(type) {
	case nil:
		return nil
	case httpError:
		return err
	case validator.ValidationErrors:
		return ValidationErrors{err.(validator.ValidationErrors)}
	}
	return httpError{error: err, sc: http.StatusInternalServerError}
}

/*****************************
	Privates
******************************/