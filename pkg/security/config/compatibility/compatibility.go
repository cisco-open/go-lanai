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

package compatibility

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/security"
)

// CompatibilityDiscoveryCustomizer implements discovery.ServiceRegistrationCustomizer
type CompatibilityDiscoveryCustomizer struct {}

func (c CompatibilityDiscoveryCustomizer) Customize(_ context.Context, reg discovery.ServiceRegistration) {
	tag := fmt.Sprintf("%s=%s", security.CompatibilityReferenceTag, security.CompatibilityReference)
	reg.AddTags(tag)
	reg.SetMeta(security.CompatibilityReferenceTag, security.CompatibilityReference)
}

