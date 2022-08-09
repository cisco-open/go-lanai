package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io"
)

func (c *RepoImpl[T]) Index(ctx context.Context, index string, document T, o ...Option[opensearchapi.IndexRequest]) error {

	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(document)
	if err != nil {
		return err
	}
	resp, err := c.client.Index(ctx, index, &buffer)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *OpenClientImpl) Index(ctx context.Context, index string, body io.Reader, o ...Option[opensearchapi.IndexRequest]) (*opensearchapi.Response, error) {
	before, after := c.GetHooks()
	defer after.Run(HookContext{ctx, CmdIndex})
	before.Run(HookContext{ctx, CmdIndex})
	options := make([]func(request *opensearchapi.IndexRequest), len(o))
	for i, v := range o {
		options[i] = v
	}
	//nolint:makezero
	options = append(options, Index.WithContext(ctx))
	return c.client.API.Index(index, body, options...)
}

// indexExt can be extended
//	func (s indexExt) WithSomething() func(request *opensearchapi.Index) {
//		return func(request *opensearchapi.Index) {
//		}
//	}
type indexExt struct {
	opensearchapi.Index
}

var Index = indexExt{}
