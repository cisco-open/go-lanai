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

package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ginRequestRewriter struct {
	engine *gin.Engine
}

func newGinRequestRewriter(engine *gin.Engine) RequestRewriter {
	return &ginRequestRewriter{
		engine: engine,
	}
}

// HandleRewrite Caution, you could loop yourself to death
func (rw ginRequestRewriter) HandleRewrite(r *http.Request) error {
	gc := GinContext(r.Context())
	if gc == nil {
		return fmt.Errorf("the request is not linked to a gin Context. Please make sure this is the right RequestRewriter to use")
	}

	rw.engine.HandleContext(gc)
	return nil
}
