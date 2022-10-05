package opensearch

import (
	opensearchutil "github.com/opensearch-project/opensearch-go/opensearchutil"
)

func (c *RepoImpl[T]) NewBulkIndexer(index string) (opensearchutil.BulkIndexer, error) {
	bi, err := c.client.NewBulkIndexer(index)
	if err != nil {
		return nil, err
	}
	return bi, nil
}

func (c *OpenClientImpl) NewBulkIndexer(index string) (opensearchutil.BulkIndexer, error) {
	bi, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Index:  index,
		Client: c.client,
	})
	if err != nil {
		return nil, err
	}
	return bi, nil
}
