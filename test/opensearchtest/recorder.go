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

package opensearchtest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
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

// NewRecorder will create a recorder configured by the RecordOptions
func NewRecorder(options ...Options) (*recorder.Recorder, error) {
	var recordOption Option
	for _, fn := range options {
		fn(&recordOption)
	}
	if recordOption.Name == "" {
		return nil, ErrNoCassetteName
	}
	rec, e := ittest.NewHttpRecorder(toHTTPVCROptions(recordOption))
	if e != nil {
		return nil, fmt.Errorf("%w, %v", ErrCreatingRecorder, e)
	}
	return rec.Recorder, nil
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
	panic(fmt.Errorf("open was not called directly or indirectly from a test file"))
}

func toHTTPVCROptions(opt Option) ittest.HTTPVCROptions {
	return func(vcrOpt *ittest.HTTPVCROption) {
		vcrOpt.Mode = ittest.ModeReplaying
		switch opt.Mode {
		case ModeRecording:
			vcrOpt.Mode = ittest.ModeRecording
		case ModeCommandline:
			if IsRecording() {
				vcrOpt.Mode = ittest.ModeRecording
			}
		default:
		}
		vcrOpt.Name = opt.Name
		vcrOpt.SavePath = opt.SavePath
		vcrOpt.RecordMatching = append(vcrOpt.RecordMatching, func(matcherOpt *ittest.RecordMatcherOption) {
			matcherOpt.BodyMatchers = append(matcherOpt.BodyMatchers, BulkJsonBodyMatcher{
				Delegate: ittest.NewRecordJsonBodyMatcher(opt.FuzzyJsonPaths...),
			})
			matcherOpt.FuzzyHeaders = append(matcherOpt.FuzzyHeaders, "User-Agent")
		})
	}
}
