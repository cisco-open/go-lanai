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

package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
)

type Parameters struct {
	openapi.Parameters
}

func (r Parameters) ExternalImports() []string {
	if r.CountFields() == 0 {
		return nil
	}
	var imports []string
	for _, parameter := range r.Parameters {
		if parameter.Ref == "" && parameter.Value.In != "path" && parameter.Value.In != "query" {
			for _, schema := range SchemaRef(*parameter.Value.Schema).AllSchemas() {
				imports = append(imports, lanaiutil.ExternalImportsFromFormat(schema)...)
			}
		}
	}
	return imports
}
