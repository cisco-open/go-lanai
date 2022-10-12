package opensearch

import (
	"context"
	"encoding/json"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"strings"
)

func (c *RepoImpl[T]) BulkIndexer(ctx context.Context, index string, action string, bulkItems *[]T, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexerStats, error) {
	arrBytes := make([][]byte, len(*bulkItems))
	for i, item := range *bulkItems {
		buffer, err := json.Marshal(item)
		if err != nil {
			return opensearchutil.BulkIndexerStats{}, err
		}
		arrBytes[i] = buffer
	}

	bi, err := c.client.BulkIndexer(ctx, index, action, arrBytes, o...)
	if err != nil {
		return opensearchutil.BulkIndexerStats{}, err
	}

	return bi.Stats(), nil
}

func (c *OpenClientImpl) BulkIndexer(ctx context.Context, index string, action string, documents [][]byte, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexer, error) {
	options := make([]func(config *opensearchutil.BulkIndexerConfig), len(o))
	for i, v := range o {
		options[i] = v
	}

	options = append(options, WithClient(c.client))
	cfg := MakeConfig(options...)
	bi, err := opensearchutil.NewBulkIndexer(*cfg)
	if err != nil {
		return nil, err
	}

	for _, item := range documents {
		bi.Add(ctx, opensearchutil.BulkIndexerItem{
			Action: action,
			Index:  index,
			Body:   strings.NewReader(string(item)),
		})
	}

	if err = bi.Close(ctx); err != nil {
		return bi, err
	}

	return bi, nil
}

func MakeConfig(options ...func(*opensearchutil.BulkIndexerConfig)) *opensearchutil.BulkIndexerConfig {
	cfg := &opensearchutil.BulkIndexerConfig{}
	for _, o := range options {
		o(cfg)
	}
	return cfg
}

func WithClient(c *opensearch.Client) func(*opensearchutil.BulkIndexerConfig) {
	return func(cfg *opensearchutil.BulkIndexerConfig) {
		cfg.Client = c
	}
}

func (e bulkCfgExt) WithWorkers(n int) func(*opensearchutil.BulkIndexerConfig) {
	return func(cfg *opensearchutil.BulkIndexerConfig) {
		cfg.NumWorkers = n
	}
}

func (e bulkCfgExt) WithRefresh(s string) func(*opensearchutil.BulkIndexerConfig) {
	return func(cfg *opensearchutil.BulkIndexerConfig) {
		cfg.Refresh = s
	}
}

type bulkCfgExt struct {
	opensearchutil.BulkIndexerConfig
}

var BulkIndexer = bulkCfgExt{}
