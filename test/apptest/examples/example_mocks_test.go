package examples

import (
	"context"
)


type mockedService struct {}

func (t *mockedService) DummyMethod(_ context.Context) error {
	return nil
}

func NewMockedService() DummyService {
	return &mockedService{}
}