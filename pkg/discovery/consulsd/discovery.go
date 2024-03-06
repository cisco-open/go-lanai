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

package consulsd

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"sync"
)

type ClientOptions func(opt *ClientConfig)

type ClientConfig struct {
	Logger          log.Logger
	Verbose         bool
	DefaultSelector discovery.InstanceMatcher
}

type consulDiscoveryClient struct {
	ctx        context.Context
	conn       *consul.Connection
	instancers map[string]*Instancer
	mutex      sync.Mutex
	config     ClientConfig
}

func NewDiscoveryClient(ctx context.Context, conn *consul.Connection, opts ...ClientOptions) discovery.Client {
	if ctx == nil {
		panic("creating ConsulDiscoveryClient with nil context")
	}

	client := consulDiscoveryClient{
		ctx:        ctx,
		conn:       conn,
		instancers: map[string]*Instancer{},
		config: ClientConfig{
			Logger:  logger.WithContext(ctx),
			Verbose: false,
		},
	}
	for _, fn := range opts {
		fn(&client.config)
	}

	return &client
}

func (c *consulDiscoveryClient) Context() context.Context {
	return c.ctx
}

func (c *consulDiscoveryClient) Instancer(serviceName string) (discovery.Instancer, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("empty service name")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	instancer, ok := c.instancers[serviceName]
	if ok {
		return instancer, nil
	}
	instancer = NewInstancer(c.ctx, func(opt *InstancerOption) {
		opt.ConsulConnection = c.conn
		opt.Name = serviceName
		opt.Logger = c.config.Logger
		opt.Verbose = c.config.Verbose
		opt.Selector = c.config.DefaultSelector
	})
	c.instancers[serviceName] = instancer

	return instancer, nil
}

func (c *consulDiscoveryClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, v := range c.instancers {
		v.Stop()
	}
	return nil
}
