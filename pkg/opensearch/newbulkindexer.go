package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
)

func (c *RepoImpl[T]) NewBulkIndexer(ctx context.Context, index string) (opensearchutil.BulkIndexer, error) {
	bi, err := c.client.NewBulkIndexer(ctx, index)
	if err != nil {
		return nil, err
	}
	return bi, nil
}

func (c *OpenClientImpl) NewBulkIndexer(ctx context.Context, index string) (opensearchutil.BulkIndexer, error) {

	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdNew, Options: &options})
	}

	bi, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Index:  index,
		Client: c.client,
	})

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndicesDelete, Options: &options, Resp: resp, Err: &err})
	}

	if err != nil {
		return nil, err
	}
	return bi, nil
}
