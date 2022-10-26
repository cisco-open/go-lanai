package ittest

import (
	"bytes"
	"fmt"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"io"
	"net/http"
	"net/url"
)

type RecordMatcherOptions func(opt *RecordMatcherOption)
type RecordMatcherOption struct {
	URLMatcher    RecordURLMatcherFunc
	QueryMatcher  RecordQueryMatcherFunc
	HeaderMatcher RecordHeaderMatcherFunc
	BodyMatchers   []RecordBodyMatcher
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
		BodyMatchers:   []RecordBodyMatcher{
			//NewRecordJsonBodyMatcher("$.time"),
			NewRecordJsonBodyMatcher(),
			//NewRecordFormBodyMatcher("time"),
			NewRecordFormBodyMatcher(),
			NewRecordLiteralBodyMatcher(),
		},
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

		if len(opt.BodyMatchers) == 0 {
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
		// find first supported body matcher and use it
		contentType := record.Headers.Get("Content-Type")
		for _, matcher := range opt.BodyMatchers {
			if !matcher.Support(contentType) {
				continue
			}
			return matcher.Matches(reqBody, []byte(record.Body))
		}
		return nil
	}
}
