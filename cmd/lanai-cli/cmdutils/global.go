package cmdutils

import (
	"os"
	"path"
	"path/filepath"
)

var (
	GlobalArgs = Global{
		WorkingDir: DefaultWorkingDir(),
		TmpDir:     DefaultTemporaryDir(),
		OutputDir:  DefaultOutputDir(),
	}
)

type Global struct {
	WorkingDir string `flag:"workspace,w" desc:"Working directory containing 'go.mod'. All non-absolute paths are relative to this directory"`
	TmpDir     string `flag:"tmp-dir" desc:"Temporary directory."`
	OutputDir  string `flag:"output,o" desc:"Output directory. All non-absolute paths for output are relative to this directory"`
}

func (g Global) AbsPath(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(base + "/" + path)
}

func DefaultWorkingDir() string {
	return goModDir()
}

func DefaultTemporaryDir() string {
	const relative = ".tmp"
	return PathRelativeToModuleDir(relative)
}

func DefaultOutputDir() string {
	const relative = "dist"
	return PathRelativeToModuleDir(relative)
}

func PathRelativeToModuleDir(relativePath string) string {
	if path.IsAbs(relativePath) {
		panic("PathRelativeToModuleDir only takes relative path")
	}
	base := goModDir()
	if base == "" {
		return DefaultWorkingDir() + "/" + relativePath
	}
	return base + "/" + relativePath
}

func currentDir() string {
	currDir, e := os.Getwd()
	if e != nil {
		panic(e)
	}
	return currDir
}

// goModDir works from current directory backward along the FS tree until find go.mod file
func goModDir() string {
	currDir, e := os.Getwd()
	if e != nil {
		panic(e)
	}

	for dir := currDir; dir != "" && dir != "."; dir = filepath.Dir(dir) {
		gomodPath := dir + "/go.mod"
		if _, e := os.Stat(gomodPath); e == nil {
			return dir
		}
	}
	return currentDir()
}