package apidocs

import (
	"context"
	"fmt"
	"dario.cat/mergo"
)

const (
	kOAS3Servers = `servers`
)

func merge(_ context.Context, docs []*apidoc) (*apidoc, error) {
	merged := make(map[string]interface{})
	for _, doc := range docs {
		if e := mergo.Merge(&merged, doc.value, mergo.WithAppendSlice); e != nil {
			return nil, fmt.Errorf("failed to merge [%s]: %v", doc.source, e)
		}
	}

	return &apidoc{
		source: ResolveArgs.Output,
		value:  postMerge(merged),
	}, nil
}

func postMerge(doc map[string]interface{}) map[string]interface{} {
	delete(doc, kOAS3Servers)
	return doc
}

func writeMergedToFile(ctx context.Context, doc *apidoc) error {
	doc.source = ResolveArgs.Output
	absPath, e := writeApiDocLocal(ctx, doc)
	if e != nil {
		return e
	}
	logger.Infof("Merged API document saved to [%s]", absPath)
	return nil
}
