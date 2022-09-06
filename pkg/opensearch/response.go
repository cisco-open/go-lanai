package opensearch

import (
	"bytes"
	"encoding/json"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io"
)

// UnmarshalResponse will take the response, read the body out of it and then
// place the bytes that were read back into the body so it can be used again after
// this call
func UnmarshalResponse[T any](resp *opensearchapi.Response) (*T, error) {
	var model T
	var respBody []byte
	var err error
	if resp.Body != nil {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return &model, err
		}
	}
	// restore the resp.Body back to original state
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	err = json.Unmarshal(respBody, &model)
	if err != nil {
		return &model, err
	}
	return &model, nil
}
