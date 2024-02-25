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

package log

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "os"
    "path/filepath"
)

/*
	Common functions that useful to any logger factory
 */

func loggerKey(name string) string {
	return utils.CamelToSnakeCase(name)
}

func convertLevelsNameToKey(byNames map[string]LoggingLevel) (byKeys map[string]LoggingLevel) {
	byKeys = map[string]LoggingLevel{}
	for k, v := range byNames {
		byKeys[loggerKey(k)] = v
	}
	return
}

func openOrCreateFile(location string) (*os.File, error) {
	if location == "" {
		return nil, fmt.Errorf("location is missing for file logger")
	}
	dir := filepath.Dir(location)
	if e := os.MkdirAll(dir, 0744); e != nil {
		return nil, e
	}
	return os.OpenFile(location, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
}
