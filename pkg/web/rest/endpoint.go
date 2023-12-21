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

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

/**********************************
	Json RequestDetails Encoder
***********************************/
// TODO Request should contains FormData bindings, URI bindings, etc.
//		Need to review if request encoder is still needed in "web" package. We have "httpclient" now
func jsonEncodeRequestFunc(_ context.Context, r *http.Request, body interface{}) error {
	// review this part
	r.Header.Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if e := json.NewEncoder(&buf).Encode(body); e != nil {
		return e
	}
	r.Body = io.NopCloser(&buf)
	return nil
}



