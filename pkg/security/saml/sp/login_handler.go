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

package sp

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/crewjam/saml/samlsp"
	"net/http"
)

type TrackedRequestSuccessHandler struct {
	tracker samlsp.RequestTracker
}

func NewTrackedRequestSuccessHandler(tracker samlsp.RequestTracker) security.AuthenticationSuccessHandler{
	return &TrackedRequestSuccessHandler{
		tracker: tracker,
	}
}

func (t *TrackedRequestSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	redirectURI := "/"
	if trackedRequestIndex := r.Form.Get("RelayState"); trackedRequestIndex != "" {
		trackedRequest, err := t.tracker.GetTrackedRequest(r, trackedRequestIndex)
		if err == nil {
			redirectURI = trackedRequest.URI
		} else {
			logger.WithContext(c).Errorf("error getting tracked request %v", err)
		}
		_ = t.tracker.StopTrackingRequest(rw, r, trackedRequestIndex)
	}
	http.Redirect(rw, r, redirectURI, 302)
	_,_ = rw.Write([]byte{})
}