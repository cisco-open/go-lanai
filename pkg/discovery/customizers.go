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
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/google/uuid"
	"strings"
)

// NewBuildInfoCustomizer returns a ServiceRegistrationCustomizer that extract service's build information
// and put it in tags and metadata
func NewBuildInfoCustomizer() ServiceRegistrationCustomizer {
	return ServiceRegistrationCustomizerFunc(func(_ context.Context, reg ServiceRegistration) {
		attrs := map[string]string{
			TagBuildVersion:  bootstrap.BuildVersion,
			TagBuildDateTime: bootstrap.BuildTime,
		}

		components := strings.Split(bootstrap.BuildVersion, "-")
		if len(components) == 2 {
			attrs[TagBuildNumber] = components[1]
		}

		for k, v := range attrs {
			reg.SetMeta(k, v)
			reg.AddTags(fmt.Sprintf("%s=%s", k, v))
		}
	})
}

var defaultPropertyPaths = map[string]string{
	`serviceName`: `application.name`,
	`context`:     `server.context-path`,
	`name`:        `info.app.attributes.displayName`,
	`description`: `info.app.description`,
	`parent`:      `info.app.attributes.parent`,
	`type`:        `info.app.attributes.type`,
}

// NewPropertiesBasedCustomizer returns a ServiceRegistrationCustomizer that populate tags and metadata
// based on service's loaded properties and the given "propertyPaths".
// "propertyPaths" is a map that contains metadata key as "key" and its corresponding property path.
func NewPropertiesBasedCustomizer(appCtx *bootstrap.ApplicationContext, propertyPaths map[string]string) ServiceRegistrationCustomizer {
	if propertyPaths == nil {
		propertyPaths = defaultPropertyPaths
	}
	return ServiceRegistrationCustomizerFunc(func(ctx context.Context, reg ServiceRegistration) {
		tags := make([]string, 0, len(propertyPaths))
		attrs := make([]string, 0, len(propertyPaths))
		// static KVs
		id := uuid.New()
		ctxPath, _ := appCtx.Value(`server.context-path`).(string)
		tags = append(tags, kvTag(TagInstanceUUID, id.String()))
		tags = append(tags, kvTag(TagServiceName, appCtx.Name()))
		tags = append(tags, kvTag(TagContextPath, ctxPath))
		reg.SetMeta(TagInstanceUUID, id)
		reg.SetMeta(TagServiceName, appCtx.Name())
		reg.SetMeta(TagContextPath, ctxPath)

		// extract properties
		for key, path := range propertyPaths {
			value := appCtx.Value(path)
			if value != nil {
				reg.SetMeta(key, value)
				attrs = append(attrs, fmt.Sprintf("%s%s%v", key, ComponentAttributeKeyValueSeparator, value))
			}
		}

		// set tags
		tags = append(tags, kvTag(TagComponentAttributes, strings.Join(attrs, ComponentAttributeDelimiter)))
		reg.AddTags(tags...)
	})
}

func kvTag(k string, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}
