package opensearch

import (
	"context"
	"encoding/json"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"strings"
)

func (c *RepoImpl[T]) BulkIndexer(ctx context.Context, action string, bulkItems *[]T, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexerStats, error) {
	arrBytes := make([][]byte, len(*bulkItems))
	for i, item := range *bulkItems {
		buffer, err := json.Marshal(item)
		if err != nil {
			return opensearchutil.BulkIndexerStats{}, err
		}
		arrBytes[i] = buffer
	}

	bi, err := c.client.BulkIndexer(ctx, action, arrBytes, o...)
	if err != nil {
		return opensearchutil.BulkIndexerStats{}, err
	}

	return bi.Stats(), nil
}

func (c *OpenClientImpl) BulkIndexer(ctx context.Context, action string, documents [][]byte, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexer, error) {
	options := make([]func(config *opensearchutil.BulkIndexerConfig), len(o))
	for i, v := range o {
		options[i] = v
	}

	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdBulk, Options: &options})
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
			Body:   strings.NewReader(string(item)),
		})
	}

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdBulk, Options: &options, Err: &err})
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

func (e bulkCfgExt) WithIndex(i string) func(*opensearchutil.BulkIndexerConfig) {
	return func(cfg *opensearchutil.BulkIndexerConfig) {
		cfg.Index = i
	}
}

type bulkCfgExt struct {
	opensearchutil.BulkIndexerConfig
}

var BulkIndexer = bulkCfgExt{}
