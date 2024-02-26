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

package profiler

import (
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"net/http/pprof"
)

type PProfRequest struct {
	Profile string `uri:"profile"`
}

type PProfController struct{}

func (c *PProfController) Mappings() []web.Mapping {
	return []web.Mapping{
		web.NewSimpleGinMapping("pprof_gin", RouteGroup, PathPrefixPProf + "/:profile", web.MethodAny, nil, c.Profile),
		web.NewSimpleMapping("pprof_index", RouteGroup, PathPrefixPProf , web.MethodAny, nil, pprof.Index),
		web.NewSimpleMapping("pprof_cli", RouteGroup, PathPrefixPProf + "/cmdline", web.MethodAny, nil, pprof.Cmdline),
		web.NewSimpleMapping("pprof_profile", RouteGroup, PathPrefixPProf + "/profile", web.MethodAny, nil, pprof.Profile),
		web.NewSimpleMapping("pprof_symbol", RouteGroup, PathPrefixPProf + "/symbol", web.MethodAny, nil, pprof.Symbol),
		web.NewSimpleMapping("pprof_trace", RouteGroup, PathPrefixPProf + "/trace", web.MethodAny, nil, pprof.Trace),
	}

}

func (c *PProfController) Profile(gc *gin.Context) {
	var req PProfRequest
	if e := gc.BindUri(&req); e != nil {
		pprof.Index(gc.Writer, gc.Request)
		return
	}

	handler := pprof.Handler(req.Profile)
	if handler == nil {
		pprof.Index(gc.Writer, gc.Request)
		return
	}

	handler.ServeHTTP(gc.Writer, gc.Request)
}
