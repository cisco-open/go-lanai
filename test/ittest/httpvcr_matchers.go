package ittest

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"io"
	"net/http"
	"net/url"
)

var (
	SensitiveHeaders = []string{"Authorization"}
	SensitiveQueries = []string{"password", "secret"}
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
// - Queries are exact match except for SensitiveQueries
// - Headers are exact match except for SensitiveHeaders
// - Body is compared as JSON
// Note: directly using this function requires knowledge about golang generics and function casting.
func NewRecordMatcher(opts...RecordMatcherOptions) GenericMatcherFunc[*http.Request, cassette.Request] {
	opt := &RecordMatcherOption{
		URLMatcher:    RecordURLMatcherFunc(NewRecordURLMatcher()),
		QueryMatcher:  RecordQueryMatcherFunc(NewRecordQueryMatcher(SensitiveQueries...)),
		HeaderMatcher: RecordHeaderMatcherFunc(NewRecordHeaderMatcher(SensitiveHeaders...)),
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
	fuzzyK := utils.NewStringSet(fuzzyKeys...)
	return func(out url.Values, record url.Values) error {
		for k, rv := range record {
			exactV := !fuzzyK.Has(k)
			ov, ok := out[k]
			if !ok || exactV && len(ov) != len(rv) {
				return fmt.Errorf("http query [%s] missing", k)
			}
			if !exactV {
				continue
			}
			// values
			for i, v := range ov {
				if rv[i] != v {
					return fmt.Errorf("http query [%s] mismatch", k)
				}
			}
		}
		return nil
	}
}

// NewRecordHeaderMatcher returns RecordHeaderMatcherFunc that compare keys and values of recorded and actual queries
// Any header value is ignored if its key is in the optional fuzzyKeys
func NewRecordHeaderMatcher(fuzzyKeys ...string) GenericMatcherFunc[http.Header, http.Header] {
	delegate := NewRecordQueryMatcher(fuzzyKeys...)
	return func(out http.Header, record http.Header) error {
		return delegate(url.Values(out), url.Values(record))
	}
}

// TODO
func NewRecordJsonBodyMatcher(gjsonPaths ...string) GenericMatcherFunc[[]byte, []byte] {
	return func(out []byte, record []byte) error {
		return nil
	}
}

