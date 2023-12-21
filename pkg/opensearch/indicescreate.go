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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesCreate(
	ctx context.Context,
	index string,
	mapping interface{},
	o ...Option[opensearchapi.IndicesCreateRequest],
) error {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(mapping)
	if err != nil {
		return fmt.Errorf("unable to encode mapping: %w", err)
	}
	o = append(o, IndicesCreate.WithBody(&buffer))
	resp, err := c.client.IndicesCreate(ctx, index, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesCreate(
	ctx context.Context,
	index string,
	o ...Option[opensearchapi.IndicesCreateRequest],
) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesCreateRequest), len(o))
	for i, v := range o {
		options[i] = v
	}
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdIndicesCreate, Options: &options})
	}

	//nolint:makezero
	options = append(options, IndicesCreate.WithContext(ctx))
	resp, err := c.client.Indices.Create(index, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndicesCreate, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

type indicesCreateExt struct {
	opensearchapi.IndicesCreate
}

var IndicesCreate = indicesCreateExt{}
