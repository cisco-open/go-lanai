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

package scheduler

import (
	"context"
	"time"
)

const (
	ModeFixedRate = iota
	ModeFixedDelay
	ModeRunOnce
	ModeDynamic
)

type Mode int

type TaskFunc func(ctx context.Context) error

type TaskCanceller interface {
	Cancelled() <-chan error
	Cancel()
}

type TaskOptions func(opt *TaskOption) error
type TaskOption struct {
	name          string
	mode          Mode
	initialTime   time.Time
	interval      time.Duration
	cancelOnError bool
	nextFunc      nextFunc
	hooks         []TaskHook
}

type TaskHook interface {
	// BeforeTrigger is invoked before each time the task is triggered. TaskHook can modify execution context
	BeforeTrigger(ctx context.Context, id string) context.Context

	// AfterTrigger is invoked after the task is triggered and executed
	AfterTrigger(ctx context.Context, id string, err error)
}

// nextFunc is used for ModeDynamic. It's currently unexported and only used for cron impl
type nextFunc func(time.Time) time.Time
