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

package testutils

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var RuntimeTest func(t *testing.T)

// PackageDirectory returns package path of the test's Test...() function.
// This function works as following:
// 		It traces back the call stack using runtime,
//		ignoring test utility packages like "test", "test/utils", "testdata" and "runtime,
// 		until it find golang's "testing" package. The last seen package is considered the package of Test...().
//
// Limitation:
//   - This function would not work in tests that have their directory ending with "test" and "test/utils"
//     The workaround is to set RuntimeTest to one of the test function.
//   - THis function would not work in any "testdata" directory. There is no workaround.
func PackageDirectory() string {
	rpc := make([]uintptr, 10)
	if n := runtime.Callers(1, rpc[:]); n < 1 {
		panic("unable find package path")
	}

	var lastPkgDir string
	frames := runtime.CallersFrames(rpc)
	LOOP:
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		dir := filepath.Dir(frame.File)
		switch {
		case strings.HasSuffix(dir, "testing"):
			break LOOP
		case strings.HasSuffix(dir, "test"):
			fallthrough
		case strings.HasSuffix(dir, "test/utils"):
			fallthrough
		case strings.HasSuffix(dir, "testdata"):
			fallthrough
		case strings.HasSuffix(dir, "runtime"):
			// Do nothing
		default:
			lastPkgDir = dir
		}
	}
	if len(lastPkgDir) == 0 {
		if RuntimeTest != nil {
			frame, _ := runtime.CallersFrames([]uintptr{reflect.ValueOf(RuntimeTest).Pointer()}).Next()
			return filepath.Dir(frame.File)
		}
		panic("unable find package path")
	}
	return lastPkgDir
}

// ProjectDirectory returns the directory of test's project root (directory containing go.mod).
// Note: this function leverage PackageDirectory(), so same limitation applies
func ProjectDirectory() string {
	pkgDir := PackageDirectory()
	for dir := pkgDir; dir != "/" && dir != ""; dir = filepath.Clean(filepath.Dir(dir)) {
		stat, e := os.Stat(filepath.Join(dir, "go.mod"))
		if e == nil && !stat.IsDir() {
			return dir
		}
	}
	return ""
}
