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

package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"net/http"
	"sort"
)

// ChangeSessionHandler
/**
	This is a high priority handler because it writes to the header.
	Therefore, it must be before any other success handler that may write the response status (e.g. redirect handler)
 */
type ChangeSessionHandler struct{}

func (h *ChangeSessionHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	if !security.IsBeingAuthenticated(from, to) {
		return
	}

	s := Get(c)
	if s == nil {
		return
	}

	//if this is a new session that hasn't been saved, then we don't need to change it
	if s.isNew {
		return
	}

	err := s.ChangeId()

	if err == nil {
		http.SetCookie(rw, NewCookie(s.Name(), s.id, s.options, r))
	} else {
		panic(security.NewInternalError("Failed to update session ID", err))
	}
}

func (h *ChangeSessionHandler) PriorityOrder() int {
	return security.HandlerOrderChangeSession
}

// ConcurrentSessionHandler This handler runs after ChangeSessionHandler so that the updated session id is indexed to the principal
type ConcurrentSessionHandler struct{
	sessionStore          Store
	sessionSettingService SettingService
}

func (h *ConcurrentSessionHandler) HandleAuthenticationSuccess(c context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	if !security.IsBeingAuthenticated(from, to) {
		return
	}

	s := Get(c)
	if s == nil {
		return
	}

	p, err := security.GetUsername(to)
	if err != nil {
		//Auth is something we don't recognize, this indicates a program error
		panic(security.NewInternalError(err.Error()))
	}

	//Adding to the index before checking the limit.
	//If done other way around, concurrent logins may be doing the check before the other request added to the index
	//thus making it possible to exceed the limit
	//By doing the check at the end, we can end up with the right number of sessions when all requests finishes.
	err = h.sessionStore.WithContext(c).AddToPrincipalIndex(p, s)
	if err != nil {
		panic(security.NewInternalError(err.Error()))
	}

	sessionName := s.Name()

	//This will also clean the expired sessions from the index, so we do it regardless if max sessions is set or not
	existing, err := h.sessionStore.WithContext(c).FindByPrincipalName(p, sessionName)

	if err != nil {
		panic(security.NewInternalError(err.Error()))
	}

	max := h.sessionSettingService.GetMaximumSessions(c)

	if len(existing) <= max || max <= 0 {
		return
	}

	sort.SliceStable(existing, func(i, j int) bool {
		return existing[i].createdOn().Before(existing[j].createdOn())
	})

	if e := h.sessionStore.WithContext(c).Invalidate(existing[:len(existing) - max]...); e != nil {
		panic(security.NewInternalError("Cannot delete session that exceeded max concurrent session limit"))
	}
}

func (h *ConcurrentSessionHandler) PriorityOrder() int {
	return security.HandlerOrderConcurrentSession
}

type DeleteSessionOnLogoutHandler struct {
	sessionStore Store
}

func (h *DeleteSessionOnLogoutHandler) HandleAuthenticationSuccess(c context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	if !security.IsBeingUnAuthenticated(from, to) {
		return
	}

	s := Get(c)

	defer func() {
		// clean context
		if mc := utils.FindMutableContext(c); mc != nil {
			mc.Set(contextKeySession, nil)
		}
	}()

	if s == nil {
		return
	}
	if e := h.sessionStore.Invalidate(s); e != nil {
		panic(e)
	}


}