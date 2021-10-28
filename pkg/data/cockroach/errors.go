package cockroach

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"github.com/jackc/pgconn"
	"github.com/lib/pq"
)

// PqErrorTranslator translate pq.Error and pgconn.PgError to data.DataError
// Note: cockroach uses gorm.io/driver/postgres, which internally uses github.com/jackc/pgx
// Ref:
// - Postgres Error: https://www.postgresql.org/docs/11/protocol-error-fields.html
// - Postgres Error Code: https://www.postgresql.org/docs/11/errcodes-appendix.html
type PqErrorTranslator struct{}

func NewPqErrorTranslator() PqErrorTranslator {
	return PqErrorTranslator{}
}

func (t PqErrorTranslator) Order() int {
	return 0
}

func (t PqErrorTranslator) Translate(_ context.Context, err error) error {
	var ec int64
	switch e := err.(type) {
	case *pgconn.PgError:
		ec = t.translateErrorCode(e.Code)
	case pq.Error:
		ec = t.translateErrorCode(string(e.Code))
	case *pq.Error:
		ec = t.translateErrorCode(string(e.Code))
	default:
		return err
	}
	return data.NewDataError(ec, err)
}

// translateErrorCode translate postgres error code to data.DataError code
// ref https://www.postgresql.org/docs/11/errcodes-appendix.html
func (t PqErrorTranslator) translateErrorCode(code string) int64 {
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
		return data.ErrorSubTypeCodeQuery
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
