package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesPutAlias(ctx context.Context, index string, name string, o ...Option[opensearchapi.IndicesPutAliasRequest]) error {
	resp, err := c.client.IndicesPutAlias(ctx, index, name, o...)
	if resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *OpenClientImpl) IndicesPutAlias(ctx context.Context, index string, name string, o ...Option[opensearchapi.IndicesPutAliasRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesPutAliasRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesPutAlias.WithContext(ctx))
	resp, err := c.client.API.Indices.PutAlias([]string{index}, name, options...)

	return resp, err
}

type indicesPutAlias struct {
	opensearchapi.IndicesPutAlias
}

var IndicesPutAlias = indicesPutAlias{}
