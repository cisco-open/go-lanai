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

package opensearchtest

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
    "fmt"
    "strings"
)

// BulkJsonBodyMatcher special body matcher for OpenSearch's _bulk API
// See https://opensearch.org/docs/2.11/api-reference/document-apis/bulk/
type BulkJsonBodyMatcher struct {
    Delegate ittest.RecordBodyMatcher
}

func (m BulkJsonBodyMatcher) Support(contentType string) bool {
    return m.Delegate.Support(contentType)
}

func (m BulkJsonBodyMatcher) Matches(out []byte, record []byte) error {
    outSplit := m.split(out)
    recordSplit := m.split(record)
    if len(outSplit) != len(recordSplit) {
        return fmt.Errorf(`mismatched number of JSON objects: expect %d but got %d`, len(recordSplit), len(outSplit))
    }
    for i := range outSplit {
        if e := m.Delegate.Matches(outSplit[i], recordSplit[i]); e != nil {
            return fmt.Errorf("mismatched JSON object at index %d: %v", i, e)
        }
    }
    return nil
}

func (m BulkJsonBodyMatcher) split(data []byte) [][]byte {
    split := strings.Split(string(data), "\n")
    rs := make([][]byte, 0, len(split))
    for i := range split {
        trimmed := strings.TrimSpace(split[i])
        if len(trimmed) != 0 {
            rs = append(rs, []byte(trimmed))
        }
    }
    return rs
}