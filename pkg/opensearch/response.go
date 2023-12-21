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
	"encoding/json"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io"
)

// UnmarshalResponse will take the response, read the body out of it and then
// place the bytes that were read back into the body so it can be used again after
// this call
func UnmarshalResponse[T any](resp *opensearchapi.Response) (*T, error) {
	var model T
	var respBody []byte
	var err error
	if resp.Body != nil {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return &model, err
		}
	}
	// restore the resp.Body back to original state
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	err = json.Unmarshal(respBody, &model)
	if err != nil {
		return &model, err
	}
	return &model, nil
}
