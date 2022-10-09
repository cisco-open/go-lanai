package opensearch

import (
	"github.com/opensearch-project/opensearch-go/opensearchutil"
)

func (c *RepoImpl[T]) NewBulkIndexer() (opensearchutil.BulkIndexer, error) {
	bi, err := c.client.NewBulkIndexer()
	if err != nil {
		return nil, err
	}
	return bi, nil
}

func (c *OpenClientImpl) NewBulkIndexer() (opensearchutil.BulkIndexer, error) {
	bi, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Client: c.client,
		// The number of worker goroutines - the default is based on the number of CPUs - for testing,
		// it is best that there is 1 goroutine to ensure the same request(s) in http/vcr
		NumWorkers: 1,
		Refresh:    "true",
	})

	if err != nil {
		return nil, err
	}
	return bi, nil
}
