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

package testdata

import (
    "context"
    "dario.cat/mergo"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web/template"
    "net/http"
    "reflect"
    "strings"
)

func IndexPage(_ context.Context, _ *http.Request) (template.ModelView, error) {
	return template.ModelView{
		View: "index.html.tmpl",
		Model: template.Model{
			"Title": "TemplateMVCTest",
		},
	}, nil
}

func RedirectPage(_ context.Context, _ *http.Request) (*template.ModelView, error) {
	return template.RedirectView("/index", http.StatusFound, false), nil
}

const ModelPrintTmpl = `%s=%v`

// PrintKV is a template function
func PrintKV(model map[string]any) string {
	lines := flattenMap(model, "")
	return strings.Join(lines, "\n")
}

func flattenMap[T any](m map[string]T, prefix string) []string {
	lines := make([]string, 0, len(m))
	for k, val := range m {
		var unknown interface{} = val
		switch v := unknown.(type) {
		case template.RequestContext:
			lines = append(lines, flattenMap(v, prefix+"."+k)...)
		case map[string]any:
			lines = append(lines, flattenMap(v, prefix+"."+k)...)
		case map[string]string:
			lines = append(lines, flattenMap(v, prefix+"."+k)...)
		case []any:
			for i := range v {
				k = fmt.Sprintf(`%s.%d`, prefix, i)
				lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v[i]))
			}
		case []string:
			for i := range v {
				k = fmt.Sprintf(`%s.%d`, prefix, i)
				lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v[i]))
			}
		case fmt.Stringer, fmt.GoStringer, error:
			k = fmt.Sprintf(`%s.%s`, prefix, k)
			lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v))
		case nil:
			// do nothing
		default:
			var converted map[string]interface{}
			switch reflect.Indirect(reflect.ValueOf(v)).Kind() {
			case reflect.Struct:
				converted = map[string]interface{}{}
				_ = mergo.Map(&converted, v)
			default:
			}
			if len(converted) != 0 {
				lines = append(lines, flattenMap(converted, prefix+"."+k)...)
			} else {
				k = fmt.Sprintf(`%s.%s`, prefix, k)
				lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v))
			}
		}
	}
	return lines
}

