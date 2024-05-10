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

package postgresql

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/lib/pq"
)

var (
	dataIntegrityRegexp = regexp.MustCompile(`\((?P<col>[^()]+)\) *= *\((?P<value>[^()]*)\)`)
)

// PostgresErrorTranslator implements data.ErrorTranslator
// it translates pq.Error and pgconn.PgError to data.DataError
// Note: cockroach uses gorm.io/driver/postgres, which internally uses github.com/jackc/pgx
// Ref:
// - Postgres Error: https://www.postgresql.org/docs/11/protocol-error-fields.html
// - Postgres Error Code: https://www.postgresql.org/docs/11/errcodes-appendix.html
type PostgresErrorTranslator struct{}

func NewPqErrorTranslator() data.ErrorTranslator {
	return data.DefaultGormErrorTranslator{
		ErrorTranslator: PostgresErrorTranslator{},
	}
}

func (t PostgresErrorTranslator) Order() int {
	return 0
}

func (t PostgresErrorTranslator) Translate(_ context.Context, err error) error {
	var ec int64
	//nolint:errorlint // we don't consider wrapped error here
	switch e := err.(type) {
	case *pgconn.PgError:
		ec = t.translateErrorCode(e.Code)
	case *pq.Error:
		ec = t.translateErrorCode(string(e.Code))
	default:
		return err
	}
	de := data.NewDataError(ec, err)

	switch {
	case errors.Is(de, data.ErrorDuplicateKey):
		return t.translateDuplicateKeyErrorMessage(de)
	default:
		return de
	}
}

func (t PostgresErrorTranslator) translateDuplicateKeyErrorMessage(e data.DataError) data.DataError {
	cause := e.Cause()
	var details string
	//nolint:errorlint // we don't consider wrapped error here
	switch ce := cause.(type) {
	case *pgconn.PgError:
		details = ce.Detail
	case *pq.Error:
		details = ce.Detail
	default:
		return e
	}

	msg := "duplicate keys"
	colMsg := ""
	valMsg := ""
	matches := dataIntegrityRegexp.FindStringSubmatch(details)
	for i, name := range dataIntegrityRegexp.SubexpNames() {
		if i >= len(matches) {
			break
		}
		if name == "value" {
			if matches[i] != "" {
				valMsg = fmt.Sprintf("duplicate keys: %s", matches[i])
			}
		}
		if name == "col" {
			if matches[i] != "" {
				colMsg = fmt.Sprintf("duplicate key in column: %s", matches[i])
			}
		}
	}
	if colMsg != "" {
		msg = fmt.Sprintf("%s; %s", msg, colMsg)
	}
	if valMsg != "" {
		msg = fmt.Sprintf("%s; %s", msg, valMsg)
	}
	return e.WithCause(e.Cause(), msg)
}

// translateErrorCode translate postgres error code to data.DataError code
// ref https://www.postgresql.org/docs/11/errcodes-appendix.html
func (t PostgresErrorTranslator) translateErrorCode(code string) int64 {
	// currently we handle selected error classes
	// TODO more detailed error translation
	var errCls string
	if len(code) == 5 {
		errCls = code[:2]
	}
	// for now based on class
	switch errCls {
	// data.ErrorSubTypeCodeQuery
	case "22", "26", "42":
		switch code {
		case "42501":
			return data.ErrorCodeInsufficientPrivilege
		default:
			return data.ErrorSubTypeCodeQuery
		}
	// data.ErrorSubTypeCodeDataRetrieval
	case "24":
		return data.ErrorCodeIncorrectRecordCount
	// data.ErrorSubTypeCodeDataIntegrity
	case "21", "23", "27":
		switch code {
		case "23505":
			return data.ErrorCodeDuplicateKey
		default:
			return data.ErrorCodeConstraintViolation
		}
	// data.ErrorSubTypeCodeTransaction
	case "25", "2D", "2d", "3B", "3b", "40":
		return data.ErrorCodeInvalidTransaction
	// data.ErrorSubTypeCodeSecurity
	case "28":
		return data.ErrorCodeAuthenticationFailed
	// data.ErrorSubTypeCodeConcurrency
	case "55":
		return data.ErrorSubTypeCodeConcurrency
	// data.ErrorTypeCodeTransient
	case "53":
		return data.ErrorTypeCodeTransient
	}

	return data.ErrorTypeCodeUncategorizedServerSide
}
