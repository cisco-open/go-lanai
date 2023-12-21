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
	"context"
	"io/fs"
	"os"
)

type DirectoryGenerator struct {
	data       map[string]interface{}
	templateFS fs.FS
	matcher        TemplateMatcher
	outputResolver TemplateOutputResolver
}

type DirOption struct {
	GeneratorOption
	Matcher TemplateMatcher
	OutputResolver TemplateOutputResolver
}

func newDirectoryGenerator(gOpt GeneratorOption, opts ...func(option *DirOption)) *DirectoryGenerator {
	o := &DirOption{
		GeneratorOption: gOpt,
		Matcher: isDir(),
		OutputResolver: regexOutputResolver(""),
	}
	for _, fn := range opts {
		fn(o)
	}
	return &DirectoryGenerator{
		data:           o.Data,
		templateFS:     o.TemplateFS,
		matcher:        o.Matcher,
		outputResolver: o.OutputResolver,
	}
}

func (d *DirectoryGenerator) Generate(ctx context.Context, tmplDesc TemplateDescriptor) error {
	if ok, e := d.matcher.Matches(tmplDesc); e != nil || !ok {
		return e
	}

	output, e := d.outputResolver.Resolve(ctx, tmplDesc, d.data)
	if e != nil {
		return e
	}

	logger.Debugf("[Dir] generating %v", output.Path)
	if err := os.MkdirAll(output.Path, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
