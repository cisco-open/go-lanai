package ittest

import (
	"bytes"
	"fmt"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var errInteractionIDMismatch = fmt.Errorf("HTTP interaction ID doesn't match")

type RecordMatcherOptions func(opt *RecordMatcherOption)
type RecordMatcherOption struct {
	// Convenient Options
	IgnoreHost     bool
	FuzzyHeaders   []string
	FuzzyQueries   []string
	FuzzyPostForm  []string
	FuzzyJsonPaths []string

	// Advanced Options, if set, will overwrite corresponding convenient options
	// Note: directly changing these defaults requires knowledge about golang generics and function casting.
	URLMatcher    RecordURLMatcherFunc
	QueryMatcher  RecordQueryMatcherFunc
	HeaderMatcher RecordHeaderMatcherFunc
	BodyMatchers  []RecordBodyMatcher
}

// NewRecordMatcher create a custom RecordMatcherFunc to compare recorded request and actual request.
// By default, the crated matcher compare following:
// - Method, Host, Path are exact match
// - Queries are exact match except for FuzzyRequestQueries
// - Headers are exact match except for FuzzyRequestHeaders
// - Body is compared as JSON or x-www-form-urlencoded Form
//
// Note: In case the request contains random/temporal data in queries/headers/form/JSON, use Fuzzy* options
func NewRecordMatcher(opts ...RecordMatcherOptions) GenericMatcherFunc[*http.Request, cassette.Request] {
	opt := resolveMatcherOption(opts)
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

func resolveMatcherOption(opts []RecordMatcherOptions) *RecordMatcherOption {
	opt := RecordMatcherOption{
		IgnoreHost:    false,
		FuzzyHeaders:  FuzzyRequestHeaders.Values(),
		FuzzyQueries:  FuzzyRequestQueries.Values(),
		FuzzyPostForm: FuzzyRequestQueries.Values(),
	}
	for _, fn := range opts {
		fn(&opt)
	}

	if opt.URLMatcher == nil {
		opt.URLMatcher = RecordURLMatcherFunc(NewRecordURLMatcher(opt.IgnoreHost))
	}

	if opt.QueryMatcher == nil {
		opt.QueryMatcher = RecordQueryMatcherFunc(NewRecordQueryMatcher(opt.FuzzyQueries...))
	}

	if opt.HeaderMatcher == nil {
		opt.HeaderMatcher = RecordHeaderMatcherFunc(NewRecordHeaderMatcher(opt.FuzzyHeaders...))
	}

	opt.BodyMatchers = append(opt.BodyMatchers,
		NewRecordJsonBodyMatcher(opt.FuzzyJsonPaths...),
		NewRecordFormBodyMatcher(opt.FuzzyPostForm...),
		NewRecordLiteralBodyMatcher(),
	)

	return &opt
}

// indexAwareMatcherWrapper is a special matcher wrapper that ensure requests are executed in the recorded order
type indexAwareMatcherWrapper struct {
	// count for total actual request have seen
	count int
}

func newIndexAwareMatcherWrapper() *indexAwareMatcherWrapper {
	return &indexAwareMatcherWrapper{
		count: 0,
	}
}

// MatcherFunc wrap given delegate with index enforcement
// Note 1: because current httpvcr lib doesn't expose the interaction ID, we stored it in header
//
//	using InteractionIndexAwareHook
//
// Note 2: This wrapper doesn't invoke delegate if expected ID doesn't match.
// Note 3: The next expected ID would increase if delegate is a match. This means if recorder couldn't match the
//
//	request with currently expected interaction, it would keep waiting on the same interaction
func (w *indexAwareMatcherWrapper) MatcherFunc(delegate RecordMatcherFunc) GenericMatcherFunc[*http.Request, cassette.Request] {
	return func(out *http.Request, record cassette.Request) error {
		recordId, e := strconv.Atoi(record.Headers.Get(xInteractionIndexHeader))
		if e != nil {
			recordId = -1
		}

		seen := len(out.Header.Get(xInteractionSeenHeader)) != 0
		if !seen {
			// a new request, we adjust the expectation and set the request to be seen
			out.Header.Set(xInteractionSeenHeader, "true")
			w.count++
		}

		// do interaction match first
		if w.count != recordId+1 {
			return errInteractionIDMismatch
		}

		// invoke delegate, increase counter if applicable
		return delegate(out, record)
	}
}

/*********************
	Matcher Options
 *********************/

// IgnoreHost returns RecordMatcherOptions that ignore host during record matching
func IgnoreHost() RecordMatcherOptions {
	return func(opt *RecordMatcherOption) {
		opt.IgnoreHost = true
	}
}

// FuzzyHeaders returns RecordMatcherOptions that ignore header values of given names during record matching
// Note: still check if the header exists, only value comparison is skipped
func FuzzyHeaders(headers ...string) RecordMatcherOptions {
	return func(opt *RecordMatcherOption) {
		opt.FuzzyHeaders = append(opt.FuzzyHeaders, headers...)
	}
}

// FuzzyQueries returns RecordMatcherOptions that ignore query value of given keys during record matching
// Note: still check if the value exists, only value comparison is skipped.
// This function dosen't consider POST form data. Use FuzzyForm for both Queries and POST form data
func FuzzyQueries(queries ...string) RecordMatcherOptions {
	return FuzzyForm(queries...)
}

// FuzzyForm returns RecordMatcherOptions that ignore form values (in queries and post body if applicable) of given keys during record matching
// Note: still check if the value exists, only value comparison is skipped
func FuzzyForm(formKeys ...string) RecordMatcherOptions {
	return func(opt *RecordMatcherOption) {
		opt.FuzzyQueries = append(opt.FuzzyQueries, formKeys...)
		opt.FuzzyPostForm = append(opt.FuzzyPostForm, formKeys...)
	}
}

// FuzzyJsonPaths returns RecordMatcherOptions that ignore fields in JSON body that matching the given JSONPaths
// JSONPath Syntax: https://goessner.net/articles/JsonPath/
func FuzzyJsonPaths(jsonPaths ...string) RecordMatcherOptions {
	return func(opt *RecordMatcherOption) {
		opt.FuzzyJsonPaths = append(opt.FuzzyJsonPaths, jsonPaths...)
	}
}
