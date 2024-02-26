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

package validation

import (
    "context"
    "encoding"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/go-playground/validator/v10"
)

func TenantAccess() validator.FuncCtx {
	return func(ctx context.Context, fl validator.FieldLevel) bool {
		i := fl.Field().Interface()
		var str string
		switch v := i.(type) {
		case string:
			str = v
		case *string:
			if v != nil {
				str = *v
			}
		case fmt.Stringer:
			str = v.String()
		case encoding.TextMarshaler:
			bytes, _ := v.MarshalText()
			str = string(bytes)
		default:
			// we don't validate non string, just fail it
			return false
		}
		return security.HasAccessToTenant(ctx, str)
	}
}

