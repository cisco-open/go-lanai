package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesDeleteAlias(ctx context.Context, index string, name string, o ...Option[opensearchapi.IndicesDeleteAliasRequest]) error {
	resp, err := c.client.IndicesDeleteAlias(ctx, index, name, o...)
	if resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *OpenClientImpl) IndicesDeleteAlias(ctx context.Context, index string, name string, o ...Option[opensearchapi.IndicesDeleteAliasRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesDeleteAliasRequest), len(o))

	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesDeleteAlias.WithContext(ctx))
	resp, err := c.client.API.Indices.DeleteAlias([]string{index}, []string{name}, options...)

	return resp, err
}

type indicesDeleteAlias struct {
	opensearchapi.IndicesDeleteAlias
}

var IndicesDeleteAlias = indicesDeleteAlias{}
