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

package monitor

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	GroupBytesAllocated DataGroup = "BytesAllocated"
	GroupGCPauses       DataGroup = "GcPauses"
	GroupCPUUsage       DataGroup = "CPUUsage"
	GroupPprof          DataGroup = "Pprof"
)

type DataGroup string

type DataStorage interface {
	Read(ctx context.Context, groups ...DataGroup) (map[DataGroup]RawEntries, error)

	// Append save data entry
	Append(ctx context.Context, group DataGroup, entry interface{}, cap int64) error

	// AppendAll save all data entries, grouped by DataGroup
	AppendAll(ctx context.Context, data map[DataGroup]interface{}, cap int64) error
}

type Feed struct {
	Ts             uint64
	BytesAllocated uint64
	GcPause        uint64
	CPUUser        float64
	CPUSys         float64
	Block          int
	Goroutine      int
	Heap           int
	Mutex          int
	Threadcreate   int
}

type RawEntries []string

func (v RawEntries) MarshalJSON() (data []byte, err error) {
	if v == nil {
		return []byte("null"), nil
	}
	return []byte("[" + strings.Join(v, ",") + "]"), nil
}

type SimplePair struct {
	Ts    uint64 `json:"Ts"`
	Value uint64 `json:"Value"`
}

func (p *SimplePair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

type CPUPair struct {
	Ts   uint64  `json:"Ts"`
	User float64 `json:"User"`
	Sys  float64 `json:"Sys"`
}

func (p *CPUPair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

type PprofPair struct {
	Ts           uint64 `json:"Ts"`
	Block        int    `json:"Block"`
	Goroutine    int    `json:"Goroutine"`
	Heap         int    `json:"Heap"`
	Mutex        int    `json:"Mutex"`
	Threadcreate int    `json:"Threadcreate"`
}

func (p *PprofPair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}
