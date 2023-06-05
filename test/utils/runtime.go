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
