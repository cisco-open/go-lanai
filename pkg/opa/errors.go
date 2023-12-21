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

package opa

import (
	"errors"
	"fmt"
)

var (
	ErrInternal           = NewError("internal error")
	ErrAccessDenied       = NewError("Access Denied")
	ErrQueriesNotResolved = NewError(`OPA cannot resolve partial queries`)
)

var errorCode int

type Error struct {
	code int
	msg  string
}

func (e Error) Error() string {
	return e.msg
}

func (e Error) Is(err error) bool {
	var v Error
	return errors.As(err, &v) && v.code == e.code
}

func (e Error) WithMessage(tmpl string, args ...interface{}) Error {
	return Error{
		code: e.code,
		msg:  fmt.Sprintf(tmpl, args...),
	}
}

func NewError(tmpl string, args ...interface{}) Error {
	errorCode++
	return Error{
		code: errorCode,
		msg:  fmt.Sprintf(tmpl, args...),
	}
}
