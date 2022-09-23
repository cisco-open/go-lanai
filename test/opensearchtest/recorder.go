package opensearchtest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"path"
	"runtime"
	"strings"
)

var (
	ErrCreatingRecorder = errors.New("unable to create recorder")
	ErrNoCassetteName   = errors.New("requires cassette name")
)

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
	Mode      Mode
	Modifiers *MatcherBodyModifiers
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
	r.SetMatcher(MatchBody(recordOption.Modifiers))
	return r, nil
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
