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

package generator

import (
    "io/fs"
)

// WithRegenRules Set re-generation rules, Fallback to default mode if no rules matches the output file
func WithRegenRules(rules RegenRules, defaultMode RegenMode) func(o *Option) {
    return func(option *Option) {
        option.RegenRules = rules
        if len(defaultMode) != 0 {
            option.DefaultRegenMode = defaultMode
        }
    }
}

func WithTemplateFS(templateFS fs.FS) func(o *Option) {
    return func(option *Option) {
        option.TemplateFS = templateFS
    }
}

// WithProject general information about the project to generate
func WithProject(project Project) func(o *Option) {
    return func(o *Option) {
        o.Project = project
    }
}

// WithComponents defines what to generate and their settings
func WithComponents(comps Components) func(o *Option) {
    return func(o *Option) {
        o.Components = comps
    }
}



