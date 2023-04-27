package testutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PackageDirectory returns package path of the test's Test...() function.
// This function works as following:
// 		It traces back the call stack using runtime, ignoring test packages like "test", "testdata" and "runtime,
// 		until it find golang's "testing" package. The last seen package is considered the package of Test...().
// Note: This function would only work if this is invoked in the same goroutine of the test itself.
func PackageDirectory() string {
	rpc := make([]uintptr, 10)
	if n := runtime.Callers(1, rpc[:]); n < 1 {
		panic("unable find package path")
	}

	var lastPkgDir string
	frames := runtime.CallersFrames(rpc)
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		dir := filepath.Dir(frame.File)
		if strings.HasSuffix(dir, "testing") {
			break
		}
		if !strings.HasSuffix(dir, "test") && !strings.HasSuffix(dir, "testdata") && !strings.HasSuffix(dir, "runtime") {
			lastPkgDir = dir
		}
	}
	if len(lastPkgDir) == 0 {
		// TODO panic instead of give wrong result
		_, file, _, _ := runtime.Caller(0)
		return filepath.Dir(file)
		//panic("unable find package path")
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
