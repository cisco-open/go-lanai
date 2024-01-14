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

package apidocs

import (
	"context"
	"fmt"
	"dario.cat/mergo"
)

const (
	kOAS3Servers = `servers`
)

func merge(_ context.Context, docs []*apidoc) (*apidoc, error) {
	merged := make(map[string]interface{})
	for _, doc := range docs {
		if e := mergo.Merge(&merged, doc.value, mergo.WithAppendSlice); e != nil {
			return nil, fmt.Errorf("failed to merge [%s]: %v", doc.source, e)
		}
	}

	return &apidoc{
		source: ResolveArgs.Output,
		value:  postMerge(merged),
	}, nil
}

func postMerge(doc map[string]interface{}) map[string]interface{} {
	delete(doc, kOAS3Servers)
	return doc
}

func writeMergedToFile(ctx context.Context, doc *apidoc) error {
	doc.source = ResolveArgs.Output
	absPath, e := writeApiDocLocal(ctx, doc)
	if e != nil {
		return e
	}
	logger.Infof("Merged API document saved to [%s]", absPath)
	return nil
}
