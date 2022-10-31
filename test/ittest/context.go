package ittest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"net/url"
	"time"
)

type Mode int

const CLIRecordModeFlag = "record-http"

// Recorder states
const (
	// ModeCommandline lets the commandline or the state in TestMain to determine the mode. Default to ModeReplaying
	ModeCommandline Mode = iota
	ModeReplaying
	ModeRecording
)

// DefaultHTTPDuration default duration of recorded HTTP interaction
const DefaultHTTPDuration = 200 * time.Microsecond

var (
	xInteractionIndexHeader = `X-Http-Record-Index`
	xInteractionSeenHeader  = `X-Http-Request-Seen`

	IgnoredRequestHeaders = utils.NewStringSet(xInteractionIndexHeader)

	FuzzyRequestHeaders    = utils.NewStringSet("Authorization")
	FuzzyRequestQueries    = utils.NewStringSet("password", "secret", "nonce", "token", "access_token")
	FuzzyRequestJsonPaths  = utils.NewStringSet()
	FuzzyResponseHeaders   = utils.NewStringSet("Date")
	FuzzyResponseJsonPaths = utils.NewStringSet("$..access_token")
)

/*************************
	HTTPVCROptions
 *************************/

type HTTPVCROptions func(opt *HTTPVCROption)
type HTTPVCROption struct {
	Name           string
	Mode           Mode
	SavePath       string
	RecordMatching []RecordMatcherOptions
	Hooks          []RecorderHook
	// special record matcher that enforce interaction order.
	// to change, use DisableHttpRecordOrdering
	indexAwareWrapper *indexAwareMatcherWrapper
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
type RecordBodyMatcher interface {
	Support(contentType string) bool
	Matches(out []byte, record []byte) error
}

/******************************
	HTTP VCR Hooks
 ******************************/

// RecorderHook wrapper of recorder.Hook
type RecorderHook interface {
	Handler() recorder.HookFunc
	Kind() recorder.HookKind
}

/******************************
	Request Matcher Logic Ops
 ******************************/

// Note: this is currently not used, we kept it for reference

// AndMatcher generic AND operator of given matchers
// Note: because golang generics requires instantiation, type casting is required.
// 		 e.g. var m RecordBodyMatcherFunc = RecordBodyMatcherFunc(AndMatcher(matcher1, matcher2))
//func AndMatcher[O, R any](matchers ...GenericMatcherFunc[O, R]) GenericMatcherFunc[O, R] {
//	return func(out O, record R) error {
//		for _, matcher := range matchers {
//			if e := matcher(out, record); e != nil {
//				return e
//			}
//		}
//		return nil
//	}
//}

// OrMatcher generic OR operator of given matchers
// Note: because golang generics requires instantiation, type casting is required.
// 		 e.g. var m RecordBodyMatcherFunc = RecordBodyMatcherFunc(AndMatcher(matcher1, matcher2))
//func OrMatcher[O, R any](matchers ...GenericMatcherFunc[O, R]) GenericMatcherFunc[O, R] {
//	return func(out O, record R) error {
//		var e error
//		for _, matcher := range matchers {
//			if e = matcher(out, record); e == nil {
//				return nil
//			}
//		}
//		return e
//	}
//}
