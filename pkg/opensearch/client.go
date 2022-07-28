package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type Request interface {
	opensearchapi.SearchRequest |
	opensearchapi.IndicesCreateRequest
}

type OpenClient interface {
	Search(ctx context.Context, o ...Option[opensearchapi.SearchRequest]) (*opensearchapi.Response, error)
	IndicesCreate(ctx context.Context, index string, o ...Option[opensearchapi.IndicesCreateRequest]) (*opensearchapi.Response, error)
}

type Option[T Request] func(request *T)

func NewClient(config Properties) (OpenClient, error) {
	client, err := opensearch.NewClient(config.GetConfig())
	if err != nil {
		logger.Errorf("unable to create new client: %v", err)
		return nil, err
	}
	return &OpenClientImpl{client}, nil
}

type OpenClientImpl struct {
	client *opensearch.Client
}
