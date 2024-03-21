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

package httpclient

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"net/http"
	"time"
)

const (
	HighestReservedHookOrder  = -10000
	LowestReservedHookOrder   = 10000
	HookOrderTokenPassthrough = HighestReservedHookOrder + 10
	HookOrderRequestLogger    = LowestReservedHookOrder
	HookOrderResponseLogger   = HighestReservedHookOrder
)

const (
	logKey = "remote-http"
)

const (
	kb = 1024
	mb = kb * kb
	gb = mb * kb
)

var ctxKeyStartTime = struct{}{}

/*********************
	Function Alias
 *********************/

// BeforeHookFunc implements Hook with only "before" operation
type BeforeHookFunc func(context.Context, *http.Request) context.Context

func (fn BeforeHookFunc) Before(ctx context.Context, req *http.Request) context.Context {
	return fn(ctx, req)
}

// AfterHookFunc implements Hook with only "after" operation
type AfterHookFunc func(context.Context, *http.Response) context.Context

func (fn AfterHookFunc) After(ctx context.Context, resp *http.Response) context.Context {
	return fn(ctx, resp)
}

/*********************
	Ordered
 *********************/

func BeforeHookWithOrder(order int, hook BeforeHook) BeforeHook {
	return &orderedBeforeHook{
		BeforeHook: hook,
		order:      order,
	}
}

// orderedBeforeHook implements BeforeHook, order.Ordered
type orderedBeforeHook struct {
	BeforeHook
	order int
}

func (h orderedBeforeHook) Order() int {
	return h.order
}

func AfterHookWithOrder(order int, hook AfterHook) AfterHook {
	return &orderedAfterHook{
		AfterHook: hook,
		order:     order,
	}
}

// orderedAfterHook implements AfterHook, order.Ordered
type orderedAfterHook struct {
	AfterHook
	order int
}

func (h orderedAfterHook) Order() int {
	return h.order
}

/****************************
	Token Passthrough Hook
 ****************************/

func HookTokenPassthrough() BeforeHook {
	hook := BeforeHookFunc(func(ctx context.Context, request *http.Request) context.Context {
		authHeader := request.Header.Get(HeaderAuthorization)
		if authHeader != "" {
			return ctx
		}

		auth, ok := security.Get(ctx).(oauth2.Authentication)
		if !ok || !security.IsFullyAuthenticated(auth) || auth.AccessToken() == nil {
			return ctx
		}

		authHeader = fmt.Sprintf("Bearer %s", auth.AccessToken().Value())
		request.Header.Set(HeaderAuthorization, authHeader)
		return ctx
	})
	return BeforeHookWithOrder(HookOrderTokenPassthrough, hook)
}

/*************************
	Logger Hook
 *************************/

type requestLoggerHook struct {
	*ClientConfig
}

func (h requestLoggerHook) Order() int {
	return HookOrderRequestLogger
}

func(h requestLoggerHook) Before(ctx context.Context, req *http.Request) context.Context {
	now := time.Now().UTC()
	logRequest(ctx, req, h.Logger, &h.Logging)
	return context.WithValue(ctx, ctxKeyStartTime, now)
}

func (h requestLoggerHook) WithConfig(cfg *ClientConfig) BeforeHook {
	return requestLoggerHook{ClientConfig: cfg}
}

func HookRequestLogger(cfg *ClientConfig) BeforeHook {
	return requestLoggerHook{}.WithConfig(cfg)
}

type responseLoggerHook struct {
	*ClientConfig
}

func (h responseLoggerHook) Order() int {
	return HookOrderResponseLogger
}

func(h responseLoggerHook) After(ctx context.Context, resp *http.Response) context.Context {
	logResponse(ctx, resp, h.Logger, &h.Logging)
	return ctx
}

func (h responseLoggerHook) WithConfig(cfg *ClientConfig) AfterHook {
	return responseLoggerHook{ClientConfig: cfg}
}

func HookResponseLogger(cfg *ClientConfig) AfterHook {
	return responseLoggerHook{}.WithConfig(cfg)
}
