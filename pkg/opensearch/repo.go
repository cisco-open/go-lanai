// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
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
	Search(ctx context.Context, dest *[]T, body interface{}, o ...Option[opensearchapi.SearchRequest]) (int, error)

	// SearchTemplate allows to use the Mustache language to pre-render a search definition
	//
	// The data will be unmarshalled and returned to the dest argument.
	// The body argument should follow the Search request body [Format].
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/search/#request-body
	SearchTemplate(ctx context.Context, dest *[]T, body interface{}, o ...Option[opensearchapi.SearchTemplateRequest]) (int, error)

	// Index will create a new Document in the index that is defined.
	//
	// The index argument defines the index name that the document should be stored in.
	Index(ctx context.Context, index string, document T, o ...Option[opensearchapi.IndexRequest]) error

	// BulkIndexer will process bulk requests of a single action type.
	//
	// The index argument defines the index name that the bulk action will target.
	// The action argument must be one of: ("index", "create", "delete", "update").
	// The bulkItems argument is the array of struct items to be actioned.
	//
	// [Ref]: https://pkg.go.dev/github.com/opensearch-project/opensearch-go/opensearchutil#BulkIndexerItem
	BulkIndexer(ctx context.Context, action BulkAction, bulkItems *[]T, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexerStats, error)

	// IndicesCreate will create a new index in the cluster.
	//
	// The index argument defines the index name to be created.
	// The mapping argument should follow the Index Create request body [Format].
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/index-apis/create-index/#request-body
	IndicesCreate(ctx context.Context, index string, mapping interface{}, o ...Option[opensearchapi.IndicesCreateRequest]) error

	// IndicesGet will return information about an index
	//
	// The index argument defines the index name we want to get
	IndicesGet(ctx context.Context, index string, o ...Option[opensearchapi.IndicesGetRequest]) (*IndicesDetail, error)

	// IndicesDelete will delete an index from the cluster.
	//
	// The index argument defines the index name to be deleted.
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/index-apis/delete-index/
	IndicesDelete(ctx context.Context, index []string, o ...Option[opensearchapi.IndicesDeleteRequest]) error

	// IndicesPutAlias will create or update an alias
	//
	// The index argument defines the index that the alias should point to
	// The name argument defines the name of the new alias
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/alias/#request-body
	IndicesPutAlias(ctx context.Context, index []string, name string, o ...Option[opensearchapi.IndicesPutAliasRequest]) error

	// IndicesDeleteAlias deletes an alias
	//
	// The index argument defines the index that the alias points to
	// The name argument defines the name of the alias we would like to delete
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/rest-api/alias/#request-body
	IndicesDeleteAlias(ctx context.Context, index []string, name []string, o ...Option[opensearchapi.IndicesDeleteAliasRequest]) error

	// IndicesPutIndexTemplate will create or update an alias
	//
	// The name argument defines the name of the template
	// The body argument defines the specified template options to apply (refer to [Format])
	//
	// [Format]: https://opensearch.org/docs/latest/opensearch/index-templates/#index-template-options
	IndicesPutIndexTemplate(ctx context.Context, name string, body interface{}, o ...Option[opensearchapi.IndicesPutIndexTemplateRequest]) error

	// IndicesDeleteIndexTemplate deletes an index template
	//
	// The name argument defines the name of the template to delete
	IndicesDeleteIndexTemplate(ctx context.Context, name string, o ...Option[opensearchapi.IndicesDeleteIndexTemplateRequest]) error

	// Ping will ping the OpenSearch cluster. If no error is returned, then the ping was successful
	Ping(ctx context.Context, o ...Option[opensearchapi.PingRequest]) error

	AddBeforeHook(hook BeforeHook)
	AddAfterHook(hook AfterHook)
	RemoveBeforeHook(hook BeforeHook)
	RemoveAfterHook(hook AfterHook)
}

type RepoImpl[T any] struct {
	client OpenClient
}

func (c *RepoImpl[T]) AddBeforeHook(hook BeforeHook) {
	c.client.AddBeforeHook(hook)
}

func (c *RepoImpl[T]) AddAfterHook(hook AfterHook) {
	c.client.AddAfterHook(hook)
}

func (c *RepoImpl[T]) RemoveBeforeHook(hook BeforeHook) {
	c.client.RemoveBeforeHook(hook)
}

func (c *RepoImpl[T]) RemoveAfterHook(hook AfterHook) {
	c.client.RemoveAfterHook(hook)
}
