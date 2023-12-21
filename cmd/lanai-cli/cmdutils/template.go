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

package cmdutils

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"text/template"
)

type TemplateOption struct {
	FS         embed.FS
	TmplName   string      // template name
	Output     string      // output path
	OutputPerm os.FileMode // output file permission when create
	Overwrite  bool        // should overwrite if output file already exists
	Model      interface{}
	Customizer func(*template.Template)
	CommonTmpl string
}

// GenerateFileWithOption generate file using given FS and template name
func GenerateFileWithOption(ctx context.Context, opt *TemplateOption) error {
	if !path.IsAbs(opt.Output) {
		return fmt.Errorf("template output path should be absolute, but got [%s]", opt.Output)
	}

	// prepare output folder
	dir := path.Dir(opt.Output)
	if e := mkdirIfNotExists(dir); e != nil {
		return fmt.Errorf("unable to create directory of template output [%s]", dir)
	}

	// prepare output file to write, return fast if file already exists and overwrite is not allowed
	if isFileExists(opt.Output) && !opt.Overwrite {
		logger.WithContext(ctx).Infof("file [%s] already exists", opt.Output)
		return nil
	}

	f, e := os.OpenFile(opt.Output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, opt.OutputPerm)
	if e != nil {
		return e
	}
	defer func() { _ = f.Close() }()

	// load template and generate
	t := template.New(opt.TmplName)
	if opt.Customizer != nil {
		opt.Customizer(t)
	}
	// load common templates
	if opt.CommonTmpl != "" {
		t, e = t.ParseFS(opt.FS, opt.CommonTmpl)
		if e != nil {
			return e
		}

	}
	t, e = t.ParseFS(opt.FS, opt.TmplName)
	if e != nil {
		return e
	}

	e = t.ExecuteTemplate(f, path.Base(opt.TmplName), opt.Model)
	if e != nil {
		return e
	}

	logger.WithContext(ctx).Infof("Generated file [%s]", opt.Output)
	return nil
}
