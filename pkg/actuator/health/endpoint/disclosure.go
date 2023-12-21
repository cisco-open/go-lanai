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

package healthep

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
)

// DefaultDisclosureControl implements health.DetailsDisclosureControl and health.ComponentsDisclosureControl
type DefaultDisclosureControl struct {
	showDetails         health.ShowMode
	showComponents      health.ShowMode
	permissions         utils.StringSet
	detailsCtrlDelegate health.DetailsDisclosureControl
	compsCtrlDelegate   health.ComponentsDisclosureControl
}

func newDefaultDisclosureControl(props *health.HealthProperties,
	detailsDelegate health.DetailsDisclosureControl,
	compsDelegate health.ComponentsDisclosureControl) (*DefaultDisclosureControl, error) {

	showComponents := props.ShowDetails
	if props.ShowComponents != nil {
		showComponents = *props.ShowComponents
	}
	// check some errors
	switch {
	case props.ShowDetails == health.ShowModeCustom && detailsDelegate == nil:
		return nil, errors.New(`health details control is set to custom but there is no health.ComponentsDisclosureControl configured`)
	case showComponents == health.ShowModeCustom && compsDelegate == nil:
		return nil, errors.New(`health components control is set to custom but there is no health.DetailsDisclosureControl configured`)
	}
	return &DefaultDisclosureControl{
		showDetails:         props.ShowDetails,
		showComponents:      showComponents,
		permissions:         utils.NewStringSet(props.Permissions...),
		detailsCtrlDelegate: detailsDelegate,
		compsCtrlDelegate:   compsDelegate,
	}, nil
}

func (c *DefaultDisclosureControl) ShouldShowDetails(ctx context.Context) bool {
	switch c.showDetails {
	case health.ShowModeNever:
		return false
	case health.ShowModeAlways:
		return true
	case health.ShowModeAuthorized:
		return c.isAuthorized(ctx)
	default:
		return c.detailsCtrlDelegate.ShouldShowDetails(ctx)
	}
}

func (c *DefaultDisclosureControl) ShouldShowComponents(ctx context.Context) bool {
	switch c.showComponents {
	case health.ShowModeNever:
		return false
	case health.ShowModeAlways:
		return true
	case health.ShowModeAuthorized:
		return c.isAuthorized(ctx)
	default:
		return c.compsCtrlDelegate.ShouldShowComponents(ctx)
	}
}

func (c *DefaultDisclosureControl) isAuthorized(ctx context.Context) bool {
	auth := security.Get(ctx)
	if auth.State() < security.StateAuthenticated || auth.Permissions() == nil {
		return false
	}
	for p, _ := range c.permissions {
		if _, ok := auth.Permissions()[p]; !ok {
			return false
		}
	}

	return true
}
