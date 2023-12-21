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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io"
)

func (c *RepoImpl[T]) SearchTemplate(ctx context.Context, dest *[]T, body interface{}, o ...Option[opensearchapi.SearchTemplateRequest]) (hits int, err error) {
	var buffer bytes.Buffer
	err = json.NewEncoder(&buffer).Encode(body)
	if err != nil {
		return 0, fmt.Errorf("unable to encode mapping: %w", err)
	}
	resp, err := c.client.SearchTemplate(ctx, &buffer, o...)
	if err != nil {
		return 0, err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return 0, fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var searchResp SearchResponse[T]
	err = json.Unmarshal(respBody, &searchResp)
	if err != nil {
		return 0, err
	}
	retModel := make([]T, len(searchResp.Hits.Hits))
	for i, hits := range searchResp.Hits.Hits {
		retModel[i] = hits.Source
	}
	*dest = retModel
	return searchResp.Hits.Total.Value, nil
}

func (c *OpenClientImpl) SearchTemplate(ctx context.Context, body io.Reader, o ...Option[opensearchapi.SearchTemplateRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.SearchTemplateRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	order.SortStable(c.beforeHook, order.OrderedFirstCompare)
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdSearchTemplate, Options: &options})
	}

	//nolint:makezero
	options = append(options, SearchTemplate.WithContext(ctx))
	resp, err := c.client.API.SearchTemplate(body, options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdSearchTemplate, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

type searchTemplateExt struct {
	opensearchapi.SearchTemplate
}

var SearchTemplate = searchTemplateExt{}
