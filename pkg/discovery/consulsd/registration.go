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
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
	netutil "github.com/cisco-open/go-lanai/pkg/utils/net"
	"github.com/hashicorp/consul/api"
)

func NewServiceRegistrar(conn *consul.Connection) discovery.ServiceRegistrar {
	return consulServiceRegistrar{
		conn: conn,
	}
}

type consulServiceRegistrar struct {
	conn  *consul.Connection
}

func (r consulServiceRegistrar) Register(ctx context.Context, registration discovery.ServiceRegistration) error {
	reg, ok := registration.(*ServiceRegistration)
	if !ok {
		return fmt.Errorf(`unsupported registration type [%T]`, registration)
	}

	if e := r.conn.Client().Agent().ServiceRegister(&reg.AgentServiceRegistration); e != nil {
		return e
	}
	logger.WithContext(ctx).WithKV(r.registrationKVs(registration)...).Infof("Register")
	return nil
}

func (r consulServiceRegistrar) Deregister(ctx context.Context, registration discovery.ServiceRegistration) error {
	if e := r.conn.Client().Agent().ServiceDeregister(registration.ID()); e != nil {
		return e
	}
	logger.WithContext(ctx).WithKV(r.registrationKVs(registration)...).Infof("Deregister")
	return nil
}

func (r consulServiceRegistrar) registrationKVs(reg discovery.ServiceRegistration) []interface{} {
	return []interface{}{
		"service", reg.Name(),
		"address", reg.Address(),
		"port", reg.Port(),
		"tags", reg.Tags(),
		"meta", reg.Meta(),
	}
}

// ServiceRegistration implements discovery.ServiceRegistration
type ServiceRegistration struct {
	api.AgentServiceRegistration
}

func (r *ServiceRegistration) ID() string {
	return r.AgentServiceRegistration.ID
}

func (r *ServiceRegistration) Name() string {
	return r.AgentServiceRegistration.Name
}

func (r *ServiceRegistration) Address() string {
	return r.AgentServiceRegistration.Address
}

func (r *ServiceRegistration) Port() int {
	return r.AgentServiceRegistration.Port
}

func (r *ServiceRegistration) Tags() []string {
	return r.AgentServiceRegistration.Tags
}

func (r *ServiceRegistration) Meta() (kvs map[string]any) {
	kvs = make(map[string]any)
	for k, v := range r.AgentServiceRegistration.Meta {
		kvs[k] = v
	}
	return
}

func (r *ServiceRegistration) SetID(id string) {
	r.AgentServiceRegistration.ID = id
}

func (r *ServiceRegistration) SetName(name string) {
	r.AgentServiceRegistration.Name = name
}

func (r *ServiceRegistration) SetAddress(addr string) {
	r.AgentServiceRegistration.Address = addr
}

func (r *ServiceRegistration) SetPort(port int) {
	r.AgentServiceRegistration.Port = port
}

func (r *ServiceRegistration) AddTags(tags ...string) {
	// add non-duplicate tags and preserve their original order
	uniqueTags := utils.NewStringSet(r.AgentServiceRegistration.Tags...)
	for _, t := range tags {
		if uniqueTags.Has(t) {
			continue
		}
		r.AgentServiceRegistration.Tags = append(r.AgentServiceRegistration.Tags, t)
		uniqueTags.Add(t)
	}
}

func (r *ServiceRegistration) RemoveTags(tags ...string) {
	var head int
	for i := range r.AgentServiceRegistration.Tags {
		var found bool
		for j := 0; j < len(tags) && !found; found, j = tags[j] == r.AgentServiceRegistration.Tags[i], j+1 {
		}
		if found {
			continue
		}
		r.AgentServiceRegistration.Tags[head] = r.AgentServiceRegistration.Tags[i]
		head++
	}
	r.AgentServiceRegistration.Tags = r.AgentServiceRegistration.Tags[:head]
}

func (r *ServiceRegistration) SetMeta(key string, value any) {
	if r.AgentServiceRegistration.Meta == nil {
		r.AgentServiceRegistration.Meta = make(map[string]string)
	}
	if value == nil {
		delete(r.AgentServiceRegistration.Meta, key)
	} else {
		r.AgentServiceRegistration.Meta[key] = fmt.Sprintf(`%v`, value)
	}
}

type RegistrationOptions func(cfg *RegistrationConfig)

type RegistrationConfig struct {
	ApplicationName            string
	IPAddress                  string
	NetworkInterface           string
	Port                       int
	Tags                       []string
	HealthCheckPath            string
	HealthPort                 int
	HealthScheme               string
	HealthCheckInterval        string
	HealthCheckCriticalTimeout string
	Customizers                []discovery.ServiceRegistrationCustomizer
}

func NewRegistration(ctx context.Context, opts ...RegistrationOptions) discovery.ServiceRegistration {
	cfg := RegistrationConfig{}
	for _, fn := range opts {
		fn(&cfg)
	}
	if len(cfg.IPAddress) == 0 {
		cfg.IPAddress, _ = netutil.GetIp(cfg.NetworkInterface)
	}
	if cfg.HealthPort == 0 {
		cfg.HealthPort = cfg.Port
	}

	reg := ServiceRegistration{
		AgentServiceRegistration: api.AgentServiceRegistration{
			Kind:    api.ServiceKindTypical,
			ID:      fmt.Sprintf("%s-%d-%x", cfg.ApplicationName, cfg.Port, cryptoutils.RandomBytes(5)),
			Name:    cfg.ApplicationName,
			Tags:    cfg.Tags,
			Port:    cfg.Port,
			Address: cfg.IPAddress,
			Check: &api.AgentServiceCheck{
				HTTP:                           fmt.Sprintf("%s://%s:%d%s", cfg.HealthScheme, cfg.IPAddress, cfg.HealthPort, cfg.HealthCheckPath),
				Interval:                       cfg.HealthCheckInterval,
				DeregisterCriticalServiceAfter: cfg.HealthCheckCriticalTimeout},
		},
	}
	for _, c := range cfg.Customizers {
		c.Customize(ctx, &reg)
	}
	return &reg
}

func RegistrationWithProperties(props *DiscoveryProperties) RegistrationOptions {
	return func(cfg *RegistrationConfig) {
		cfg.IPAddress = props.IpAddress
		cfg.NetworkInterface = props.Interface
		cfg.Port = props.Port
		cfg.HealthCheckPath = props.HealthCheckPath
		cfg.HealthScheme = props.Scheme
		cfg.HealthCheckInterval = props.HealthCheckInterval
		cfg.HealthCheckCriticalTimeout = props.HealthCheckCriticalTimeout
		cfg.Tags = append(cfg.Tags, fmt.Sprintf("secure=%t", props.Scheme == "https"))
		cfg.Tags = append(cfg.Tags, props.Tags...)
	}
}

func RegistrationWithAppContext(appCtx *bootstrap.ApplicationContext) RegistrationOptions {
	return func(cfg *RegistrationConfig) {
		cfg.ApplicationName = appCtx.Name()
	}
}

func RegistrationWithCustomizers(customizers ...discovery.ServiceRegistrationCustomizer) RegistrationOptions {
	return func(cfg *RegistrationConfig) {
		cfg.Customizers = append(cfg.Customizers, customizers...)
	}
}

