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
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/gin-gonic/gin"
    "net/http"
)

type RequestPreProcessorName string

type RequestPreProcessor interface {
	Process(r *http.Request) error
	Name() RequestPreProcessorName
}

type Engine struct {
	*gin.Engine
	requestPreProcessor []RequestPreProcessor
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, p := range e.requestPreProcessor {
		err := p.Process(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, "Internal error with request cache")
			return
		}
	}
	e.Engine.ServeHTTP(w, r)
}

func (e *Engine) addRequestPreProcessor(p RequestPreProcessor) {
	e.requestPreProcessor = append(e.requestPreProcessor, p)
}

func NewEngine() *Engine {
	if bootstrap.DebugEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	eng := &Engine{
		Engine: gin.New(),
	}
	return eng
}