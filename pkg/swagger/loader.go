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

package swagger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

type OASDocLoader struct {
	path string
	searchFS []fs.FS
}

func newOASDocLoader(path string, searchFS ...fs.FS) *OASDocLoader {
	if len(searchFS) == 0 {
		searchFS = []fs.FS{os.DirFS(".")}
	}
	return &OASDocLoader{
		path:     path,
		searchFS: searchFS,
	}
}

func (l OASDocLoader) Load() (*OpenApiSpec, error) {
	// find docs file
	var file fs.File
	var e error
	for _, fsys := range l.searchFS {
		switch file, e = fsys.Open(l.path); {
		case errors.Is(e, fs.ErrNotExist):
			continue
		case e != nil:
			return nil, e
		}
		break
	}
	if file == nil {
		return nil, fs.ErrNotExist
	}
	defer func() { _ = file.Close() }()

	// load docs
	var oas OpenApiSpec
	switch fileExt := strings.ToLower(path.Ext(l.path)); fileExt {
	case ".yml", ".yaml":
		data, e := io.ReadAll(file)
		if e != nil {
			return nil, e
		}
		if e := yaml.Unmarshal(data, &oas); e != nil {
			return nil, e
		}
	case ".json", ".json5":
		decoder := json.NewDecoder(file)
		if e := decoder.Decode(&oas); e != nil {
			return nil, e
		}
	default:
		return nil, fmt.Errorf("unsupported file extension for OAS document: %s", fileExt)
	}

	return &oas, nil
}
