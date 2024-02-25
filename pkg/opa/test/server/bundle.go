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

package opatestserver

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    sdktest "github.com/open-policy-agent/opa/sdk/test"
    "go.uber.org/fx"
    "io/fs"
    "path/filepath"
    "strings"
)

var logger = log.New("OPA.Test")

type BundleServerOptions func(cfg *BundleServerConfig)
type BundleServerConfig struct {
	BundleName    string
	BundleSources []fs.FS
}

// WithBundleSources is a BundleServerOptions to add bundle sources from bundle system
func WithBundleSources(fsys ...fs.FS) BundleServerOptions {
	return func(cfg *BundleServerConfig) {
		cfg.BundleSources = append(cfg.BundleSources, fsys...)
	}
}

// WithBundleName is a BundleServerOptions to set bundle name
func WithBundleName(name string) BundleServerOptions {
	return func(cfg *BundleServerConfig) {
		cfg.BundleName = name
	}
}

func NewBundleServer(ctx context.Context, opts ...BundleServerOptions) (*sdktest.Server, error) {
	cfg := BundleServerConfig{
		BundleName:    "test",
		BundleSources: []fs.FS{},
	}
	for _, fn := range opts {
		fn(&cfg)
	}

	policies := map[string]string{}
	for name, fsys := range cfg.BundleSources {
		if e := loadBundleFiles(fsys, policies); e != nil {
			logger.WithContext(ctx).Warnf("unable to load bundle [%s]: ", name, e)
			continue
		}
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("failed to start OPA bundle server, unable to load any bundle")
	}

	ready := make(chan struct{}, 1)
	defer func() { close(ready) }()
	server, e := sdktest.NewServer(sdktest.MockBundle("/bundles/"+cfg.BundleName, policies), sdktest.Ready(ready))
	if e != nil {
		return nil, fmt.Errorf("failed to start OPA bundle server: %v", e)
	}
	logger.WithContext(ctx).Infof("OPA Bundles served at %q", server.URL())
	return server, nil
}

func InitializeBundleServer(lc fx.Lifecycle, server *sdktest.Server) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			server.Stop()
			return nil
		},
	})
}

func loadBundleFiles(fsys fs.FS, dest map[string]string) error {
	// find and read all files
	files := map[string][]byte{}
	rootPath := "."
	e := fs.WalkDir(fsys, rootPath, func(path string, d fs.DirEntry, _ error) error {
		// Note we ignore any error and let it walk through entire tree
		if d.IsDir() {
			return nil
		}
		data, e := fs.ReadFile(fsys, path)
		if e != nil {
			return e
		}
		if d.Name() == ".manifest" {
			rootPath = filepath.Dir(path)
		}
		files[path] = data
		return nil
	})
	if e != nil {
		return e
	} else if len(files) == 0 {
		return fmt.Errorf("no files was found in bundle FS")
	}

	// prepare bundle content
	for path, data := range files {
		name, e := filepath.Rel(rootPath, path)
		if e != nil {
			name = path
		}
		if strings.HasSuffix(name, ".json") {
			// nested data documents are not implemented in the dummy server
			name = strings.ReplaceAll(path, "/", "_")
		}
		dest[name] = string(data)
	}
	return nil
}
