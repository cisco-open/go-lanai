// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package apidocs

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
)

const (
	pathSeparator       = `.`
	refKey              = `$ref`
	refLocalPrefix      = `#`
	refKeySuffix        = pathSeparator + refKey
	replaceArgSeparator = `=>`
)

var (
	extSourceReplace map[string]string
	extSourceLookup  = map[string]string{}
)

func tryResolveExtRefs(ctx context.Context, docs []*apidoc) ([]*apidoc, error) {
	if e := populateExtSourceReplace(); e != nil {
		return nil, e
	}

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
			// SuppressWarnings go:S108 empty block is intended
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
	src := altExtSource(split[0])
	if len(src) != 0 {
		doc, e := loadApiDoc(ctx, src)
		if e != nil {
			return "", nil, e
		}
		return refLocalPrefix + split[1], doc, nil
	}
	return refLocalPrefix + split[1], nil, nil
}

type traversalHandler func(val reflect.Value, path string, key, parent reflect.Value) (stop bool)

func traverse(val reflect.Value, path string, key, parent reflect.Value, handler traversalHandler) bool {
	for ; val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr; val = val.Elem() {
		// SuppressWarnings go:S108 empty block is intended
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

func altExtSource(url string) (alt string) {
	if result, ok := extSourceLookup[url]; ok {
		return result
	}

	defer func() {
		extSourceLookup[url] = alt
	}()

	k, e := normalizeURL(url)
	if e != nil {
		return url
	}

	if result, ok := extSourceReplace[k]; ok {
		return result
	}
	return url
}

func populateExtSourceReplace() error {
	extSourceReplace = map[string]string{}
	// parse from ResolveConfig
	for _, v := range ResolveConf.ReplaceExtSources {
		key, e := normalizeURL(v.Url)
		if e != nil {
			return fmt.Errorf(`invalid external source [%s]: %v`, v.Url, e)
		}
		extSourceReplace[key] = os.ExpandEnv(v.To)
	}

	// parse from ResolveArguments
	for _, pair := range ResolveArgs.ReplaceExtSources {
		split := strings.SplitN(pair, replaceArgSeparator, 2)
		if len(split) != 2 {
			return fmt.Errorf(`invalid external source replacement value. Expect "<original>=><replacement>"`)
		}
		k, e := normalizeURL(split[0])
		if e != nil {
			return fmt.Errorf(`invalid external source [%s]: %v`, split[0], e)
		}
		extSourceReplace[k] = os.ExpandEnv(split[1])

	}
	return nil
}

func normalizeURL(val string) (string, error) {
	parsed, e := url.Parse(os.ExpandEnv(val))
	if e != nil {
		return "", e
	}
	return parsed.String(), nil
}
