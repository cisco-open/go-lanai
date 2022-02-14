package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"io"
	"mime"
	"net/http"
	"reflect"
)

type Response struct {
	StatusCode int
	Headers http.Header
	Body interface{}
	RawBody []byte `json:"-"`
}

type ResponseOptions func(opt *responseOption)

type responseOption struct {
	body       interface{}
	errBody    ErrorResponseBody
	decodeFunc httptransport.DecodeResponseFunc
}

func fallbackResponseOptions(opt *responseOption) {
	if opt.decodeFunc == nil {
		if opt.body == nil {
			opt.body = &map[string]interface{}{}
		}
		if opt.errBody == nil {
			opt.errBody = &defaultErrorBody{}
		}
		opt.decodeFunc = makeJsonDecodeResponseFunc(opt)
	}
}

// JsonBody returns a ResponseOptions that specify interface{} to use for parsing response body as JSON
func JsonBody(body interface{}) ResponseOptions {
	return func(opt *responseOption) {
		opt.body = body
	}
}

// JsonErrorBody returns a ResponseOptions that specify interface{} to use for parsing error response as JSON
func JsonErrorBody(errBody ErrorResponseBody) ResponseOptions {
	return func(opt *responseOption) {
		opt.errBody = errBody
	}
}

// CustomResponseDecoder returns a ResponseOptions that specify custom decoding function of http.Response
// this options overwrite JsonBody and JsonErrorBody
func CustomResponseDecoder(dec httptransport.DecodeResponseFunc) ResponseOptions {
	return func(opt *responseOption) {
		//opt.decodeFunc = makeJsonDecodeResponseFunc(body)
		opt.decodeFunc = dec
	}
}

func makeJsonDecodeResponseFunc(opt *responseOption) httptransport.DecodeResponseFunc {
	if opt.decodeFunc != nil {
		return opt.decodeFunc
	}

	// standard decode func
	return func(ctx context.Context, resp *http.Response) (response interface{}, err error) {
		if resp.StatusCode > 299 {
			return nil, handleStatusCodeError(resp, opt.errBody)
		}

		// decode
		body := opt.body
		raw, e := decodeJsonBody(resp, body)
		if e != nil {
			return nil, e
		}

		// dereference if needed
		rv := reflect.ValueOf(body)
		if rv.Kind() == reflect.Ptr {
			ev := rv.Elem()
			switch ev.Kind() {
			case reflect.Map, reflect.Slice, reflect.Interface:
				body = ev.Interface()
			default:
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       body,
			RawBody:    raw,
		}, nil
	}
}

func handleStatusCodeError(resp *http.Response, errBody interface{}) error {
	raw, e := decodeJsonBody(resp, errBody)
	if e != nil {
		//nolint:errorlint
		if httpE, ok := e.(*Error); ok {
			return httpE.WithMessage("unable to parse error response: %v", e)
		} else {
			return e
		}
	}
	return NewErrorWithStatusCode(errBody, resp, raw)
}

// decodeJsonBody read body from http.Response and decode into given "body"
// function panic if "body" is nil
func decodeJsonBody(resp *http.Response, body interface{}) ([]byte, error) {
	defer resp.Body.Close()

	// check media type
	if e := validateMediaType(MediaTypeJson, resp); e != nil {
		return nil, e
	}

	// decode, and keep the raw bytes
	var data []byte
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return nil, NewSerializationError(fmt.Errorf("response IO error: %s", e), resp, data)
	}

	if e := json.Unmarshal(data, body); e != nil {
		return data, NewSerializationError(fmt.Errorf("response unmarshal error: %s", e), resp, data)
	}
	return data, nil
}

func validateMediaType(expected string, resp *http.Response) *Error {
	contentType := resp.Header.Get(HeaderContentType)
	mediaType, _, e := mime.ParseMediaType(contentType)
	if e != nil {
		return NewMediaTypeError(fmt.Errorf("received invalid content type %s", contentType), resp, nil, e)
	}

	if mediaType != MediaTypeJson {
		return NewMediaTypeError(fmt.Errorf("unsupported media type: %s, expected %s", mediaType, expected), resp, nil)
	}
	return nil
}

/*************************
	Error Unmarshal
 *************************/

type jsonErrorBody struct {
	Error   string            `json:"error,omitempty"`
	Message string            `json:"message,omitempty"`
	Desc    string            `json:"error_description,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// defaultErrorBody implements ErrorResponseBody, json.Marshaler, json.Unmarshaler
type defaultErrorBody struct {
	jsonErrorBody
}

func (b defaultErrorBody) Error() string {
	return b.jsonErrorBody.Error
}

func (b defaultErrorBody) Message() string {
	if b.jsonErrorBody.Message == "" {
		return b.jsonErrorBody.Desc
	}
	return b.jsonErrorBody.Message
}

func (b defaultErrorBody) Details() map[string]string {
	return b.jsonErrorBody.Details
}

// MarshalJSON implements json.Marshaler
func (b defaultErrorBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.jsonErrorBody)
}

// UnmarshalJSON implements json.Unmarshaler
func (b *defaultErrorBody) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.jsonErrorBody)
}

