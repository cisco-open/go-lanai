package rest

import (
	"bytes"
	"context"
	"encoding/json"
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
	// TODO review this part
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}


