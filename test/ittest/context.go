package ittest

import (
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"net/url"
)

type Mode recorder.Mode

// Recorder states
const (
	// ModeCommandline lets the commandline or the state in TestMain to determine the mode. Default to ModeReplaying
	ModeCommandline Mode = iota
	ModeReplaying
	ModeRecording
)

/*************************
	HttpVCROptions
 *************************/

type HttpVCROptions func(opt *HttpVCROption)
type HttpVCROption struct {
	Name           string
	Mode           Mode
	SavePath       string
	RecordMatching []RecordMatcherOptions
	Hooks          []recorder.Hook
}

/******************************
	HTTP VCR Request Matching
 ******************************/

type GenericMatcherFunc[O, R any] func(O, R) error

type RecordMatcherFunc GenericMatcherFunc[*http.Request, cassette.Request]
type RecordURLMatcherFunc GenericMatcherFunc[*url.URL, *url.URL]
type RecordQueryMatcherFunc GenericMatcherFunc[url.Values, url.Values]
type RecordHeaderMatcherFunc GenericMatcherFunc[http.Header, http.Header]
type RecordBodyMatcherFunc GenericMatcherFunc[[]byte, []byte]

/******************************
	Request Matcher Logic Ops
 ******************************/

// AndMatcher generic AND operator of given matchers
// Note: because golang generics requires instantiation, type casting is required.
// 		 e.g. var m RecordBodyMatcherFunc = RecordBodyMatcherFunc(AndMatcher(matcher1, matcher2))
func AndMatcher[O, R any](matchers ...GenericMatcherFunc[O, R]) GenericMatcherFunc[O, R] {
	return func(out O, record R) error {
		for _, matcher := range matchers {
			if e := matcher(out, record); e != nil {
				return e
			}
		}
		return nil
	}
}

// OrMatcher generic OR operator of given matchers
// Note: because golang generics requires instantiation, type casting is required.
// 		 e.g. var m RecordBodyMatcherFunc = RecordBodyMatcherFunc(AndMatcher(matcher1, matcher2))
func OrMatcher[O, R any](matchers ...GenericMatcherFunc[O, R]) GenericMatcherFunc[O, R] {
	return func(out O, record R) error {
		var e error
		for _, matcher := range matchers {
			if e = matcher(out, record); e == nil {
				return nil
			}
		}
		return e
	}
}
