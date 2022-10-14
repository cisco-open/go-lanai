package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesDeleteIndexTemplate(ctx context.Context, name string, o ...Option[opensearchapi.IndicesDeleteIndexTemplateRequest]) error {
	resp, err := c.client.IndicesDeleteIndexTemplate(ctx, name, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesDeleteIndexTemplate(ctx context.Context, name string, o ...Option[opensearchapi.IndicesDeleteIndexTemplateRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesDeleteIndexTemplateRequest), len(o))

	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesDeleteIndexTemplate.WithContext(ctx))
	resp, err := c.client.API.Indices.DeleteIndexTemplate(name, options...)

	return resp, err
}

type indicesDeleteIndexTemplate struct {
	opensearchapi.IndicesDeleteIndexTemplate
}

var IndicesDeleteIndexTemplate = indicesDeleteIndexTemplate{}
