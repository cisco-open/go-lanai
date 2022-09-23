package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io/ioutil"
)

func (c *RepoImpl[T]) IndicesGet(ctx context.Context, dest *[]byte, index string, o ...Option[opensearchapi.IndicesGetRequest]) error {

	resp, err := c.client.IndicesGet(ctx, index, o...)
	if resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}

	*dest = respBody
	return nil

}

func (c *OpenClientImpl) IndicesGet(ctx context.Context, index string, o ...Option[opensearchapi.IndicesGetRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesGetRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesGet.WithContext(ctx))
	resp, err := c.client.Indices.Get([]string{index}, options...)

	return resp, err
}

type indicesGetExt struct {
	opensearchapi.IndicesGet
}

var IndicesGet = indicesGetExt{}
