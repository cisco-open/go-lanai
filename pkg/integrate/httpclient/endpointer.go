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
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/utils/matcher"
    "github.com/go-kit/kit/endpoint"
    "net/url"
    "time"
)

/***********************
	Dynamic Endpointer
 ***********************/

// EndpointerOptions allows control of endpointCache behavior.
type EndpointerOptions func(*EndpointerOption)

type EndpointerOption struct {
	ServiceName       string
	EndpointFactory   EndpointFactory
	Selector          discovery.InstanceMatcher
	InvalidateOnError bool
	InvalidateTimeout time.Duration
	Logger            log.ContextualLogger
}

// EndpointerConfig is a subset of EndpointerOption, which can be changed after endpointer is created
type EndpointerConfig struct {
	EndpointFactory   EndpointFactory
	Selector          discovery.InstanceMatcher
	Logger            log.ContextualLogger
}

// KitEndpointer implements sd.Endpointer interface and works with discovery.Instancer.
// When created with NewKitEndpointer function, it automatically registers
// as a subscriber to events from the Instances and maintains a list
// of active Endpoints.
type KitEndpointer struct {
	instancer discovery.Instancer
	selector  discovery.InstanceMatcher
	timeout   time.Duration
	logger    log.ContextualLogger
	factory   EndpointFactory
}

// NewKitEndpointer creates a slim custom sd.Endpointer that work with discovery.Instancer
// and uses factory f to create Endpoints. If src notifies of an error, the Endpointer
// keeps returning previously created Endpoints assuming they are still good, unless
// this behavior is disabled via InvalidateOnError option.
func NewKitEndpointer(instancer discovery.Instancer, opts ...EndpointerOptions) (*KitEndpointer, error) {
	opt := EndpointerOption{
		Selector: matcher.Any(),
		Logger: logger,
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

	// create and start
	ret := &KitEndpointer{
		instancer: instancer,
		selector:  opt.Selector,
		factory:   opt.EndpointFactory,
		logger:    opt.Logger,
		timeout:   opt.InvalidateTimeout,
	}
	return ret, nil
}

func (ke *KitEndpointer) WithConfig(cfg *EndpointerConfig) Endpointer {
	cp := ke.shallowCopy()
	if cfg.Logger != nil {
		cp.logger = cfg.Logger
	}

	if cfg.Selector != nil {
		cp.selector = cfg.Selector
	}

	if cfg.EndpointFactory != nil {
		cp.factory = cfg.EndpointFactory
	}
	return cp
}

// Endpoints implements sd.Endpointer.
func (ke *KitEndpointer) Endpoints() ([]endpoint.Endpoint, error) {
	if ke.factory == nil {
		return nil, NewInternalError("endpoint is not properly configured: endpoint factory not set")
	}
	return ke.endpoints()
}

// handleError is NOT goroutine-safe and returns a boolean indicating last known endpoints should be returned
func (ke *KitEndpointer) handleError(svc *discovery.Service) bool {
	switch {
	case ke.timeout < 0:
		// nothing to do
		return true
	case ke.timeout == 0 || svc.FirstErrAt.IsZero():
		// do not return last known
		return false
	default:
		return svc.FirstErrAt.Add(ke.timeout).Before(time.Now())
	}
}

func (ke *KitEndpointer) endpoints() (ret []endpoint.Endpoint, err error) {
	svc := ke.instancer.Service()
	if svc == nil {
		return nil, NewNoEndpointFoundError(fmt.Errorf("cannot find service [%s]", ke.instancer.ServiceName()))
	} else if svc.Err != nil && !ke.handleError(svc) {
		return nil, NewDiscoveryDownError(fmt.Errorf("cannot find service [%s]", ke.instancer.ServiceName()), svc.Err)
	}

	// prepare endpoints
	insts := svc.Instances(ke.selector)
	ret = make([]endpoint.Endpoint, len(insts))
	for i, inst := range insts {
		// create endpoint
		if ep, e := ke.factory(inst); e != nil {
			ret[i] = ke.errorEndpoint(e)
		} else {
			ret[i] = ep
		}
	}
	return
}

// errorEndpoint makes a dummy endpoint that logs error
func (ke *KitEndpointer) errorEndpoint(err error) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		e := fmt.Errorf("error creating HTTP endpoint: %v", err)
		ke.logger.WithContext(ctx).Errorf(e.Error())
		return nil, NewInternalError(e, err)
	}
}

func (ke *KitEndpointer) shallowCopy() *KitEndpointer {
	return &KitEndpointer{
		instancer: ke.instancer,
		selector:  ke.selector,
		timeout:   ke.timeout,
		logger:    ke.logger,
		factory:   ke.factory,
	}
}

/***********************
	Simple Endpointer
 ***********************/

type SimpleEndpointer struct {
	baseUrls []*url.URL
	factory  EndpointFactory
}

func NewSimpleEndpointer(baseUrls ...string) (*SimpleEndpointer, error) {
	urls := make([]*url.URL, len(baseUrls))
	for i, base := range baseUrls {
		uri, e := url.Parse(base)
		if e != nil {
			return nil, e
		} else if !uri.IsAbs() {
			return nil, fmt.Errorf(`expect abslolute base URL, but got "%s"`, base)
		}
		urls[i] = uri
	}
	return &SimpleEndpointer{
		baseUrls: urls,
	}, nil
}

func (se *SimpleEndpointer) Endpoints() (ret []endpoint.Endpoint, err error) {
	if se.factory == nil {
		return nil, NewInternalError("endpoint is not properly configured: endpoint factory not set")
	}

	ret = make([]endpoint.Endpoint, len(se.baseUrls))
	for i, base := range se.baseUrls {
		ep, e := se.factory(base)
		if e != nil {
			return nil, NewNoEndpointFoundError(fmt.Errorf("cannot create endpoint for base URL [%s]", base))
		}
		ret[i] = ep
	}
	return
}

func (se *SimpleEndpointer) WithConfig(cfg *EndpointerConfig) Endpointer {
	return &SimpleEndpointer{
		baseUrls: se.baseUrls,
		factory: cfg.EndpointFactory,
	}
}


