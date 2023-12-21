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

import "context"

type AnonymousCandidate map[string]interface{}

// Principal implements security.Candidate
func (ac AnonymousCandidate) Principal() interface{} {
	return "anonymous"
}

// Credentials implements security.Candidate
func (_ AnonymousCandidate) Credentials() interface{} {
	return ""
}

// Details implements security.Candidate
func (ac AnonymousCandidate) Details() interface{} {
	return ac
}

type AnonymousAuthentication struct {
	candidate AnonymousCandidate
}

func (aa *AnonymousAuthentication) Principal() interface{} {
	return aa.candidate.Principal()
}

func (_ *AnonymousAuthentication) Permissions() Permissions {
	return map[string]interface{}{}
}

func (_ *AnonymousAuthentication) State() AuthenticationState {
	return StateAnonymous
}

func (aa *AnonymousAuthentication) Details() interface{} {
	return aa.candidate.Details()
}

type AnonymousAuthenticator struct{}

func (a *AnonymousAuthenticator) Authenticate(_ context.Context, candidate Candidate) (auth Authentication, err error) {
	if ac, ok := candidate.(AnonymousCandidate); ok {
		return &AnonymousAuthentication{candidate: ac}, nil
	}
	return nil, nil
}
