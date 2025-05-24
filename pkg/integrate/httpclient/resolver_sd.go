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
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"net/url"
	"path"
	"time"
)

/****************************
	SD TargetResolver
 ****************************/

// SDOptions allows control of endpointCache behavior.
type SDOptions func(opt *SDOption)

type SDOption struct {
	// Selector Used to filter targets during service discovery.
	// Default: discovery.InstanceIsHealthy()
	Selector discovery.InstanceMatcher
	// InvalidateOnError Whether to return previously known targets in case service discovery is temporarily unavailable
	// Default: true
	InvalidateOnError bool
	// InvalidateTimeout How long to keep previously known targets in case service discovery is temporarily unavailable.
	// < 0:  Always use previously known targets, equivalent to InvalidateOnError = false
	// == 0: Never use previously known targets.
	// > 0:  Use previously known targets for the specified duration since the first error received from SD client
	// Default: -1 if InvalidateOnError = false, 0 if InvalidateOnError = true
	InvalidateTimeout time.Duration
	// Scheme HTTP scheme to use.
	// If not set, the actual scheme is resolved from target instance's Meta/Tags and DefaultScheme value.
	// Possible values: "http", "https", "" (empty string).
	// Default: ""
	Scheme string
	// DefaultScheme Default HTTP scheme to use, if Scheme is not set and resolver cannot resolve scheme from Meta/Tags.
	// Possible values: "http", "https.
	// Default: "http"
	DefaultScheme string
	// ContextPath Path prefix for any given Request.
	// If not set, the context path is resolved from target instance's Meta/Tags.
	// e.g. "/auth/api"
	// Default: ""
	ContextPath string
}

// SDTargetResolver implements TargetResolver interface that use the discovery.Instancer to resolve target's address.
// It also attempts to resolve the http scheme and context path from instance's tags/meta.
// In case of failed service discovery with error, this resolver keeps using previously found instances assuming they
// are still good of period of time configured by SDOption.
// Currently, this resolver only support round-robin load balancing.
type SDTargetResolver struct {
	SDOption
	instancer discovery.Instancer
	balancer  balancer[*discovery.Instance]
}

// NewSDTargetResolver creates a TargetResolver that work with discovery.Instancer.
// See SDTargetResolver
func NewSDTargetResolver(instancer discovery.Instancer, opts ...SDOptions) (*SDTargetResolver, error) {
	opt := SDOption{
		Selector:          discovery.InstanceIsHealthy(),
		DefaultScheme:     "http",
		InvalidateOnError: true,
	}
	for _, f := range opts {
		f(&opt)
	}

	// some validation
	if !opt.InvalidateOnError {
		opt.InvalidateTimeout = -1
	} else if opt.InvalidateTimeout < 0 {
		opt.InvalidateTimeout = 0 // invalidate immediately
	}

	return &SDTargetResolver{
		SDOption:  opt,
		instancer: instancer,
		balancer:  newRoundRobinBalancer[*discovery.Instance](),
	}, nil
}

func (ke *SDTargetResolver) Resolve(_ context.Context, req *Request) (*url.URL, error) {
	svc := ke.instancer.Service()
	if svc == nil {
		return nil, NewNoEndpointFoundError(fmt.Errorf("cannot find service [%s]", ke.instancer.ServiceName()))
	} else if svc.Err != nil && !ke.handleError(svc) {
		return nil, NewDiscoveryDownError(fmt.Errorf("cannot find service [%s]", ke.instancer.ServiceName()), svc.Err)
	}

	// prepare endpoints
	inst, e := ke.balancer.Balance(svc.Instances(ke.Selector))
	if e != nil || inst == nil {
		return nil, NewNoEndpointFoundError(fmt.Errorf("cannot find service [%s]", ke.instancer.ServiceName()))
	}
	return ke.targetURL(inst, req)
}

func (ke *SDTargetResolver) targetURL(inst *discovery.Instance, req *Request) (target *url.URL, err error) {
	ctxPath := ke.ContextPath
	if len(ctxPath) == 0 && inst.Meta != nil {
		ctxPath, _ = inst.Meta[discovery.InstanceMetaKeyContextPath]
	}

	scheme := ke.Scheme
	if len(scheme) == 0 {
		if m, e := httpInstanceMatcher.Matches(inst); m && e == nil {
			scheme = "http"
		} else if m, e := httpsInstanceMatcher.Matches(inst); m && e == nil {
			scheme = "https"
		} else {
			scheme = ke.DefaultScheme
		}
	}

	host := inst.Address
	if inst.Port > 0 && inst.Port <= 0xffff {
		host = fmt.Sprintf("%s:%d", inst.Address, inst.Port)
	}

	target = &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path.Join(ctxPath, req.Path),
	}
	return
}

// handleError is NOT goroutine-safe and returns a boolean indicating last known endpoints should be returned
func (ke *SDTargetResolver) handleError(svc *discovery.Service) bool {
	switch {
	case ke.InvalidateTimeout < 0:
		// nothing to do
		return true
	case ke.InvalidateTimeout == 0 || svc.FirstErrAt.IsZero():
		// do not return last known
		return false
	default:
		return svc.FirstErrAt.Add(ke.InvalidateTimeout).Before(time.Now())
	}
}
