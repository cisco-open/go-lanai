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

package web

import (
    "errors"
    "github.com/bmatcuk/doublestar/v4"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
    "io/fs"
    "os"
    "path"
    "path/filepath"
    "strings"
)

const (
	DirFSAllowListDirectory DirFSOption = 1 << iota
	//...
)

type DirFSOption int64

// dirFS implements fs.FS and fs.GlobFS. It is similar to http.Dir, but support fs.FS and allow option to not list directory contents
type dirFS struct {
	dir  string
	fs   fs.FS
	opts DirFSOption
}

func NewOSDirFS(dir string, opts ...DirFSOption) fs.FS {
	return NewDirFS("", os.DirFS(dir), opts...)
}

func NewDirFS(dir string, fsys fs.FS, opts ...DirFSOption) fs.FS {
	options := DirFSOption(0)
	for _, opt := range opts {
		options = options | opt
	}
	return &dirFS{
		dir:  dir,
		fs:   fsys,
		opts: options,
	}
}

func (f *dirFS) Open(name string) (fs.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("invalid character in file path")
	}
	dir := f.dir
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
	file, e := f.fs.Open(fullName)
	if e != nil {
		return nil, f.translateError(e, fullName)
	}

	// apply options
	if !f.hasOption(DirFSAllowListDirectory) {
		if stat, e := file.Stat(); e != nil {
			return nil, f.translateError(e, fullName)
		} else if stat.IsDir() {
			return nil, fs.ErrNotExist
		}
	}
	return file, nil
}

func (f *dirFS) Glob(pattern string) (ret []string, err error) {
	return doublestar.Glob(f.fs, pattern)
}

func (f *dirFS) hasOption(opt DirFSOption) bool {
	return f.opts & opt != 0
}

// translateError maps the provided non-nil error from opening name
// to a possibly better non-nil error. In particular, it turns OS-specific errors
// about opening files in non-directories into fs.ErrNotExist. see http.mapDirOpenError
func (f *dirFS) translateError(err error, name string) error {
	if err == fs.ErrNotExist || err == fs.ErrPermission {
		return err
	}

	parts := strings.Split(name, string(filepath.Separator))
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		fi, e := os.Stat(strings.Join(parts[:i+1], string(filepath.Separator)))
		if e != nil {
			return e
		} else if fi != nil && !fi.IsDir() {
			return fs.ErrNotExist
		}
	}
	return err
}

// orderedFS implements fs.FS and order.Ordered
type orderedFS struct {
	fs.FS
	order int
}

// OrderedFS returns a fs.FS that also implements order.Ordered
// if the given fs.FS is already implement the order.Ordered, "defaultOrder" is ignored
func OrderedFS(fsys fs.FS, defaultOrder int) fs.FS {
	return &orderedFS{
		FS: fsys,
		order: defaultOrder,
	}
}

func (f orderedFS) Order() int {
	return f.order
}

// MergedFS implements fs.FS and fs.GlobFS
type MergedFS struct {
	srcFS []fs.FS
}

func NewMergedFS(atLeastOne fs.FS, fs ...fs.FS) *MergedFS {
	src := append(fs, atLeastOne)
	order.SortStable(src, order.OrderedFirstCompare)
	m := &MergedFS{
		srcFS: src,
	}
	return m
}

func (m *MergedFS) Open(name string) (fs.File, error) {
	for _, f := range m.srcFS {
		if file, e := f.Open(name); e == nil {
			return file, nil
		}
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (m *MergedFS) Glob(pattern string) (ret []string, err error) {
	// loop through all FS sources in reversed order
	for i := len(m.srcFS) - 1; i >= 0; i-- {
		paths, e := doublestar.Glob(m.srcFS[i], pattern)
		if e != nil {
			return nil, e
		}
		ret = append(ret, paths...)
	}
	return
}
