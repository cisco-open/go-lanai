package rest

import (
	"bytes"
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
	"io/ioutil"
	"net/http"
)

/**********************************
	Json Request Encoder
***********************************/
// FIXME this is not a correct implementation, because request should contains FormData bindings, URI bindings, etc
func jsonEncodeRequestFunc(_ context.Context, r *http.Request, request interface{}) error {
	// review this part
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

/**********************************
	JSON Response Encoder
***********************************/
func jsonEncodeResponseFunc(_ context.Context, w http.ResponseWriter, response interface{}) error {
	// overwrite headers
	if headerer, ok := response.(httptransport.Headerer); ok {
		w = web.NewLazyHeaderWriter(w)
		overwriteHeaders(w, headerer)
	}

	// additional headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// write header and status code
	if coder, ok := response.(httptransport.StatusCoder); ok {
		w.WriteHeader(coder.StatusCode())
	}

	if entity, ok := response.(web.BodyContainer); ok {
		response = entity.Body()
		// For Debug
		//return web.NewHttpError(405, errors.New(fmt.Sprintf("Cannot serialize response %T", entity.Body())) )
	}
	return json.NewEncoder(w).Encode(response)
}

func overwriteHeaders(w http.ResponseWriter, h httptransport.Headerer) {
	for key, values := range h.Headers() {
		for i, val := range values {
			if i == 0 {
				w.Header().Set(key, val)
			} else {
				w.Header().Add(key, val)
			}
		}
	}
}

/*****************************
	JSON Error Encoder
******************************/
func jsonErrorEncoder(c context.Context, err error, w http.ResponseWriter) {
	if _,ok := err.(json.Marshaler); !ok {
		err = web.ToHttpError(err)
	}
	httptransport.DefaultErrorEncoder(c, err, w)
}


