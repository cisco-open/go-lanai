package migration

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"io/fs"
	"strings"
)

func migrationFuncFromTextFile(fs fs.FS, filePath string, db *gorm.DB) (MigrationFunc){
	file, err := fs.Open(filePath)
	if err != nil {
		panic(errors.New(fmt.Sprintf("%s does not exist or is not a file", filePath)))
	}

	sql, err := io.ReadAll(file)

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