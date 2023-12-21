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
	"encoding/base64"
	"fmt"
	"github.com/crewjam/saml/samlsp"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

//This implementation is similar to that found in samlsp.CookieRequestTracker
//However, we don't need a reference to a ServiceProvider instance
//and we set the cookie's Secure and SsoPath attribute explicitly
//and we let the cookie's domain be determined by the request itself.
//This is because our tracker needs to work with multiple ServiceProvider instances each talking to a different idp.

// CookieRequestTracker tracks requests by setting a uniquely named
// cookie for each request.
type CookieRequestTracker struct {
	NamePrefix      string
	Codec           samlsp.TrackedRequestCodec
	MaxAge          time.Duration
	SameSite        http.SameSite
	Secure			bool
	Path 			string
}

// TrackRequest starts tracking the SAML request with the given ID. It returns an
// `index` that should be used as the RelayState in the SAMl request flow.
func (t CookieRequestTracker) TrackRequest(w http.ResponseWriter, r *http.Request, samlRequestID string) (string, error) {
	trackedRequest := samlsp.TrackedRequest{
		Index:         base64.RawURLEncoding.EncodeToString([]byte(uuid.New().String())),
		SAMLRequestID: samlRequestID,
		URI:           r.URL.String(),
	}
	signedTrackedRequest, err := t.Codec.Encode(trackedRequest)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     t.NamePrefix + trackedRequest.Index,
		Value:    signedTrackedRequest,
		MaxAge:   int(t.MaxAge.Seconds()),
		HttpOnly: true,
		SameSite: t.SameSite,
		Secure:   t.Secure,
		Path:     t.Path,
	})

	return trackedRequest.Index, nil
}

// StopTrackingRequest stops tracking the SAML request given by index, which is a string
// previously returned from TrackRequest
func (t CookieRequestTracker) StopTrackingRequest(w http.ResponseWriter, r *http.Request, index string) error {
	cookie, err := r.Cookie(t.NamePrefix + index)
	if err != nil {
		return err
	}
	cookie.Value = ""
	cookie.Expires = time.Unix(1, 0) // past time as close to epoch as possible, but not zero time.Time{}
	cookie.Path = t.Path
	cookie.Secure = t.Secure
	cookie.SameSite = t.SameSite
	cookie.HttpOnly = true
	http.SetCookie(w, cookie)
	return nil
}

// GetTrackedRequests returns all the pending tracked requests
func (t CookieRequestTracker) GetTrackedRequests(r *http.Request) []samlsp.TrackedRequest {
	var rv []samlsp.TrackedRequest
	for _, cookie := range r.Cookies() {
		if !strings.HasPrefix(cookie.Name, t.NamePrefix) {
			continue
		}

		trackedRequest, err := t.Codec.Decode(cookie.Value)
		if err != nil {
			continue
		}
		index := strings.TrimPrefix(cookie.Name, t.NamePrefix)
		if index != trackedRequest.Index {
			continue
		}

		rv = append(rv, *trackedRequest)
	}
	return rv
}

// GetTrackedRequest returns a pending tracked request.
func (t CookieRequestTracker) GetTrackedRequest(r *http.Request, index string) (*samlsp.TrackedRequest, error) {
	cookie, err := r.Cookie(t.NamePrefix + index)
	if err != nil {
		return nil, err
	}

	trackedRequest, err := t.Codec.Decode(cookie.Value)
	if err != nil {
		return nil, err
	}
	if trackedRequest.Index != index {
		return nil, fmt.Errorf("expected index %q, got %q", index, trackedRequest.Index)
	}
	return trackedRequest, nil
}
