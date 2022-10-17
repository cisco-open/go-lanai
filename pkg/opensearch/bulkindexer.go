package opensearch

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"encoding/json"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"strconv"
	"strings"
)

type bulkAction string

const (
	BulkActionIndex  bulkAction = "index"
	BulkActionCreate bulkAction = "create"
	BulkActionUpdate bulkAction = "update"
	BulkActionDelete bulkAction = "delete"
)

func (c *RepoImpl[T]) BulkIndexer(ctx context.Context, action bulkAction, bulkItems *[]T, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexerStats, error) {
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

func (c *OpenClientImpl) BulkIndexer(ctx context.Context, action bulkAction, documents [][]byte, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexer, error) {
	options := make([]func(config *opensearchutil.BulkIndexerConfig), len(o))
	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, WithClient(c.client))
	order.SortStable(c.beforeHook, order.OrderedFirstCompare)
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdBulk, Options: &options})
	}

	cfg := MakeConfig(options...)
	bi, err := opensearchutil.NewBulkIndexer(*cfg)
	if err != nil {
		return nil, err
	}

	for _, item := range documents {
		err = bi.Add(ctx, opensearchutil.BulkIndexerItem{
			Action: string(action),
			Body:   strings.NewReader(string(item)),
		})
		if err != nil {
			return nil, err
		}
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

func (e bulkCfgExt) WithRefresh(b bool) func(*opensearchutil.BulkIndexerConfig) {
	return func(cfg *opensearchutil.BulkIndexerConfig) {
		cfg.Refresh = strconv.FormatBool(b)
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
