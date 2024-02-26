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

package data

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	errorutils "github.com/cisco-open/go-lanai/pkg/utils/error"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

var (
	dataIntegrityRegexp = regexp.MustCompile(`\((?P<col>[^()]+)\) *= *\((?P<value>[^()]*)\)`)
)

//goland:noinspection GoNameStartsWithPackageName
// WebDataErrorTranslator implements web.ErrorTranslator
type WebDataErrorTranslator struct{}

//goland:noinspection GoNameStartsWithPackageName
func NewWebDataErrorTranslator() ErrorTranslator {
	return WebDataErrorTranslator{}
}

func (WebDataErrorTranslator) Order() int {
	return ErrorTranslatorOrderData
}

func (t WebDataErrorTranslator) Translate(ctx context.Context, err error) error {
	//nolint:errorlint
	if _, ok := err.(errorutils.ErrorCoder); !ok || !errors.Is(err, ErrorCategoryData) {
		return err
	}

	switch {
	case errors.Is(err, ErrorRecordNotFound), errors.Is(err, ErrorIncorrectRecordCount):
		return t.errorWithStatusCode(ctx, err, http.StatusNotFound)
	case errors.Is(err, ErrorSubTypeDataIntegrity):
		return t.dataIntegrityErrorWithStatusCode(ctx, err, http.StatusConflict)
	case errors.Is(err, ErrorSubTypeQuery):
		return t.errorWithStatusCode(ctx, err, http.StatusBadRequest)
	case errors.Is(err, ErrorSubTypeTimeout):
		return t.errorWithStatusCode(ctx, err, http.StatusRequestTimeout)
	case errors.Is(err, ErrorTypeTransient):
		return t.errorWithStatusCode(ctx, err, http.StatusServiceUnavailable)
	default:
		return t.errorWithStatusCode(ctx, err, http.StatusInternalServerError)
	}
}

//nolint:errorlint
func (t WebDataErrorTranslator) errorWithStatusCode(_ context.Context, err error, sc int) error {
	return NewErrorWithStatusCode(err.(DataError), sc)
}

//nolint:errorlint
func (t WebDataErrorTranslator) dataIntegrityErrorWithStatusCode(_ context.Context, err error, sc int) error {
	switch err.(DataError).RootCause().(type) {
	case *pgconn.PgError, *pq.Error:
	default:
		return NewErrorWithStatusCode(err.(DataError), sc)
	}
	msg := "duplicate keys"
	matches := dataIntegrityRegexp.FindStringSubmatch(err.Error())
	for i, name := range dataIntegrityRegexp.SubexpNames() {
		if i >= len(matches) {
			break
		}
		if name == "value" {
			if matches[i] != "" {
				msg = fmt.Sprintf("duplicate value: %s", matches[i])
			}
			break
		}
	}
	return NewErrorWithStatusCode(err.(DataError), sc).WithMessage(msg)
}
