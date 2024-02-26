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
    "context"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "strings"
)

/*****************************
	Common Template Matchers
 *****************************/

const patternWithFilePrefix = `**/%s.*.tmpl`

// matchPatterns match template path with wildcard patterns. e.g. **/project.*.tmpl
func matchPatterns(patterns ...string) TemplateMatcher {
    return &cmdutils.GenericMatcher[TemplateDescriptor]{
        Description: strings.Join(patterns, ", "),
        MatchFunc: func(ctx context.Context, tmplDesc TemplateDescriptor) (bool, error) {
            for _, pattern := range patterns {
                if match, e := cmdutils.MatchPathPattern(pattern, tmplDesc.Path); e == nil && match {
                    return true, nil
                }
            }
            return false, nil
        },
    }
}

func isDir() TemplateMatcher {
    return &cmdutils.GenericMatcher[TemplateDescriptor]{
        Description: "directory",
        MatchFunc: func(ctx context.Context, tmplDesc TemplateDescriptor) (bool, error) {
            return tmplDesc.FileInfo.IsDir(), nil
        },
    }
}

func isTmplFile() TemplateMatcher {
    return &cmdutils.GenericMatcher[TemplateDescriptor]{
        Description: "file",
        MatchFunc: func(ctx context.Context, tmplDesc TemplateDescriptor) (bool, error) {
            return !tmplDesc.FileInfo.IsDir(), nil
        },
    }
}
