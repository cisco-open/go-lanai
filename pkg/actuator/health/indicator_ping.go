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

package health

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

type PingIndicator struct {

}

func (b PingIndicator) Name() string {
	return "ping"
}

func (b PingIndicator) Health(ctx context.Context, options Options) Health {
	// very basic check: if the given context is *gin.Context, it means the health check is invoked via web endpoint.
	// therefore the web framework is still working
	if g := web.GinContext(ctx); g != nil {
		return NewDetailedHealth(StatusUp, "ping", nil)
	}
	return NewDetailedHealth(StatusUnknown, "ping", nil)
}

