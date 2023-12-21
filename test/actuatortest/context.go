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

package actuatortest

type ActuatorOptions func(opt *ActuatorOption)
type ActuatorOption struct {
	// Default to false. When set true, the default health, info and env endpoints are not initialized
	DisableAllEndpoints    bool
	// Default to true. When set to false, the default authentication is installed.
	// Depending on the defualt authentication (currently tokenauth), more dependencies might be needed
	DisableDefaultAuthentication bool
}

// DisableAllEndpoints is an ActuatorOptions that disable all endpoints in test.
// Any endpoint need to be installed manually via apptest.WithModules(...)
func DisableAllEndpoints() ActuatorOptions {
	return func(opt *ActuatorOption) {
		opt.DisableAllEndpoints = true
	}
}

