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

package cassandra

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "strings"
    "time"
)

const (
	CassandraPropertiesPrefix = "data.cassandra"
)

type CassandraProperties struct {
	ContactPoints      string        `json:"contact-points"` // comma separated
	Port               int           `json:"port"`
	KeySpaceName	   string `json:"keyspace-name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  utils.Duration `json:"timeout"`
	Consistency string `json:"consistency"`
}

func (p CassandraProperties) Hosts() []string {
	hosts := strings.Split(p.ContactPoints, ",")
	for i, h := range hosts {
		hostParts := strings.SplitN(h, ":", 2)
		if len(hostParts) == 1 {
			hosts[i] = fmt.Sprintf("%s:%d", h, p.Port)
		}
	}
	return hosts
}

func NewCassandraProperties() *CassandraProperties{
	return &CassandraProperties{
		ContactPoints: "127.0.0.1",
		Port: 9042,
		Timeout: utils.Duration(15*time.Second),
		Consistency: "Quorum",
	}
}