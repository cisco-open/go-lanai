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

func (c *RepoImpl[T]) IndicesPutIndexTemplate(ctx context.Context, name string, body interface{}, o ...Option[opensearchapi.IndicesPutIndexTemplateRequest]) error {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(body)
	if err != nil {
		return fmt.Errorf("unable to encode mapping: %w", err)
	}
	resp, err := c.client.IndicesPutIndexTemplate(ctx, name, &buffer, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Debugf("error response: %s", resp.String())
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) IndicesPutIndexTemplate(ctx context.Context, name string, body io.Reader, o ...Option[opensearchapi.IndicesPutIndexTemplateRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesPutIndexTemplateRequest), len(o))

	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesPutIndexTemplate.WithContext(ctx))
	resp, err := c.client.API.Indices.PutIndexTemplate(name, body, options...)

	return resp, err
}

type indicesPutIndexTemplate struct {
	opensearchapi.IndicesPutIndexTemplate
}

var IndicesPutIndexTemplate = indicesPutIndexTemplate{}
