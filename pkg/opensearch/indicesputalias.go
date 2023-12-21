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
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) IndicesPutAlias(ctx context.Context,
	index []string,
	name string,
	o ...Option[opensearchapi.IndicesPutAliasRequest],
) error {
	resp, err := c.client.IndicesPutAlias(ctx, index, name, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesPutAlias(ctx context.Context, index []string, name string, o ...Option[opensearchapi.IndicesPutAliasRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesPutAliasRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdIndicesPutAlias, Options: &options})
	}

	//nolint:makezero
	options = append(options, IndicesPutAlias.WithContext(ctx))
	resp, err := c.client.API.Indices.PutAlias(index, name, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdIndicesPutAlias, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

type indicesPutAlias struct {
	opensearchapi.IndicesPutAlias
}

var IndicesPutAlias = indicesPutAlias{}
