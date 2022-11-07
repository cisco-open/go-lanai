package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesDelete(ctx context.Context, index []string, o ...Option[opensearchapi.IndicesDeleteRequest]) error {
	resp, err := c.client.IndicesDelete(ctx, index, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesDelete(ctx context.Context, index []string, o ...Option[opensearchapi.IndicesDeleteRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesDeleteRequest), len(o))
	for i, v := range o {
		options[i] = v
	}
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdIndicesDelete, Options: &options})
	}

	//nolint:makezero
	options = append(options, IndicesDelete.WithContext(ctx))
	resp, err := c.client.API.Indices.Delete(index, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndicesDelete, Options: &options, Resp: resp, Err: &err})
	}
	return resp, err
}

type indicesDeleteExt struct {
	opensearchapi.IndicesDelete
}

var IndicesDelete = indicesDeleteExt{}
