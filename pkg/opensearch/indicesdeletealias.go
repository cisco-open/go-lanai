package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesDeleteAlias(ctx context.Context, index []string, name []string, o ...Option[opensearchapi.IndicesDeleteAliasRequest]) error {
	resp, err := c.client.IndicesDeleteAlias(ctx, index, name, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesDeleteAlias(ctx context.Context, index []string, name []string, o ...Option[opensearchapi.IndicesDeleteAliasRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesDeleteAliasRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdIndicesDeleteAlias, Options: &options})
	}
	//nolint:makezero
	options = append(options, IndicesDeleteAlias.WithContext(ctx))
	resp, err := c.client.API.Indices.DeleteAlias(index, name, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndicesDeleteAlias, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

type indicesDeleteAlias struct {
	opensearchapi.IndicesDeleteAlias
}

var IndicesDeleteAlias = indicesDeleteAlias{}
