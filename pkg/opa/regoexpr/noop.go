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

package regoexpr

import (
	"context"
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"strings"
)

type NoopPartialQueryMapper struct{}

func (m NoopPartialQueryMapper) MapResults(pq *rego.PartialQueries) (interface{}, error) {
	return TranslatePartialQueries(context.Background(), pq, func(opts *TranslateOption[string]) {
		opts.Translator = NoopQueryTranslator{}
	})
}

func (m NoopPartialQueryMapper) ResultToJSON(result interface{}) (interface{}, error) {
	return result, nil
}

type NoopQueryTranslator struct{}

func (t NoopQueryTranslator) Negate(_ context.Context, expr string) string {
	return fmt.Sprintf(`!%s`, expr)
}

func (t NoopQueryTranslator) And(_ context.Context, exprs ...string) string {
	return strings.Join(exprs, " && ")
}

func (t NoopQueryTranslator) Or(_ context.Context, exprs ...string) string {
	return strings.Join(exprs, " || ")
}

func (t NoopQueryTranslator) Comparison(_ context.Context, op ast.Ref, colRef ast.Ref, val interface{}) (string, error) {
	return fmt.Sprintf(`%v %v %v`, colRef, op, val), nil
}
