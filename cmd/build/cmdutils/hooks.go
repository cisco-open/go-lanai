package cmdutils

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var logger = log.New("Build")

type RunE func(cmd *cobra.Command, args []string) error

func MergeRunE(funcs ...RunE) RunE {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range funcs {
			if f == nil {
				continue
			}
			if e := f(cmd, args); e != nil {
				return e
			}
		}
		return nil
	}
}

func EnsureDir(dirPath *string, base string, cleanIfExist bool, desc string) RunE {
	return func(cmd *cobra.Command, args []string) error {
		if e := ensureDir(dirPath, base, cleanIfExist); e != nil {
			return fmt.Errorf("%s: %v", desc, e)

		}
		return nil
	}
}

func EnsureGlobalDirectories() RunE {
	return func(_ *cobra.Command, _ []string) error {
		if e := ensureDir(&GlobalArgs.WorkingDir, currentDir(), false); e != nil {
			return fmt.Errorf("working directory: %v", e)
		}
		if e := ensureDir(&GlobalArgs.TmpDir, GlobalArgs.WorkingDir, true); e != nil {
			return fmt.Errorf("tmp directory: %v", e)
		}

		if e := ensureDir(&GlobalArgs.OutputDir, GlobalArgs.WorkingDir, false); e != nil {
			return fmt.Errorf("output directory: %v", e)
		}
		return nil
	}
}

func PrintEnvironment() RunE {
	return func(cmd *cobra.Command, _ []string) error {
		logger := logger.WithContext(cmd.Context())
		logger.Debugf("%18s: %s", "Working Directory", GlobalArgs.WorkingDir)
		logger.Debugf("%18s: %s", "Tmp Directory", GlobalArgs.TmpDir)
		logger.Debugf("%18s: %s", "Output Directory", GlobalArgs.OutputDir)
		return nil
	}
}

func ensureDir(path *string, base string, cleanIfExist bool) (err error) {
	if path == nil || *path == "" {
		return fmt.Errorf("not set")
	}

	defer func() {
		if err == nil && cleanIfExist {
			err = removeContents(*path)
		}
	}()

	// if the path is already absolute, we just make sure the dir exists
	if filepath.IsAbs(*path) {
		err = mkdirIfNotExists(*path)
		return
	}

	*path = filepath.Clean(base + "/" + *path)
	err = mkdirIfNotExists(*path)
	return
}

func mkdirIfNotExists(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s is not absolute path", path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if e := os.Mkdir(path, 0744); e != nil {
			return e
		}
	}
	return nil
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() {_ = d.Close()}()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
