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

package loop

import (
	"time"
)

// FixedRepeatInterval returns a TaskOptions which set repeat interval to be fixed duration
func FixedRepeatInterval(interval time.Duration) TaskOptions {
	return func(opt *TaskOption) {
		opt.RepeatIntervalFunc = fixedRepeatIntervalFunc(interval)
	}
}

func fixedRepeatIntervalFunc(interval time.Duration) RepeatIntervalFunc {
	return func(_ interface{}, _ error) time.Duration {
		return interval
	}
}

// ExponentialRepeatIntervalOnError returns a TaskOptions
// which set repeat interval to be exponentially increased if error is not nil.
// the repeat interval is reset to "init" if error is nil
func ExponentialRepeatIntervalOnError(init time.Duration, factor float64) TaskOptions {
	if factor < 1 {
		panic("attempt to use ExponentialRepeatIntervalOnError with a factor less than 1")
	}
	return func(opt *TaskOption) {
		opt.RepeatIntervalFunc = exponentialRepeatIntervalOnErrorFunc(init, factor)
	}
}

func exponentialRepeatIntervalOnErrorFunc(init time.Duration, factor float64) RepeatIntervalFunc {
	curr := init
	return func(_ interface{}, err error) time.Duration {
		if err == nil {
			curr = init
		} else {
			curr = time.Duration(float64(curr) * factor)
		}
		return curr
	}
}
