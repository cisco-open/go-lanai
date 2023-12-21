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

func (c *RepoImpl[T]) IndicesDeleteIndexTemplate(ctx context.Context, name string, o ...Option[opensearchapi.IndicesDeleteIndexTemplateRequest]) error {
	resp, err := c.client.IndicesDeleteIndexTemplate(ctx, name, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesDeleteIndexTemplate(ctx context.Context, name string, o ...Option[opensearchapi.IndicesDeleteIndexTemplateRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesDeleteIndexTemplateRequest), len(o))

	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesDeleteIndexTemplate.WithContext(ctx))
	resp, err := c.client.API.Indices.DeleteIndexTemplate(name, options...)

	return resp, err
}

type indicesDeleteIndexTemplate struct {
	opensearchapi.IndicesDeleteIndexTemplate
}

var IndicesDeleteIndexTemplate = indicesDeleteIndexTemplate{}
