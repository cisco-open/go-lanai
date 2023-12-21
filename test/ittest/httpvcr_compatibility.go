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

package ittest

import (
    "fmt"
    "github.com/ghodss/yaml"
    "net/http"
    "net/url"
    "os"
    "strconv"
)

type V1Cassette struct {
    Version      int             `json:"version"`
    Interactions []V1Interaction `json:"interactions"`
}

type V1Interaction struct {
    Request  map[string]interface{} `json:"request"`
    Response map[string]interface{} `json:"response"`
}

type V2Cassette struct {
    Version      int             `json:"version"`
    Interactions []V2Interaction `json:"interactions"`
}

type V2Interaction struct {
    ID       int                    `json:"id"`
    Request  map[string]interface{} `json:"request"`
    Response map[string]interface{} `json:"response"`
}

// ConvertCassetteFileV1toV2 is a utility function that help with migrating from httpvcr/v3 (using version 1 format)
// to httpvcr/v3 (using version 2 format).
// Note: Usually test authors should re-record interactions instead of using this utility. However, there might be cases
// that re-recording is not possible due to lack of remote server setup.
func ConvertCassetteFileV1toV2(src, dest string) error {
    srcBytes, e := os.ReadFile(src)
    if e != nil {
        return fmt.Errorf("unable to convert record file: %v", e)
    }
    var v1 V1Cassette
    if e := yaml.Unmarshal(srcBytes, &v1); e != nil {
        return fmt.Errorf("unable to convert record file, invalid source file: %v", e)
    }

    v2 := convertCassetteV1ToV2(v1)
    v2bytes, e := yaml.Marshal(v2)
    if e != nil {
        return fmt.Errorf("unable to convert record file: %v", e)
    }

    //nolint:gosec // G306: Expect WriteFile permissions to be 0600 or less - 0600 is too low
    if e := os.WriteFile(dest, v2bytes, 0644); e != nil {
        return fmt.Errorf("unable to convert record file, failed to write to destination: %v", e)
    }
    return nil
}

func convertCassetteV1ToV2(v1 V1Cassette) V2Cassette {
    v2 := V2Cassette{
        Version:      2,
        Interactions: make([]V2Interaction, len(v1.Interactions)),
    }
    for i, record := range v1.Interactions {
        v2.Interactions[i] = convertInteractionV1ToV2(i, record)
    }
    return v2
}

func convertInteractionV1ToV2(id int, v1 V1Interaction) V2Interaction {
    v2 := V2Interaction{
        ID: id,
        Request: v1.Request,
        Response: v1.Response,
    }
    // Add host field to each request (required for matching) if possible
    if rawUrl, ok := v1.Request["url"].(string); ok {
        parsed, e := url.Parse(rawUrl)
        if e == nil {
            v2.Request["host"] = parsed.Host
        }
    }
    // Add Interaction Index if possible
    var headers http.Header
    if rawHeaders, ok := v1.Request["headers"].(map[string][]string); !ok {
        headers = http.Header{}
    } else {
        headers = rawHeaders
    }
    headers.Set(xInteractionIndexHeader, strconv.Itoa(id))
    v2.Request["headers"] = headers
    v2.Request["order"] = id
    // remove duration if empty
    if v, ok := v1.Response["duration"].(string); ok && len(v) == 0 {
        delete(v2.Response, "duration")
    }

    return v2
}
