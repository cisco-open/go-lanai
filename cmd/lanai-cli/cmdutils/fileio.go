package cmdutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
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

// BindYamlFile find, read and bind YAML file, returns absolute path of loaded file
func BindYamlFile(bind interface{}, filepath string, additionalLookupDirs ...string) (string, error) {
	absPath, file, e := LookupFile(filepath, additionalLookupDirs...)
	if e != nil {
		return "", e
	}
	defer func() {_ = file.Close()}()

	// read and parse file
	encoded, e := ioutil.ReadAll(file)
	if e != nil {
		return "", fmt.Errorf("unable to read file %s: %v", absPath, e)
	}

	jsonData, e := yaml.YAMLToJSON(encoded)
	if e != nil {
		return "", fmt.Errorf("unable to parse YAML file %s: %v", absPath, e)
	}

	if e := json.Unmarshal(jsonData, bind); e != nil {
		return "", fmt.Errorf("unable to bind file %s to %T: %v", absPath, bind, e)
	}
	return absPath, nil
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
		absPath, err = filepath.Abs(path.Join(dir, filePath))
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

