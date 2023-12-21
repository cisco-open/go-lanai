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
	"context"
	"strings"
)

const (
	StatusUnknown Status = iota
	StatusUp
	StatusOutOfService
	StatusDown
)

type Status int

// fmt.Stringer
func (s Status) String() string {
	switch s {
	case StatusUp:
		return "UP"
	case StatusDown:
		return "DOWN"
	case StatusOutOfService:
		return "OUT_OF_SERVICE"
	default:
		return "UNKNOWN"
	}
}

// MarshalText implements encoding.TextMarshaler
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

//UnmarshalText implements encoding.TextUnmarshaler
func (s *Status) UnmarshalText(data []byte) error {
	value := strings.ToUpper(string(data))
	switch value {
	case "UP":
		*s = StatusUp
	case "DOWN":
		*s = StatusDown
	case "OUT_OF_SERVICE":
		*s = StatusOutOfService
	default:
		*s = StatusUnknown
	}
	return nil
}

const (
	// ShowModeNever Never show the item in the response.
	ShowModeNever ShowMode = iota
	// ShowModeAuthorized Show the item in the response when accessed by an authorized user.
	ShowModeAuthorized
	// ShowModeAlways Always show the item in the response.
	ShowModeAlways
	// ShowModeCustom Shows the item in response with a customized rule.
	ShowModeCustom
)

// ShowMode is options for showing items in responses from the HealthEndpoint web extensions.
type ShowMode int

// fmt.Stringer
func (m ShowMode) String() string {
	switch m {
	case ShowModeAuthorized:
		return "authorized"
	case ShowModeAlways:
		return "always"
	case ShowModeCustom:
		return "custom"
	default:
		return "never"
	}
}

// MarshalText implements encoding.TextMarshaler
func (m ShowMode) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (m *ShowMode) UnmarshalText(data []byte) error {
	value := strings.ToLower(string(data))
	switch value {
	case "authorized", "when_authorized", "whenAuthorized", "when-authorized":
		*m = ShowModeAuthorized
	case "always":
		*m = ShowModeAlways
	case "custom":
		*m = ShowModeCustom
	default:
		*m = ShowModeNever
	}
	return nil
}

type Registrar interface {
	// Register configure SystemHealthRegistrar and HealthEndpoint
	// supported input parameters are:
	// 	- Indicator
	// 	- StatusAggregator
	// 	- DetailsDisclosureControl
	// 	- ComponentsDisclosureControl
	//  - DisclosureControl
	Register(items ...interface{}) error

	// MustRegister same as Register, but panic if there is error
	MustRegister(items ...interface{})
}

type StatusAggregator interface {
	Aggregate(context.Context, ...Status) Status
}

type StatusCodeMapper interface {
	StatusCode(context.Context, Status) int
}

type Health interface {
	Status() Status
	Description() string
}

type Options struct {
	ShowDetails    bool
	ShowComponents bool
}

type Indicator interface {
	Name() string
	Health(context.Context, Options) Health
}

type DetailsDisclosureControl interface {
	ShouldShowDetails(ctx context.Context) bool
}

type ComponentsDisclosureControl interface {
	ShouldShowComponents(ctx context.Context) bool
}

type DisclosureControl interface {
	DetailsDisclosureControl
	ComponentsDisclosureControl
}

// DisclosureControlFunc convert function to DisclosureControl
// This type can be registered via Registrar.Register
type DisclosureControlFunc func(ctx context.Context) bool

func (fn DisclosureControlFunc) ShouldShowDetails(ctx context.Context) bool {
	return fn(ctx)
}

func (fn DisclosureControlFunc) ShouldShowComponents(ctx context.Context) bool {
	return fn(ctx)
}