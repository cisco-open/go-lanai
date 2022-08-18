package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// This can be the function the user of this repo needs
func NewRepo[T any](model *T, client OpenClient) Repo[T] {
	return RepoImpl[T]{client}
}

type Repo[T any] interface {
	// Search will search the cluster for data.
	//
	// The data will be unmarshalled and returned to the dest argument.
	// The body argument should follow the Search request body [Format].
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/search/#request-body
	Search(ctx context.Context, dest *[]T, body interface{}, o ...Option[opensearchapi.SearchRequest]) error

	// IndicesCreate will create a new index in the cluster.
	//
	// The index argument defines the index name to be created.
	// The mapping argument should follow the Index Create request body [Format].
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/index-apis/create-index/#request-body
	IndicesCreate(ctx context.Context, index string, mapping interface{}, o ...Option[opensearchapi.IndicesCreateRequest]) error
}

type RepoImpl[T any] struct {
	client OpenClient
}
