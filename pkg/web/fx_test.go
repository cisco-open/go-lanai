package web

import (
	"context"
	"errors"
	"go.uber.org/fx"
	"net/http"
	"reflect"
	"testing"
)
import . "github.com/onsi/gomega"

type TestController struct {
	Description string
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

func TestFxControllerProviderWorksForMultipleReturnValues(t *testing.T) {
	cIface := Controller(&TestController{Description: "Provided as Controller"})
	cPtr := &TestController{Description: "Provided as *TestController"}
	notCptr := &NotController{}

	tests := []struct {
		name                string
		target              interface{}
		expectedControllers []interface{}
		expectedPanic       bool
		expectedFxError     bool
	}{
		{
			name: "target provides multiple controllers",
			target: func() (Controller, *TestController, error) {
				return cIface, cPtr, nil
			},
			expectedControllers: []interface{}{cIface, cPtr},
			expectedPanic:       false,
			expectedFxError:     false,
		},
		{
			name: "target provides none controller",
			target: func() (Controller, *TestController, *NotController, error) {
				return cIface, cPtr, notCptr, nil
			},
			expectedControllers: nil,
			expectedPanic:       true,
			expectedFxError:     false,
		},
		{
			name: "target provides no controller",
			target: func() error {
				return nil
			},
			expectedControllers: nil,
			expectedPanic:       true,
			expectedFxError:     false,
		},
		{
			name: "target provides result in error",
			target: func() (Controller, error) {
				return nil, errors.New("error providing controller")
			},
			expectedControllers: nil,
			expectedPanic:       false,
			expectedFxError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			defer func() {
				r := recover()
				g.Expect(r != nil).To(Equal(tt.expectedPanic))
			}()

			app := fx.New(
				FxControllerProviders(tt.target),
				fx.Invoke(func(di controllerDI) {
					g.Expect(len(di.Controllers)).To(Equal(len(tt.expectedControllers)))
					for i := 0; i < len(tt.expectedControllers); i++ {
						found := false
						for j := 0; j < len(di.Controllers); j++ {
							if tt.expectedControllers[i] == di.Controllers[j] {
								found = true
							}
						}
						g.Expect(found).To(BeTrue())
					}
				}),
			)

			ctx := context.TODO()
			err := app.Start(ctx)
			if !tt.expectedFxError {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(HaveOccurred())
			}

			err = app.Stop(ctx)
			g.Expect(err).ToNot(HaveOccurred())
		})
	}
}

//TODO: change this to a loop
func TestValidatingSupportedTarget(t *testing.T) {
	c := &TestController{}
	actualType := reflect.TypeOf(c)

	supported := isSupportedType(typeController, actualType)

	if !supported {
		t.Errorf("expect TestController to be supported because it implements Controller interface")
	}

	notC := &NotController{}
	actualType = reflect.TypeOf(notC)

	supported = isSupportedType(typeController, actualType)

	if supported {
		t.Errorf("expect NotController to not be supported because it does not implement Controller interface")
	}

	var iface Controller
	iface = &TestController{}
	actualType = reflect.TypeOf(iface)

	supported = isSupportedType(typeController, actualType)
	if !supported {
		t.Errorf("expect controller to be supported because it implement Controller interface")
	}

	err := errors.New("some error")
	actualType = reflect.TypeOf(err)

	supported = isSupportedType(typeError, actualType)
	if !supported {
		t.Errorf("expect err to be supported because it implement error interface")
	}
}
