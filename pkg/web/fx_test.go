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

type Service struct {
}

type ParamsDI struct {
	fx.In
	S *Service
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
		{
			name: "target uses Fx.In and provides Controller",
			target: func(p ParamsDI) Controller {
				return cIface
			},
			expectedControllers: []interface{}{cIface},
			expectedPanic:       false,
			expectedFxError:     false,
		},
		{
			name: "target uses Fx.In and provides *TestController",
			target: func(p ParamsDI) *TestController {
				return cPtr
			},
			expectedControllers: nil,
			expectedPanic:       true,
			expectedFxError:     false,
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
				fx.Provide(func() *Service { return &Service{} }),
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
