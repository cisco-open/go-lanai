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
    "encoding/json"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
    "github.com/opensearch-project/opensearch-go"
    "github.com/opensearch-project/opensearch-go/opensearchutil"
    "strconv"
    "strings"
)

// BulkAction is intended to be used as an enum type for bulk actions
//
// [REF]: https://opensearch.org/docs/1.2/opensearch/rest-api/document-apis/bulk/#request-body
type BulkAction string

const (
	BulkActionIndex  BulkAction = "index"  // Will add in a document and will override any duplicate (based on ID)
	BulkActionCreate BulkAction = "create" // Will add a document if it doesn't exist or return an error
	BulkActionUpdate BulkAction = "update" // Will update an existing document if it exists or return an error
	BulkActionDelete BulkAction = "delete" // Will delete a document if it exists or return a `not_found`
)

func (c *RepoImpl[T]) BulkIndexer(ctx context.Context, action BulkAction, bulkItems *[]T, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexerStats, error) {
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

func (c *OpenClientImpl) BulkIndexer(ctx context.Context, action BulkAction, documents [][]byte, o ...Option[opensearchutil.BulkIndexerConfig]) (opensearchutil.BulkIndexer, error) {
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
