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

package tenancy

import (
	"errors"
	"fmt"
	"strings"
)

const spoPrefix = "spo"
func BuildSpsString(subject string, predict string, object... string) string {
	if len(object) == 0 {
		return fmt.Sprintf("%s:%s:%s", spoPrefix, subject, predict)
	} else {
		return fmt.Sprintf("%s:%s:%s:%s", spoPrefix, subject, predict, object[0])
	}
}

func GetObjectOfSpo(spo string) (string, error) {
	parts := strings.Split(spo, ":")

	if len(parts) == 4 {
		return parts[3], nil
	} else {
		return "", errors.New("spo relation has no object part")
	}
}

func ZInclusive(min string) string {
	return fmt.Sprintf("[%s", min)
}