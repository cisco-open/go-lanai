package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// NewRepo will return a OpenSearch repository for any model type T
func NewRepo[T any](model *T, client OpenClient) Repo[T] {
	return &RepoImpl[T]{
		client: client,
	}
}

type Repo[T any] interface {
	// Search will search the cluster for data.
	//
	// The data will be unmarshalled and returned to the dest argument.
	// The body argument should follow the Search request body [Format].
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/search/#request-body
	Search(ctx context.Context, dest *[]T, body interface{}, o ...Option[opensearchapi.SearchRequest]) error

	// Index will create a new Document in the index that is defined
	//
	// The index argument defines the index name that the document should be stored in.
	// The body
	Index(ctx context.Context, index string, document T, o ...Option[opensearchapi.IndexRequest]) error

	// IndicesCreate will create a new index in the cluster.
	//
	// The index argument defines the index name to be created.
	// The mapping argument should follow the Index Create request body [Format].
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/index-apis/create-index/#request-body
	IndicesCreate(ctx context.Context, index string, mapping interface{}, o ...Option[opensearchapi.IndicesCreateRequest]) error

	// IndicesDelete will delete an index from the cluster.
	//
	// The index argument defines the index name to be created.
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/index-apis/create-index/#request-body
	IndicesDelete(ctx context.Context, index string, o ...Option[opensearchapi.IndicesDeleteRequest]) error

	AddHook(hook HookContainer)
}

type RepoImpl[T any] struct {
	client OpenClient
}

func (c *RepoImpl[T]) AddHook(hook HookContainer) {
	c.client.AddHook(hook)
}
