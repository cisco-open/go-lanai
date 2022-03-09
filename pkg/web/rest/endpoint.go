package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

/**********************************
	Json RequestDetails Encoder
***********************************/
// TODO Request should contains FormData bindings, URI bindings, etc.
//		Need to review if request encoder is still needed in "web" package. We have "httpclient" now
func jsonEncodeRequestFunc(_ context.Context, r *http.Request, body interface{}) error {
	// review this part
	r.Header.Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if e := json.NewEncoder(&buf).Encode(body); e != nil {
		return e
	}
	r.Body = io.NopCloser(&buf)
	return nil
}



