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

package assets

import (
	"github.com/cisco-open/go-lanai/pkg/web"
	"net/http"
)

type assetsMapping struct {
	path    string
	root    string
	aliases map[string]string
}

func New(relativePath string, assetsRootPath string) web.StaticMapping {
	return &assetsMapping{
		path: relativePath,
		root: assetsRootPath,
		aliases: map[string]string{},
	}
}

/*****************************
	StaticMapping Interface
******************************/

func (m *assetsMapping) Name() string {
	return m.path
}

func (m *assetsMapping) Path() string {
	return m.path
}

func (m *assetsMapping) Method() string {
	return http.MethodGet
}

func (m *assetsMapping) StaticRoot() string {
	return m.root
}

func (m *assetsMapping) Aliases() map[string]string {
	return m.aliases
}

func (m *assetsMapping) AddAlias(path, filePath string) web.StaticMapping {
	m.aliases[path] = filePath
	return m
}