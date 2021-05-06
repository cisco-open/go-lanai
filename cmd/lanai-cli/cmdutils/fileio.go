package cmdutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func isFileExists(filepath string) bool {
	info, e := os.Stat(filepath)
	return !os.IsNotExist(e) && !info.IsDir()
}

func mkdirIfNotExists(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s is not absolute path", path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if e := os.MkdirAll(path, 0744); e != nil {
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

func copyFiles(ctx context.Context, files map[string]string) error {
	opts := []ShCmdOptions{
		ShellShowCmd(true),
		ShellUseWorkingDir(),
	}
	for src, dst := range files {
		opts = append(opts, ShellCmd(fmt.Sprintf("cp -r %s %s", src, dst)) )
	}
	opts = append(opts, ShellStdOut(os.Stdout))
	_, e := RunShellCommands(ctx, opts...)
	return e
}

