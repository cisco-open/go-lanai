package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesPutAlias(ctx context.Context,
	index []string,
	name string,
	o ...Option[opensearchapi.IndicesPutAliasRequest],
) error {
	resp, err := c.client.IndicesPutAlias(ctx, index, name, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesPutAlias(ctx context.Context, index []string, name string, o ...Option[opensearchapi.IndicesPutAliasRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesPutAliasRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdIndicesPutAlias, Options: &options})
	}

	//nolint:makezero
	options = append(options, IndicesPutAlias.WithContext(ctx))
	resp, err := c.client.API.Indices.PutAlias(index, name, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndicesPutAlias, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

type indicesPutAlias struct {
	opensearchapi.IndicesPutAlias
}

var IndicesPutAlias = indicesPutAlias{}
