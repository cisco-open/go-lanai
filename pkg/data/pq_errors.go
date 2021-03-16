package data

import (
	"context"
	"github.com/lib/pq"
)

type PqErrorTranslator struct{}

func NewPqErrorTranslator() *PqErrorTranslator {
	return &PqErrorTranslator{}
}

func (t PqErrorTranslator) Order() int {
	return ErrorTranslatorOrderPq
}

func (t PqErrorTranslator) Translate(_ context.Context, err error) error {
	// TODO more detailed error translation based on pq.ErrorCode
	switch err.(type) {
	case pq.Error, *pq.Error:
		return NewDataError(ErrorCodeOrmMapping, err.Error(), err)
	default:
		return err
	}
}
