package log

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
)

/*
	Common functions that useful to any logger factory
 */

func loggerKey(name string) string {
	return utils.CamelToSnakeCase(name)
}

func convertLevelsNameToKey(byNames map[string]LoggingLevel) (byKeys map[string]LoggingLevel) {
	byKeys = map[string]LoggingLevel{}
	for k, v := range byNames {
		byKeys[loggerKey(k)] = v
	}
	return
}

func openOrCreateFile(location string) (*os.File, error) {
	if location == "" {
		return nil, fmt.Errorf("location is missing for file logger")
	}
	dir := filepath.Dir(location)
	if e := os.MkdirAll(dir, 0744); e != nil {
		return nil, e
	}
	return os.OpenFile(location, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
}
