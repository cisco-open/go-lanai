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

import "runtime"

type Stacktracer func() (frames []*runtime.Frame, fallback interface{})

// RuntimeStacktracer find stacktrace frames with runtime package
// skip: skip certain number of stacks from the top, including the call to the Stacktracer itself
// depth: max number of stack frames to extract
func RuntimeStacktracer(skip int, depth int) Stacktracer {
    return func() (frames []*runtime.Frame, fallback interface{}) {
        rpc := make([]uintptr, depth)
        count := runtime.Callers(skip, rpc)
        rawFrames := runtime.CallersFrames(rpc)
        frames = make([]*runtime.Frame, count)
        for i := 0; i < count; i++ {
            frame, more := rawFrames.Next()
            frames[i] = &frame
            if !more {
                break
            }
        }
        return frames, nil
    }
}

// RuntimeCaller equivalent to RuntimeStacktracer(skip, 1)
func RuntimeCaller(skip int) Stacktracer {
    return RuntimeStacktracer(skip, 1)
}

