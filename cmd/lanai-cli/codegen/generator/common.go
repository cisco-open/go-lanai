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

package generator

import (
    "errors"
    "fmt"
    "github.com/bmatcuk/doublestar/v4"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "io"
    "os"
    "path"
    "path/filepath"
    "strings"
    "text/template"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func GenerateFileFromTemplate(gc GenerationContext, template *template.Template) error {
	if gc.templatePath == "" {
		return fmt.Errorf("no templatePath name defined")
	}
	if !path.IsAbs(gc.filename) {
		return fmt.Errorf("templatePath output path should be absolute, but got [%s]", gc.filename)
	}

	outputFolder := path.Dir(gc.filename)
	if err := mkdirIfNotExists(outputFolder); err != nil {
		return fmt.Errorf("unable to create directory of templatePath output [%s]", outputFolder)
	}
	wc, err := applyRegenRule(&gc)
	if err != nil {
		return err
	}
	defer func() { _ = wc.Close() }()

	if err := template.ExecuteTemplate(wc, gc.templatePath, gc.model); err != nil {
		return fmt.Errorf("templatePath could not be executed: %v", err)
	}

	if e := FormatFile(gc.filename, FileTypeUnknown); e != nil && !errors.Is(e, errFormatterUnsupportedFileType) {
		return fmt.Errorf("error formatting go code for file %v: %v", gc.filename, e)
	}

	return nil
}

func mkdirIfNotExists(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s is not absolute path", path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if e := os.MkdirAll(path, 0755); e != nil {
			return e
		}
	}
	return nil
}

type emptyWriteCloser struct{}

func (e *emptyWriteCloser) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (e *emptyWriteCloser) Close() (err error) {
	return nil
}

func applyRegenRule(gc *GenerationContext) (f io.WriteCloser, err error) {
	if fileExists(gc.filename) {
		switch gc.regenMode {
		case RegenModeIgnore:
			logger.Infof("ignore rule defined for existing file %v, ignoring", gc.filename)
			//	make an empty applyRegenRule to allow the template to be executed (and keep any runtime logic consistent)
			return &emptyWriteCloser{}, nil
		case RegenModeReference:
			gc.filename += "ref"
			fallthrough
		case RegenModeOverwrite:
			break
		}
	}
	f, err = os.OpenFile(gc.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func copyOf(data map[string]interface{}) map[string]interface{} {
	dataCopy := make(map[string]interface{})
	for k, v := range data {
		dataCopy[k] = v
	}
	return dataCopy
}

func getApplicableRegenRules(outputFile string, rules RegenRules, defaultMode RegenMode) (RegenMode, error) {
	pathAfterOutputDir := strings.TrimPrefix(outputFile, cmdutils.GlobalArgs.OutputDir+"/")
	mode := defaultMode
	for _, r := range rules {
		match, err := doublestar.Match(r.Pattern, pathAfterOutputDir)
		if err != nil {
			return "", err
		}
		if match {
			mode = r.Mode
		}
	}
	return mode, nil
}

// counter used to count touched file, grouped by directories
type counter map[string]int

func (c counter) Record(filePath string) {
	dir := filepath.Dir(filePath)
	v, _ := c[dir]
	c[dir] = v + 1
}

func (c counter) Cleanup(filePath string) {
	dir := filepath.Dir(filePath)
	v, _ := c[dir]
	c[dir] = v - 1
}
