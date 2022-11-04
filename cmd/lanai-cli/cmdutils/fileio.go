package cmdutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/ghodss/yaml"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// OpenFile open file relative to GlobalArgs.WorkingDir.
// Returns absolute path and opened file handle if successful.
// Otherwise, return error
func OpenFile(filePath string, flag int, perm os.FileMode) (absPath string, file *os.File, err error) {
	return lookupFile(filePath, flag, perm)
}

// LookupFile search and open file for read, relative to GlobalArgs.WorkingDir or any additional lookup directories.
// Returns absolute path and opened file handle if successful.
// Otherwise, return error
func LookupFile(filePath string, additionalLookupDirs ...string) (absPath string, file *os.File, err error) {
	return lookupFile(filePath, os.O_RDONLY, 0, additionalLookupDirs...)
}

// LookupFiles search files using provided pattern under given lookup directories relative to GlobalArgs.WorkingDir.
// Returns list of absolute path if successful.
// Otherwise, return error
func LookupFiles(pattern string, dirs ...string) (absPaths []string, err error) {
	for i, dir := range dirs {
		dirs[i] = GlobalArgs.WorkingDir + "/" + dir
	}
	return lookupFiles(pattern, dirs)
}

// BindYamlFile find, read and bind YAML file, returns absolute path of loaded file
func BindYamlFile(bind interface{}, filepath string, additionalLookupDirs ...string) (string, error) {
	absPath, file, e := LookupFile(filepath, additionalLookupDirs...)
	if e != nil {
		return "", e
	}
	defer func() {_ = file.Close()}()

	// read and parse file
	if e := BindYaml(file, bind); e != nil {
		return "", fmt.Errorf("unable to parse YAML file %s: %v", absPath, e)
	}
	return absPath, nil
}

// BindYaml read from given io.Reader and parse as YAML
func BindYaml(reader io.Reader, bind interface{}) error {
	// read and parse file
	encoded, e := io.ReadAll(reader)
	if e != nil {
		return e
	}

	if e := yaml.Unmarshal(encoded, bind); e != nil {
		return e
	}
	return nil
}

// BindJsonFile find, read and bind JSON file, returns absolute path of loaded file
func BindJsonFile(bind interface{}, filepath string, additionalLookupDirs ...string) (string, error) {
	absPath, file, e := LookupFile(filepath, additionalLookupDirs...)
	if e != nil {
		return "", e
	}
	defer func() {_ = file.Close()}()

	decoder := json.NewDecoder(file)
	if e := decoder.Decode(bind); e != nil {
		return "", fmt.Errorf("unable to bind file %s to %T: %v", absPath, bind, e)
	}
	return absPath, nil
}

func lookupFile(filePath string, flag int, perm os.FileMode, additionalLookupDirs ...string) (absPath string, file *os.File, err error) {
	// look up the file
	lookup := append([]string{GlobalArgs.WorkingDir}, additionalLookupDirs...)
	for _, dir := range lookup {
		if filepath.IsAbs(filePath) {
			absPath = filePath
		} else {
			absPath, err = filepath.Abs(path.Join(dir, filePath))
		}
		if err == nil {
			// open file
			f, e := os.OpenFile(absPath, flag, perm)
			if e != nil {
				return "", nil, e
			}
			file = f
			break
		}
	}
	if file == nil {
		absPath = ""
		err = fmt.Errorf("unable to find file %s in directories %s", filePath, strings.Join(lookup, ":"))
	}
	return
}

func lookupFiles(pattern string, lookup []string) (paths []string, err error) {
	// look up the file
	for _, dir := range lookup {
		stat, e := os.Stat(dir)
		if e != nil {
			return nil, e
		}
		files := searchFiles(dir, stat, pattern)
		paths = append(paths, files...)
	}
	return
}

// searchFiles recursively search for all files in given path if it's a directory
func searchFiles(path string, stat fs.FileInfo, pattern string) []string {
	if !stat.IsDir() {
		if match, e := doublestar.Match(pattern, path); e == nil && match {
			return []string{path}
		} else {
			return nil
		}
	}

	entries, e := os.ReadDir(path)
	if e != nil {
		return nil
	}
	expanded := make([]string, 0)
	for _, entry := range entries {
		if info, e := entry.Info(); e == nil {
			subPath := filepath.Join(path, info.Name())
			sub := searchFiles(subPath, info, pattern)
			expanded = append(expanded, sub...)
		}
	}
	return expanded
}

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

