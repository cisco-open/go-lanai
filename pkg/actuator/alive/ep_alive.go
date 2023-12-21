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

package alive

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"net/http"
)

const (
	ID                   = "alive"
	EnableByDefault      = true
)

type Input struct{}

type Output struct{
	sc int
	Message string `json:"msg"`
}

// http.StatusCoder
func (o Output) StatusCode() int {
	return o.sc
}

// AliveEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type AliveEndpoint struct {
	actuator.WebEndpointBase
}

func new(di regDI) *AliveEndpoint {
	ep := AliveEndpoint{}
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &di.MgtProperties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})
	return &ep
}

// Read never returns error
func (ep *AliveEndpoint) Read(ctx context.Context, input *Input) (Output, error) {
	return Output{
		sc: http.StatusOK,
		Message: "I'm good",
	}, nil
}
