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
	"github.com/robfig/cron/v3"
)

var cronOptions = cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional

// Cron schedules a task using CRON expression
// Supported CRON expression is "<second> <minutes> <hours> <day of month> <month> [day of week]",
// where "day of week" is optional
// Note 1: do not support 'L'
// Note 2: any options affecting start time and repeat rate (StartAt, AtRate, etc.) would take no effect
func Cron(expr string, taskFunc TaskFunc, opts ...TaskOptions) (TaskCanceller, error) {
	opts = append([]TaskOptions{TaskHooks(defaultTaskHooks...)}, opts...)
	opts = append(opts, withCronExpression(expr))
	return newTask(taskFunc, opts...)
}

func withCronExpression(expr string) TaskOptions {
	return func(opt *TaskOption) error {
		nextFn, e := cronNextFunc(expr)
		if e != nil {
			return e
		}
		return dynamicNext(nextFn)(opt)
	}
}

func cronNextFunc(expr string) (nextFunc, error) {
	schedule, e := cron.NewParser(cronOptions).Parse(expr)
	if e != nil {
		return nil, e
	}
	return schedule.Next, nil
}