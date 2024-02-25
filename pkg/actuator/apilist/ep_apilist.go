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

package apilist

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/actuator"
    "github.com/cisco-open/go-lanai/pkg/web"
    "io/fs"
    "net/http"
)

const (
	ID                   = "apilist"
	EnableByDefault      = false
)

// ApiListEndpoint implements actuator.Endpoint, actuator.WebEndpoint
//goland:noinspection GoNameStartsWithPackageName
type ApiListEndpoint struct {
	actuator.WebEndpointBase
	staticPath string
}

func newEndpoint(di regDI) *ApiListEndpoint {
	if !fs.ValidPath(di.Properties.StaticPath) {
		panic("invalid static-path for apilist endpoint")
	}
	ep := ApiListEndpoint{
		staticPath: di.Properties.StaticPath,
	}
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
func (ep *ApiListEndpoint) Read(ctx context.Context, _ *struct{}) (interface{}, error) {
	resp, e := parseFromStaticFile(ep.staticPath)
	if e != nil {
		// Note we don't expose error. Instead, we return 404 like nothing is there
		logger.WithContext(ctx).Warnf(`unable to load static API list file "%s": %v`, ep.staticPath, e)
		return nil, web.NewHttpError(http.StatusNotFound, fmt.Errorf("APIList is not available"))
	}
	return resp, nil
}

func parseFromStaticFile(path string) (ret interface{}, err error) {
	// open
	var file fs.File
	var e error
	for _, fsys := range staticFS {
		if file, e = fsys.Open(path); e == nil {
			break
		}
	}
	if e != nil {
		return nil, e
	}

	// read
	defer func(){ _ = file.Close() }()
	decoder := json.NewDecoder(file)
	if e := decoder.Decode(&ret); e != nil {
		return nil, e
	}
	return
}

