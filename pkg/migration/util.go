package migration

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"strings"
)

func migrationFuncFromTextFile(filePath string, db *gorm.DB) (MigrationFunc){
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) || info.IsDir() {
		panic(errors.New(fmt.Sprintf("%s does not exist or is not a file", filePath)))
	}

	file, _ := os.Open(filePath);
	sql, err := ioutil.ReadAll(file)

	if err != nil {
		panic(err)
	}

	return func(ctx context.Context) error {
		for _, query := range strings.Split(string(sql), ";") {
			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}
			logger.Debugf("executing query %s", query)
			result := db.WithContext(ctx).Exec(query)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	}
}