package migration

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
)

func MigrationFuncFromTextFile(filePath string) MigrationFunc {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) || !info.IsDir() {
		//TODO: error
	}

	file, _ := os.Open(filePath);
	cql, err := ioutil.ReadAll(file)

	return func(ctx context.Context) error {
		for _, query := range strings.Split(string(cql), ";") {
			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}
			//TODO: execute the migration using the actual db driver
			logger.Infof("executing query %s", query)
		}
		return nil
	}
}