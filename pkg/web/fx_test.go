package web

import (
	"context"
	"errors"
	"go.uber.org/fx"
	"net/http"
	"testing"
)
import . "github.com/onsi/gomega"

type TestController struct {
	Description string
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

type controllerDI struct {
	fx.In
	Controllers []Controller `group:"controllers"`
}

func TestFxControllerProvider(t *testing.T) {
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
			name: "target provides a *TestController",
			target: func() *TestController {
				return cPtr
			},
			expectedControllers: []interface{}{cPtr},
			expectedPanic:       false,
			expectedFxError:     false,
		},
		{
			name: "target provides a none Controller",
			target: func() *NotController {
				return notCptr
			},
			expectedControllers: nil,
			expectedPanic:       true,
			expectedFxError:     false,
		},
		{
			name: "target provides a Controller",
			target: func() Controller {
				return cIface
			},
			expectedControllers: []interface{}{cIface},
			expectedPanic:       false,
			expectedFxError:     false,
		},
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
