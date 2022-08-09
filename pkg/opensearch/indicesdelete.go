package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesDelete(ctx context.Context, index string, o ...Option[opensearchapi.IndicesDeleteRequest]) error {
	resp, err := c.client.IndicesDelete(ctx, index, o...)
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

func (c *OpenClientImpl) IndicesDelete(ctx context.Context, index string, o ...Option[opensearchapi.IndicesDeleteRequest]) (*opensearchapi.Response, error) {
	before, after := c.GetHooks()
	defer after.Run(HookContext{ctx, CmdIndicesDelete})
	before.Run(HookContext{ctx, CmdIndicesDelete})
	options := make([]func(request *opensearchapi.IndicesDeleteRequest), len(o))
	for i, v := range o {
		options[i] = v
	}
	//nolint:makezero
	options = append(options, IndicesDelete.WithContext(ctx))
	return c.client.API.Indices.Delete([]string{index}, options...)
}

type indicesDeleteExt struct {
	opensearchapi.IndicesDelete
}

var IndicesDelete = indicesDeleteExt{}
