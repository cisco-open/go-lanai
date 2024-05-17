package errorhandling

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/data/repo"
	"gorm.io/gorm"
	"strings"
)

// TranslateDataIntegrityError used for translating data integrity error (especially duplicate keys error)
// to an error with sensible message
func TranslateDataIntegrityError(ctx context.Context, err error) error {
	//nolint:errorlint // intended, we ignore wrapped error
	de, ok := err.(data.DataError)
	if !ok || !errors.Is(err, data.ErrorCategoryData) {
		return err
	}

	var msg string
	switch {
	case errors.Is(err, data.ErrorDuplicateKey):
		msg = translateDuplicateKeyMessage(ctx, de)
	default:
		return err
	}
	return de.WithMessage(msg)
}

func translateDuplicateKeyMessage(ctx context.Context, err data.DataError) string {
	// resolve duplicated fields
	model := extractModel(ctx, err)
	dups, e := repo.Utils().CheckUniqueness(ctx, model)
	switch {
	case errors.Is(e, data.ErrorDuplicateKey):
	default:
		return err.Error()
	}
	pairs := make([]string, 0, len(dups))
	for k, v := range dups {
		pairs = append(pairs, fmt.Sprintf(`%s: "%v"`, k, v))
	}

	// resolve model name
	resolver, e := repo.Utils().ResolveSchema(ctx, model)
	if e != nil {
		return e.Error()
	}

	return fmt.Sprintf("%s already exists with [%s]", resolver.ModelName(), strings.Join(pairs, ", "))
}

func extractModel(_ context.Context, err data.DataError) interface{} {
	switch details := err.Details().(type) {
	case *gorm.Statement:
		return details.Model
	default:
		return nil
	}
}
