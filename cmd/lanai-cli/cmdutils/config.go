package cmdutils

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func LoadYamlConfig(bind interface{}, filepath string, additionalLookupDirs ...string) error {
	fullpath, file, e := LookupFile(filepath, additionalLookupDirs...)
	if e != nil {
		return e
	}
	defer func() {_ = file.Close()}()

	// read and parse file
	encoded, e := ioutil.ReadAll(file)
	if e != nil {
		return fmt.Errorf("unable to read file %s: %v", fullpath, e)
	}

	jsonData, e := yaml.YAMLToJSON(encoded)
	if e != nil {
		return fmt.Errorf("unable to parse YAML file %s: %v", fullpath, e)
	}

	if e := json.Unmarshal(jsonData, bind); e != nil {
		return fmt.Errorf("unable to bind file %s to %T: %v", fullpath, bind, e)
	}
	return nil
}

func LookupFile(filepath string, additionalLookupDirs ...string) (fullpath string, file *os.File, err error) {
	// look up the file
	lookup := append([]string{GlobalArgs.WorkingDir}, additionalLookupDirs...)
	for _, dir := range lookup {
		fullpath = path.Join(dir, filepath)
		if isFileExists(fullpath) {
			// open file
			f, e := os.Open(fullpath)
			if e != nil {
				return "", nil, fmt.Errorf("unable to open file %s: %v", fullpath, e)
			}
			file = f
			break
		}
	}
	if file == nil {
		err = fmt.Errorf("unable to find file %s in directories %s", filepath, strings.Join(lookup, ":"))
	}
	return
}
