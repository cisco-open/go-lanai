package opensearchtest

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/cassette"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"strings"
)

var (
	ErrCreatingRecorder = errors.New("unable to create recorder")
	ErrNoCassetteName   = errors.New("requires cassette name")
)

// MatchBodyModifier will modify the body of a request that goes to the cassette Matcher
// to remove things that might make matching difficult.
// Example being time parameters in queries, or randomly generated values.
// To see this in use, check out SubTestTimeBasedQuery in opensearch_test.go
type MatchBodyModifier func(*[]byte)

type Mode recorder.Mode

// Recorder states
const (
	ModeRecording Mode = iota
	ModeReplaying
	// ModeCommandline lets the commandline or the state in TestMain to determine the mode
	ModeCommandline
)

// options is unexported because the Properties that these would edit
// are unexported also, so there's no point in creating any options out of this package
type RecordOptions func(c *RecordOption)
type RecordOption struct {
	CassetteLocation   string
	Mode               Mode
	MatchBodyModifiers []MatchBodyModifier
}

func CassetteLocation(location string) RecordOptions {
	return func(c *RecordOption) {
		c.CassetteLocation = location
	}
}

func ReplayMode(mode Mode) RecordOptions {
	return func(c *RecordOption) {
		c.Mode = mode
	}
}

// GetCassetteLocation will look for the testdata/ directory where the test file is.
// If that testdata/ directory does not exist, then this function will create it. The
// recording file if named hello_test.go will be named hello.httpvcr
func GetCassetteLocation() string {
	fileName := findTestFile()
	dirName := path.Join(path.Dir(fileName), "testdata")
	fileName = path.Base(fileName[:len(fileName)-3]) + ".httpvcr"
	pathName := path.Join(dirName, fileName)
	return pathName
}

// WithRecorder will add a recorder configured by the RecordOptions to *opensearch.Properties
func GetRecorder(options ...RecordOptions) (*recorder.Recorder, error) {
	recordOption := RecordOption{}
	for _, fn := range options {
		fn(&recordOption)
	}
	if recordOption.CassetteLocation == "" {
		return nil, ErrNoCassetteName
	}
	httpTransport := http.DefaultTransport
	r, err := recorder.NewAsMode(
		recordOption.CassetteLocation,
		recorder.Mode(recordOption.Mode),
		httpTransport,
	)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrCreatingRecorder, err)
	}
	r.SetInOrderInteractions(true)
	r.SetMatcher(matchBody(recordOption.MatchBodyModifiers))
	return r, nil
}

// matchBody will ensure that the matcher also matches the contents of the body
func matchBody(modifiers []MatchBodyModifier) func(r *http.Request, i cassette.Request) bool {
	return func(r *http.Request, i cassette.Request) bool {
		if r.Body == nil {
			return cassette.DefaultMatcher(r, i)
		}
		var b bytes.Buffer
		if _, err := b.ReadFrom(r.Body); err != nil {
			return false
		}
		r.Body = ioutil.NopCloser(&b)
		requestBody := b.Bytes()
		recordingBody := []byte(i.Body)
		for _, modifier := range modifiers {
			modifier(&requestBody)
			modifier(&recordingBody)
		}
		return cassette.DefaultMatcher(r, i) &&
			(string(requestBody) == "" || string(requestBody) == string(recordingBody))
	}
}

// findTestFile - copied from copyist.go - Searches the call stack, looking for the test that called
// copyist.Open. It searches up to N levels, looking for the last file that
// ends in "_test.go" and returns that filename.
func findTestFile() string {
	const levels = 10
	var lastTestFilename string
	for i := 0; i < levels; i++ {
		_, fileName, _, _ := runtime.Caller(2 + i)
		if strings.HasSuffix(fileName, "_test.go") {
			lastTestFilename = fileName
		}
	}
	if lastTestFilename != "" {
		return lastTestFilename
	}
	panic(fmt.Errorf("Open was not called directly or indirectly from a test file"))
}
