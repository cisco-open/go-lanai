package apidocs

import "context"

// resolveExtRef traverse given document, inspect all "$ref" fields and try to load additional external documents
// returns additional documents required
func resolveExtRef(ctx context.Context, doc *apidoc) (extra []*apidoc, err error) {
	//TODO implement this
	return nil, nil
}
