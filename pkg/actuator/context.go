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

package actuator

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/web"
)

const (
	OperationRead OperationMode = iota
	OperationWrite
)
type OperationMode int

// OperationFunc is a func that have following signature:
// 	func(ctx context.Context, input StructsOrPointerType1) (StructsOrPointerType2, error)
// where
//	- StructsOrPointerType1 and StructsOrPointerType2 can be any structs or struct pointers
//  - input might be ignored by particular Endpoint impl.
//  - 1st output is optional for "write" operations
//
// Note: golang doesn't have generics yet...
type OperationFunc interface{}

type Operation interface {
	Mode() OperationMode
	Func() OperationFunc
	Matches(ctx context.Context, mode OperationMode, input interface{}) bool
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

type Endpoint interface {
	Id() string
	EnabledByDefault() bool
	Operations() []Operation
}

type WebEndpoint interface {
	Mappings(op Operation, group string) ([]web.Mapping, error)
}

type EndpointExecutor interface {
	ReadOperation(ctx context.Context, input interface{}) (interface{}, error)
	WriteOperation(ctx context.Context, input interface{}) (interface{}, error)
}

type WebEndpoints map[string][]web.Mapping

func (w WebEndpoints) EndpointIDs() (ret []string) {
	ret = make([]string, 0, len(w))
	for k, _ := range w {
		ret = append(ret, k)
	}
	return
}

// Paths returns all path patterns of given endpoint ID.
// only web.RoutedMapping & web.StaticMapping is possible to extract this information
func (w WebEndpoints) Paths(id string) []string {
	mappings, ok := w[id]
	if !ok {
		return []string{}
	}

	paths := utils.NewStringSet()
	for _, v := range mappings {
		switch v.(type) {
		case web.RoutedMapping:
			paths.Add(v.(web.RoutedMapping).Path())
		case web.StaticMapping:
			paths.Add(v.(web.StaticMapping).Path())
		}
	}
	return paths.Values()
}
