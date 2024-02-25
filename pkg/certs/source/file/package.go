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

package filecerts

import (
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/certs"
    certsource "github.com/cisco-open/go-lanai/pkg/certs/source"
    "go.uber.org/fx"
)

const (
	sourceType = certs.SourceFile
)

type factoryDI struct {
	fx.In
	Props certs.Properties `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: certs.FxGroup,
		Target: func(di factoryDI) (certs.SourceFactory, error) {
			var rawDefaults json.RawMessage
			if di.Props.Sources != nil {
				rawDefaults, _ = di.Props.Sources[sourceType]
			}
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, NewFileProvider)
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
