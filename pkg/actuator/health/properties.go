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

package health

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"strings"
)

const (
	HealthPropertiesPrefix = "management.endpoint.health"
)

type HealthProperties struct {
	Status StatusProperties `json:"status"`

	// When to show components. If not specified the 'show-details' setting will be used.
	ShowComponents *ShowMode `json:"show-components"`

	// When to show full health details.
	ShowDetails ShowMode `json:"show-details"`

	// Permisions used to determine whether or not a user is authorized to be shown details.
	// When empty, all authenticated users are authorized.
	Permissions utils.CommaSeparatedSlice `json:"permissions"`
}

type StatusOrders []Status

// encoding.TextUnmarshaler
func (o *StatusOrders) UnmarshalText(data []byte) error {
	result := []Status{}
	split := strings.Split(string(data), ",")
	for _, s := range split {
		s = strings.TrimSpace(s)
		status := StatusUnknown
		if e := status.UnmarshalText([]byte(s)); e != nil {
			return e
		}
		result = append(result, status)
	}
	*o = result
	return nil
}

type StatusProperties struct {
	// Comma-separated list of health statuses in order of severity.
	Orders StatusOrders `json:"order"`

	// Mapping of health statuses to HTTP status codes. By default, registered health
	// statuses map to sensible defaults (for example, UP maps to 200).
	ScMapping map[Status]int `json:"http-mapping"`
}

//NewSessionProperties create a SessionProperties with default values
func NewHealthProperties() *HealthProperties {
	return &HealthProperties{
		Status: StatusProperties{
			Orders: StatusOrders{StatusDown, StatusOutOfService, StatusUp, StatusUnknown},
			ScMapping: map[Status]int{},
		},
		Permissions: []string{},
	}
}

//BindHealthProperties create and bind SessionProperties, with a optional prefix
func BindHealthProperties(ctx *bootstrap.ApplicationContext) HealthProperties {
	props := NewHealthProperties()
	if err := ctx.Config().Bind(props, HealthPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind HealthProperties"))
	}
	return *props
}
