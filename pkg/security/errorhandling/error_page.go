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

package errorhandling

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"errors"
	"fmt"
	"net/http"
)

func ErrorWithStatus(ctx context.Context, _ web.EmptyRequest) (int, *template.ModelView, error) {
	s := session.Get(ctx)
	if s == nil {
		err := fmt.Errorf("error message not available")
		return http.StatusInternalServerError, nil, err
	}

	code, codeOk := s.Flash(redirect.FlashKeyPreviousStatusCode).(int)
	if !codeOk {
		code = 500
	}

	err, errOk := s.Flash(redirect.FlashKeyPreviousError).(error)
	if !errOk {
		err = errors.New("unknown error")
	}

	return code, nil, web.NewHttpError(code, err)
}
