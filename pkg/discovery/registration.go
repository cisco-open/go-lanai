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

package discovery

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
    netutil "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/net"
    "fmt"
    "github.com/google/uuid"
    "github.com/hashicorp/consul/api"
    "strings"
)

var logger = log.New("Discovery")

const kPropertyContextPath = `server.context-path`

func Register(ctx context.Context, connection *consul.Connection, registration *api.AgentServiceRegistration) error {
	e := connection.Client().Agent().ServiceRegister(registration)
	if e != nil {
        return e
	}
	logger.WithContext(ctx).WithKV(registrationKVs(registration)...).Infof("Register")
    return nil
}

func Deregister(ctx context.Context, connection *consul.Connection, registration *api.AgentServiceRegistration) error {
    e := connection.Client().Agent().ServiceDeregister(registration.ID)
    if e != nil {
        return e
    }
    logger.WithContext(ctx).WithKV(registrationKVs(registration)...).Infof("Deregister")
    return nil
}

func registrationKVs(reg *api.AgentServiceRegistration) []interface{} {
	return []interface{}{
		"service", reg.Name,
		"address", reg.Address,
		"port", reg.Port,
		"tags", reg.Tags,
		"meta", reg.Meta,
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
}

func NewRegistration(opts ...RegistrationOptions) *api.AgentServiceRegistration {
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

	registration := &api.AgentServiceRegistration{
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
	}
	return registration
}

func RegistrationWithProperties(appCtx *bootstrap.ApplicationContext, props DiscoveryProperties) RegistrationOptions {
	return func(cfg *RegistrationConfig) {
		cfg.ApplicationName = appCtx.Name()
		cfg.IPAddress = props.IpAddress
		cfg.NetworkInterface = props.Interface
		cfg.Port = props.Port
		cfg.HealthCheckPath = props.HealthCheckPath
		cfg.HealthScheme = props.Scheme
		cfg.HealthCheckInterval = props.HealthCheckInterval
		cfg.HealthCheckCriticalTimeout = props.HealthCheckCriticalTimeout
		cfg.Tags = createTags(appCtx, props)
	}
}

func createTags(appCtx *bootstrap.ApplicationContext, discoveryProperties DiscoveryProperties) []string {
	tags := make([]string, len(discoveryProperties.Tags), len(discoveryProperties.Tags)+2)
	copy(tags, discoveryProperties.Tags)
	contextPath, _ := appCtx.Value(kPropertyContextPath).(string)
	tags = append(tags,
		fmt.Sprintf("secure=%t", discoveryProperties.Scheme == "https"),
		fmt.Sprintf("contextPath=%s", contextPath),
	)
	return tags
}

var COMPONENT_ATTRIBUTES_MAPPING = map[string]string{
	"serviceName": "application.name",
	"context":     "server.context-path",
	"name":        "info.app.attributes.displayName",
	"description": "info.app.description",
	"parent":      "info.app.attributes.parent",
	"type":        "info.app.attributes.type",
}

type DefaultCustomizer struct {
	instanceUuid        uuid.UUID
	componentAttributes map[string]string
}

func NewDefaultCustomizer(appContext *bootstrap.ApplicationContext) *DefaultCustomizer {
	return &DefaultCustomizer{
		instanceUuid:        uuid.New(),
		componentAttributes: getComponentAttributes(appContext),
	}
}

func (d *DefaultCustomizer) Customize(ctx context.Context, registration *api.AgentServiceRegistration) {
	//The tag was extracted by admin services and injected into DB for UI to list all components
	registration.Tags = append(registration.Tags, kvTag(TAG_INSTANCE_ID, d.instanceUuid.String()), TAG_MANAGED_SERVICE)

	if registration.Meta == nil {
		registration.Meta = make(map[string]string)
	}
	registration.Meta[TAG_INSTANCE_ID] = d.instanceUuid.String()

	var attributeStrings []string
	for k, v := range d.componentAttributes {
		registration.Meta[k] = v
		attributeStrings = append(attributeStrings, fmt.Sprintf("%s:%s", k, v))
	}
	registration.Tags = append(registration.Tags, kvTag(TAG_COMPONENT_ATTRIBUTES, strings.Join(attributeStrings, COMPONENT_ATTRIBUTE_DELIMITER)),
		kvTag(TAG_SERVICE_NAME, d.componentAttributes[TAG_SERVICE_NAME]))
}

func kvTag(k string, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

func getComponentAttributes(appContext *bootstrap.ApplicationContext) map[string]string {
	attributes := make(map[string]string)
	for k, v := range COMPONENT_ATTRIBUTES_MAPPING {
		attributes[k] = fmt.Sprintf("%v", appContext.Value(v))
	}
	return attributes
}
