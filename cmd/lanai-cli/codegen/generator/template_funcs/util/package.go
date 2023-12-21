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

package util

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"strings"
	"text/template"
)

var logger = log.New("Internal")
var FuncMap = template.FuncMap{
	"toTitle":       ToTitle,
	"toLower":       ToLower,
	"concat":        Concat,
	"basePath":      BasePath,
	"hasPrefix":     strings.HasPrefix,
	"replaceDashes": ReplaceDash,
	"args":          args,
	"increment":     increment,
	"log":           templateLog,
	"listContains":  ListContains,
	"derefBoolPtr":  derefBoolPtr,
}
