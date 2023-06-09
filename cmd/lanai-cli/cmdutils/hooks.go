package cmdutils

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
)

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
		if e := ensureDir(&GlobalArgs.WorkingDir, goModDir(), false); e != nil {
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
		if !GlobalArgs.Verbose {
			return nil
		}

		const tmpl = `%18s: %s`
		ctxLog := logger.WithContext(cmd.Context())
		ctxLog.Debugf(tmpl, "Working Directory", GlobalArgs.WorkingDir)
		ctxLog.Debugf(tmpl, "Tmp Directory", GlobalArgs.TmpDir)
		ctxLog.Debugf(tmpl, "Output Directory", GlobalArgs.OutputDir)
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
