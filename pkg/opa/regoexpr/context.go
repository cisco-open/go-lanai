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
	"github.com/cisco-open/go-lanai/pkg/opa"
	"github.com/open-policy-agent/opa/ast"
)

var (
	ParsingError = opa.NewError(`generic OPA partial query parsing error`)
)

type QueryTranslator[EXPR any] interface {
	Negate(ctx context.Context, expr EXPR) EXPR
	And(ctx context.Context, expr ...EXPR) EXPR
	Or(ctx context.Context, expr ...EXPR) EXPR
	Comparison(ctx context.Context, op ast.Ref, colRef ast.Ref, val interface{}) (EXPR, error)
}

var (
	TermInternal = ast.VarTerm("internal")
	OpInternal   = ast.Ref([]*ast.Term{TermInternal})
	OpIn         = ast.Member.Ref()
	OpEqual      = ast.Equality.Ref()
	OpEq         = ast.Equal.Ref()
	OpNeq        = ast.NotEqual.Ref()
	OpLte        = ast.LessThanEq.Ref()
	OpLt         = ast.LessThan.Ref()
	OpGte        = ast.GreaterThanEq.Ref()
	OpGt         = ast.GreaterThan.Ref()
)

var (
	OpHashEqual = OpEqual.Hash()
	OpHashEq    = OpEq.Hash()
	OpHashNeq   = OpNeq.Hash()
	OpHashLte   = OpLte.Hash()
	OpHashLt    = OpLt.Hash()
	OpHashGte   = OpGte.Hash()
	OpHashGt    = OpGt.Hash()
	OpHashIn    = OpIn.Hash()
)
