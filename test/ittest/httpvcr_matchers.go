package ittest

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type RecordMatcherOptions func(opt *RecordMatcherOption)
type RecordMatcherOption struct {
	URLMatcher    RecordURLMatcherFunc
	QueryMatcher  RecordQueryMatcherFunc
	HeaderMatcher RecordHeaderMatcherFunc
	BodyMatcher   RecordBodyMatcherFunc
}

// NewRecordMatcher create a custom RecordMatcherFunc to compare recorded request and actual request.
// By default, the crated matcher compare following:
// - Method, Host, Path are exact match
// - Queries are exact match except for SensitiveRequestQueries
// - Headers are exact match except for SensitiveRequestHeaders
// - Body is compared as JSON
// Note: directly using this function requires knowledge about golang generics and function casting.
func NewRecordMatcher(opts ...RecordMatcherOptions) GenericMatcherFunc[*http.Request, cassette.Request] {
	opt := &RecordMatcherOption{
		URLMatcher:    RecordURLMatcherFunc(NewRecordURLMatcher()),
		QueryMatcher:  RecordQueryMatcherFunc(NewRecordQueryMatcher(SensitiveRequestQueries.Values()...)),
		HeaderMatcher: RecordHeaderMatcherFunc(NewRecordHeaderMatcher(SensitiveRequestHeaders.Values()...)),
		BodyMatcher:   RecordBodyMatcherFunc(NewRecordJsonBodyMatcher()),
	}
	for _, fn := range opts {
		fn(opt)
	}
	return func(out *http.Request, record cassette.Request) error {
		recordUrl, e := url.Parse(record.URL)
		if e != nil {
			return fmt.Errorf("invalid recorded URL")
		}

		if out.Method != record.Method {
			return fmt.Errorf("http method mismatch")
		}
		if e := opt.URLMatcher(out.URL, recordUrl); e != nil {
			return e
		}
		if e := opt.QueryMatcher(out.URL.Query(), recordUrl.Query()); e != nil {
			return e
		}
		if e := opt.HeaderMatcher(out.Header, record.Headers); e != nil {
			return e
		}

		if opt.BodyMatcher == nil {
			return nil
		}

		var reqBody []byte
		if out.Body != nil && out.Body != http.NoBody {
			data, e := io.ReadAll(out.Body)
			if e != nil {
				return fmt.Errorf("unable to read request's body")
			}
			reqBody = data
			out.Body.Close()
			out.Body = io.NopCloser(bytes.NewBuffer(data))
		}
		if e := opt.BodyMatcher(reqBody, []byte(record.Body)); e != nil {
			return e
		}

		return nil
	}
}

// NewRecordIndexAwareMatcher is a special matcher that ensure requests are executed in the recorded order
// Note 1: because current httpvcr lib doesn't expose the interaction ID, we stored it in header
// 		   using InteractionIndexAwareHook
// Note 2: the next expected ID would increase regardless if ID matches. So this index should be used together with
// 		   other matchers: "other_matcher AND NewRecordIndexAwareMatcher()". This matcher need to be put behind
// 		   any other matchers
func NewRecordIndexAwareMatcher() GenericMatcherFunc[*http.Request, cassette.Request] {
	var id int
	return func(out *http.Request, record cassette.Request) error {
		recordId, e := strconv.Atoi(record.Headers.Get(xInteractionIndexHeader))
		if e != nil {
			return nil
		}
		defer func() {id++}()
		if id != recordId {
			return fmt.Errorf("HTTP interaction ID doesn't match")
		}

		return nil
	}
}

// NewRecordHostIgnoringURLMatcher returns RecordURLMatcherFunc that compares Method, Path
func NewRecordHostIgnoringURLMatcher() GenericMatcherFunc[*url.URL, *url.URL] {
	return func(out *url.URL, record *url.URL) error {
		if out.Path != record.Path {
			return fmt.Errorf("http path mismatch")
		}
		return nil
	}
}

// NewRecordURLMatcher returns RecordURLMatcherFunc that compares Method, Path, Host and Port
func NewRecordURLMatcher() GenericMatcherFunc[*url.URL, *url.URL] {
	return func(out *url.URL, record *url.URL) error {
		if out.Host != record.Host {
			return fmt.Errorf("http host mismatch")
		}
		if out.Path != record.Path {
			return fmt.Errorf("http path mismatch")
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

// TODO
func NewRecordJsonBodyMatcher(gjsonPaths ...string) GenericMatcherFunc[[]byte, []byte] {
	return func(out []byte, record []byte) error {
		//lroot, e := ajson.Unmarshal(out)
		//if e != nil {
		//	return e
		//}
		//rroot, e := ajson.Unmarshal(record)
		//if e != nil {
		//	return e
		//}
		//lroot.
		return nil
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
