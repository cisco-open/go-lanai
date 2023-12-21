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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	SamplingRate         = time.Second
	SampleMaxSize  int64 = 86400
	SampleProfiles       = utils.NewStringSet("block", "goroutine", "heap", "mutex", "threadcreate")
)

type dataCollector struct {
	mtx          sync.RWMutex
	storage      DataStorage
	process      *process.Process
	prevSysTime  float64
	prevUserTime float64
	prevNumGC    uint32

	// Mutex protected fields
	ticker      *time.Ticker
	canceller   context.CancelFunc
	subscribers map[string]chan Feed
}

func NewDataCollector(storage DataStorage) *dataCollector {
	proc, e := process.NewProcess(int32(os.Getpid()))
	if e != nil {
		panic(e)
	}
	return &dataCollector{
		storage: storage,
		process: proc,
	}
}

func (c *dataCollector) Start(ctx context.Context) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.ticker != nil {
		return
	}

	var cancelCtx context.Context
	cancelCtx, c.canceller = context.WithCancel(ctx)
	c.ticker = time.NewTicker(SamplingRate)
	c.subscribers = map[string]chan Feed{}
	go c.collectFunc(cancelCtx, c.ticker)()
}

func (c *dataCollector) Stop() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.ticker == nil {
		return
	}

	if c.canceller != nil {
		c.canceller()
		c.canceller = nil
	}

	for _, ch := range c.subscribers {
		close(ch)
	}
	c.subscribers = nil

	c.ticker.Stop()
	c.ticker = nil
}

func (c *dataCollector) Subscribe() (<-chan Feed, string, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.ticker == nil {
		return nil, "", fmt.Errorf("cannot subscribe before data collector started")
	}
	id := uuid.New().String()
	ch := make(chan Feed)
	c.subscribers[id] = ch
	return ch, id, nil
}

func (c *dataCollector) Unsubscribe(id string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.subscribers == nil {
		return
	}
	if ch, ok := c.subscribers[id]; ok {
		close(ch)
		delete(c.subscribers, id)
	}
}

func (c *dataCollector) collectFunc(ctx context.Context, ticker *time.Ticker) func() {
	return func() {
	LOOP:
		for {
			select {
			case now := <-ticker.C:
				c.collect(ctx, now)
			case <-ctx.Done():
				break LOOP
			}
		}
		c.Stop()
	}
}

func (c *dataCollector) collect(ctx context.Context, now time.Time) {
	// collect facts
	timestamp := uint64(now.Unix()) * 1000
	cpuTimes, e := c.process.Times()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// pprof data
	profiles := make(map[string]*pprof.Profile)
	for _, p := range pprof.Profiles() {
		if SampleProfiles.Has(p.Name()) {
			profiles[p.Name()] = p
		}
	}

	gPprof := &PprofPair{
		Ts:           timestamp,
		Block:        profiles["block"].Count(),
		Goroutine:    profiles["goroutine"].Count(),
		Heap:         profiles["heap"].Count(),
		Mutex:        profiles["mutex"].Count(),
		Threadcreate: profiles["threadcreate"].Count(),
	}

	// CPU usage data

	if e != nil {
		cpuTimes = &cpu.TimesStat{}
	}

	gCpu := &CPUPair{
		Ts:   timestamp,
		User: cpuTimes.User - c.prevUserTime,
		Sys:  cpuTimes.System - c.prevSysTime,
	}
	c.prevUserTime = cpuTimes.User
	c.prevSysTime = cpuTimes.System

	// memory data
	gMemAlloc := &SimplePair{
		Ts:    timestamp,
		Value: ms.Alloc,
	}

	// GC data
	var gGCPause *SimplePair
	var gcPause uint64
	if c.prevNumGC == 0 || c.prevNumGC != ms.NumGC {
		gcPause = ms.PauseNs[(ms.NumGC+255)%256]
		gGCPause = &SimplePair{
			Ts:    timestamp,
			Value: gcPause,
		}
		c.prevNumGC = ms.NumGC
	}

	// Create data
	data := map[DataGroup]interface{}{
		GroupPprof:          gPprof,
		GroupGCPauses:       gGCPause,
		GroupCPUUsage:       gCpu,
		GroupBytesAllocated: gMemAlloc,
	}
	if gGCPause == nil {
		data[GroupGCPauses] = nil
	}

	// Create feed
	feed := Feed{
		Ts:             timestamp,
		BytesAllocated: gMemAlloc.Value,
		GcPause:        gcPause,
		CPUUser:        gCpu.User,
		CPUSys:         gCpu.Sys,
		Block:          gPprof.Block,
		Goroutine:      gPprof.Goroutine,
		Heap:           gPprof.Heap,
		Mutex:          gPprof.Mutex,
		Threadcreate:   gPprof.Threadcreate,
	}

	// Save and broadcast
	if e := c.storage.AppendAll(ctx, data, SampleMaxSize); e != nil {
		logger.Debugf("Failed to save profiling data: %v", e)
	}

	c.mtx.RLock()
	defer c.mtx.RUnlock()
	for _, ch := range c.subscribers {
		ch <- feed
	}
}
