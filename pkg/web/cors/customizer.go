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

package cors

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/rs/cors"
	"time"
)

// Customizer implements web.Customizer
type Customizer struct {
	properties CorsProperties
}

func newCustomizer(properties CorsProperties) web.Customizer {
	return &Customizer{
		properties: properties,
	}
}

func (c *Customizer) Customize(ctx context.Context, r *web.Registrar) (err error) {
	if !c.properties.Enabled {
		return
	}

	mw := New(cors.Options{
		AllowedOrigins:     c.properties.AllowedOrigins(),
		AllowedMethods:     c.properties.AllowedMethods(),
		AllowedHeaders:     c.properties.AllowedHeaders(),
		ExposedHeaders:     c.properties.ExposedHeaders(),
		MaxAge:             int(time.Duration(c.properties.MaxAge).Seconds()),
		AllowCredentials:   c.properties.AllowCredentials,
		OptionsPassthrough: false,
		//Debug:              true,
	})
	err = r.AddGlobalMiddlewares(mw)
	return
}

