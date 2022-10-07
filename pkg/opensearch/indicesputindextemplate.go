package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io"
)

func (c *RepoImpl[T]) IndicesPutIndexTemplate(ctx context.Context, name string, body interface{}, o ...Option[opensearchapi.IndicesPutIndexTemplateRequest]) error {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(body)
	if err != nil {
		return fmt.Errorf("unable to encode mapping: %w", err)
	}
	resp, err := c.client.IndicesPutIndexTemplate(ctx, name, &buffer, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesPutIndexTemplate(ctx context.Context, name string, body io.Reader, o ...Option[opensearchapi.IndicesPutIndexTemplateRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesPutIndexTemplateRequest), len(o))

	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesPutIndexTemplate.WithContext(ctx))
	resp, err := c.client.API.Indices.PutIndexTemplate(name, body, options...)

	return resp, err
}

type indicesPutIndexTemplate struct {
	opensearchapi.IndicesPutIndexTemplate
}

var IndicesPutIndexTemplate = indicesPutIndexTemplate{}
