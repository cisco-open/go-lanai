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

package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

/***********************
	Session
************************/

const SessionPropertiesPrefix = "security.session"

type SessionProperties struct {
	Cookie               CookieProperties
	IdleTimeout          utils.Duration `json:"idle-timeout"`
	AbsoluteTimeout      utils.Duration `json:"absolute-timeout"`
	MaxConcurrentSession int            `json:"max-concurrent-sessions"`
	DbIndex              int            `json:"db-index"`
}

type CookieProperties struct {
	Domain         string `json:"domain"`
	MaxAge         int    `json:"max-age"`
	Secure         bool   `json:"secure"`
	HttpOnly       bool   `json:"http-only"`
	SameSiteString string `json:"same-site"`
	Path           string `json:"path"`
}

func (cp CookieProperties) SameSite() http.SameSite {
	switch strings.ToLower(cp.SameSiteString) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

// NewSessionProperties create a SessionProperties with default values
func NewSessionProperties() *SessionProperties {
	return &SessionProperties{
		Cookie: CookieProperties{
			HttpOnly: true,
			Path:     "/",
		},
		IdleTimeout:          utils.Duration(900 * time.Second),
		AbsoluteTimeout:      utils.Duration(1800 * time.Second),
		MaxConcurrentSession: 0, //unlimited
	}
}

// BindSessionProperties create and bind SessionProperties, with a optional prefix
func BindSessionProperties(ctx *bootstrap.ApplicationContext) SessionProperties {
	props := NewSessionProperties()
	if err := ctx.Config().Bind(props, SessionPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SessionProperties"))
	}
	return *props
}

const TimeoutPropertiesPrefix = "security.timeout-support"

type TimeoutSupportProperties struct {
	DbIndex     int    `json:"db-index"`
}

func NewTimeoutSupportProperties() *TimeoutSupportProperties {
	return &TimeoutSupportProperties{}
}

func BindTimeoutSupportProperties(ctx *bootstrap.ApplicationContext) TimeoutSupportProperties {
	props := NewTimeoutSupportProperties()
	if err := ctx.Config().Bind(props, TimeoutPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind TimeoutSupportProperties"))
	}
	return *props
}
