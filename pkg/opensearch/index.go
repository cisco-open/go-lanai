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
	"io"
)

func (c *RepoImpl[T]) Index(ctx context.Context, index string, document T, o ...Option[opensearchapi.IndexRequest]) error {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(document)
	if err != nil {
		return err
	}
	resp, err := c.client.Index(ctx, index, &buffer, o...)
	if err != nil {
		return err
	}
	if resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) Index(ctx context.Context, index string, body io.Reader, o ...Option[opensearchapi.IndexRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndexRequest), len(o))
	for i, v := range o {
		options[i] = v
	}
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdIndex, Options: &options})
	}

	//nolint:makezero
	options = append(options, Index.WithContext(ctx))
	resp, err := c.client.API.Index(index, body, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndex, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

// indexExt can be extended
//	func (s indexExt) WithSomething() func(request *opensearchapi.Index) {
//		return func(request *opensearchapi.Index) {
//		}
//	}
type indexExt struct {
	opensearchapi.Index
}

var Index = indexExt{}
