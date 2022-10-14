package web

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

type TestController struct {
	ID uuid.UUID
}

func NewTestController() *TestController {
	ret := &TestController{}
	return ret
}

func NewTestControllerAsController() Controller {
	return &TestController{}
}

func (c *TestController) Mappings() []Mapping {
	return []Mapping{}
}

func (c *TestController) Test(_ context.Context, _ *http.Request) (response interface{}, err error) {
	return map[string]string{
		"message": "ok",
	}, nil
}

type NotController struct {
}

func NewNotController() *NotController {
	return &NotController{}
}

type controllerPtrDI struct {
	fx.In
	Controllers []Controller `group:"controllers"`
}

func TestFxControllerProviderWorksForImpl(t *testing.T) {
	app := fx.New(
		FxControllerProviders(NewTestController),
		fx.Invoke(func(di controllerPtrDI) {
			if len(di.Controllers) != 1 {
				t.Error("expect a Controller interface to be provided as value group")
			}
		}),
	)

	ctx := context.TODO()
	err := app.Start(ctx)

	if err != nil {
		t.Error("fx failed to start")
	}

	err = app.Stop(ctx)

	if err != nil {
		t.Error("fx failed to stop")
	}
}

func TestFxControllerProviderError(t *testing.T) {
	app := fx.New(
		FxControllerProviders(NewNotController),
		fx.Invoke(func(di controllerPtrDI) {}),
	)

	ctx := context.TODO()
	err := app.Start(ctx)

	if err == nil {
		t.Error("expect fx to fail because the provided ptr does not implement Controller interface")
	}

	err = app.Stop(ctx)

	if err != nil {
		t.Error("fx failed to stop")
	}
}

type controllerDI struct {
	fx.In
	Controllers []Controller `group:"controllers"`
}

func TestFxControllerProviderWorksForInterface(t *testing.T) {
	app := fx.New(
		FxControllerProviders(NewTestControllerAsController),
		fx.Invoke(func(di controllerDI) {
			if len(di.Controllers) != 1 {
				t.Error("expect a Controller interface to be provided as value group")
			}
		}),
	)

	ctx := context.TODO()
	err := app.Start(ctx)

	if err != nil {
		t.Error("fx failed to start")
	}

	err = app.Stop(ctx)

	if err != nil {
		t.Error("fx failed to stop")
	}
}
