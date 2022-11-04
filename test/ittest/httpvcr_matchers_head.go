package ittest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"net/http"
	"net/url"
)

// NewRecordURLMatcher returns RecordURLMatcherFunc that compares Method, Path, Host and Port
func NewRecordURLMatcher(ignoreHost bool) GenericMatcherFunc[*url.URL, *url.URL] {
	return func(out *url.URL, record *url.URL) error {
		if out.Path != record.Path {
			return fmt.Errorf("http path mismatch")
		}
		if !ignoreHost && out.Host != record.Host {
			return fmt.Errorf("http host mismatch")
		}
		return nil
	}
}

// NewRecordQueryMatcher returns RecordQueryMatcherFunc that compare keys and values of recorded and actual queries
// Any query value is ignored if its key is in the optional fuzzyKeys
func NewRecordQueryMatcher(fuzzyKeys ...string) GenericMatcherFunc[url.Values, url.Values] {
	return newValuesMatcher("query", nil, fuzzyKeys...)
}

// NewRecordHeaderMatcher returns RecordHeaderMatcherFunc that compare keys and values of recorded and actual queries
// Any header value is ignored if its key is in the optional fuzzyKeys
func NewRecordHeaderMatcher(fuzzyKeys ...string) GenericMatcherFunc[http.Header, http.Header] {
	delegate := newValuesMatcher("header", IgnoredRequestHeaders, fuzzyKeys...)
	return func(out http.Header, record http.Header) error {
		return delegate(url.Values(out), url.Values(record))
	}
}

/**********************
	helpers
 **********************/

// newValuesMatcher returns GenericMatcherFunc[url.Values, url.Values] that compare keys and values of given url.Values
// Any value is ignored if its key is in the optional fuzzyKeys
func newValuesMatcher(name string, ignoredKeys utils.StringSet, fuzzyKeys ...string) GenericMatcherFunc[url.Values, url.Values] {
	fuzzyK := utils.NewStringSet(fuzzyKeys...)
	return func(out url.Values, record url.Values) error {
		for k, rv := range record {
			if ignoredKeys != nil && ignoredKeys.Has(k) {
				continue
			}

			exactV := !fuzzyK.Has(k)
			ov, ok := out[k]
			if !ok || exactV && len(ov) != len(rv) {
				return fmt.Errorf("http %s [%s] missing", name, k)
			}
			if !exactV {
				continue
			}
			// values
			for i, v := range ov {
				if rv[i] != v {
					return fmt.Errorf("http %s [%s] mismatch", name, k)
				}
			}
		}
		return nil
	}
}

