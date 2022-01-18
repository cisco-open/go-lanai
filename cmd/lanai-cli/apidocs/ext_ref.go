package apidocs

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

const (
	pathSeparator  = `.`
	refKey         = `$ref`
	refLocalPrefix = `#`
	refKeySuffix   = pathSeparator + refKey
)

var (
	ExtRefSourceReplacement = map[string]string{
		"https://api.swaggerhub.com/domains/Cisco-Systems46/msx-common-domain/8": "github://cto-github.cisco.com/raw/NFV-BU/msx-platform-specs/v1.0.8/common-domain-8.yaml",
	}
)

func tryResolveExtRefs(ctx context.Context, docs []*apidoc) ([]*apidoc, error) {
	seen := map[string]*apidoc{}
	for _, doc := range docs {
		seen[doc.source] = doc
	}

	for i := 0; i < len(docs); i++ {
		extra, e := resolveExtRefs(ctx, docs[i], seen)
		if e != nil {
			return nil, e
		}
		for _, doc := range extra {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

// resolveExtRefs traverse given document, inspect all "$ref" fields and try to load additional external documents
// returns additional documents required
func resolveExtRefs(ctx context.Context, doc *apidoc, seen map[string]*apidoc) (extra []*apidoc, err error) {
	refResolver := func(val reflect.Value, path string, key, parent reflect.Value) bool {
		if !strings.HasSuffix(path, refKeySuffix) || !key.IsValid() || parent.Kind() != reflect.Map {
			return false
		}

		for ; val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr; val = val.Elem() {
		}
		if val.Kind() != reflect.String {
			// we don't handle non-string $ref field, bail and continue
			return false
		}

		ref := strings.TrimSpace(val.String())
		if strings.HasPrefix(ref, refLocalPrefix) {
			// local reference, nothing to do here
			return false
		}

		// resolve
		resolved, loaded, e := doResolveExtRef(ctx, ref)
		if e != nil {
			err = fmt.Errorf("unable to resolve external reference [%s] at [%s]: %v", ref, path, e)
			return true
		}

		// update value and record extra
		parent.SetMapIndex(key, reflect.ValueOf(resolved))
		if loaded != nil {
			if _, ok := seen[loaded.source]; !ok {
				extra = append(extra, loaded)
				seen[loaded.source] = loaded
				logger.WithContext(ctx).Infof("Loaded external document [%s]", loaded.source)
			}
		}
		return false
	}

	// traverse doc and apply ref Resolver
	traverse(reflect.ValueOf(doc.value), "$", reflect.Value{}, reflect.Value{}, refResolver)
	return
}

func doResolveExtRef(ctx context.Context, ref string) (resolved string, loaded *apidoc, err error) {
	split := strings.SplitN(ref, refLocalPrefix, 2)
	if len(split) != 2 {
		// invalid ref, we leave it as-is
		return ref, nil, nil
	}

	// try load additional docs
	src := split[0]
	if replace, ok := ExtRefSourceReplacement[src]; ok && replace != "" {
		src = replace
	}
	doc, e := loadApiDoc(ctx, src)
	if e != nil {
		return "", nil, e
	}
	return refLocalPrefix + split[1], doc, nil
}

type traversalHandler func(val reflect.Value, path string, key, parent reflect.Value) (stop bool)

func traverse(val reflect.Value, path string, key, parent reflect.Value, handler traversalHandler) bool {
	for ; val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr; val = val.Elem() {
	}

	switch val.Kind() {
	case reflect.Map:
		for i := val.MapRange(); i.Next(); {
			k := i.Key()
			p := fmt.Sprintf("%s.%s", path, k.String())
			v := i.Value()
			if stop := traverse(v, p, k, val, handler); stop {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		length := val.Len()
		for i := 0; i < length; i++ {
			p := fmt.Sprintf("%s[%d]", path, i)
			v := val.Index(i)
			if stop := traverse(v, p, reflect.ValueOf(i), val, handler); stop {
				return true
			}
		}
	default:
		if handler != nil {
			return handler(val, path, key, parent)
		}
	}
	return false
}
