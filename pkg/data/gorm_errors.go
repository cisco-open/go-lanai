package data

import (
	"context"
)



type GormErrorTranslator struct {}

func NewGormErrorTranslator() *GormErrorTranslator {
	return &GormErrorTranslator{}
}

func (t GormErrorTranslator) Translate(ctx context.Context, err error) error {
	panic("implement me")
}
