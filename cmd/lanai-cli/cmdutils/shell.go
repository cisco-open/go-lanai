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
	"io"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
	"os"
	"strings"
)

// ShCmdLogDisabled Disable shell command logging. When set to true, ShCmdOption.ShowCmd no longer take effect
var ShCmdLogDisabled bool

// ShCmdOptions is the options for the RunShellCommands func
type ShCmdOptions func(opt *ShCmdOption)

type ShCmdOption struct {
	Cmds    []string
	Dir     string
	Env     []string
	ShowCmd bool
	Stdin io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// ShellCmd add shell commends
func ShellCmd(cmds ...string) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.Cmds = append(opt.Cmds, cmds...)
	}
}

// ShellEnv set additional environment variable in format of "key=value"
func ShellEnv(env ...string) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.Env = append(opt.Env, env...)
	}
}

// ShellDir set working dir of to-be-executed commands
func ShellDir(dir string) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.Dir = dir
	}
}

// ShellUseWorkingDir set working dir of to-be-executed commands to GlobalArgs.WorkingDir
func ShellUseWorkingDir() ShCmdOptions {
	return ShellDir(GlobalArgs.WorkingDir)
}

// ShellUseTmpDir set working dir of to-be-executed commands to GlobalArgs.TmpDir
func ShellUseTmpDir() ShCmdOptions {
	return ShellDir(GlobalArgs.TmpDir)
}

// ShellUseOutputDir set working dir of to-be-executed commands to GlobalArgs.OutputDir
func ShellUseOutputDir() ShCmdOptions {
	return ShellDir(GlobalArgs.OutputDir)
}

// ShellShowCmd toggle display command in log
func ShellShowCmd(show bool) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.ShowCmd = show
	}
}

// ShellStdIn toggle display command in log
func ShellStdIn(r io.Reader) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.Stdin = r
	}
}

// ShellStdOut toggle display command in log
func ShellStdOut(w io.Writer) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.Stdout = w
	}
}

// ShellStdErr toggle display command in log
func ShellStdErr(w io.Writer) ShCmdOptions {
	return func(opt *ShCmdOption) {
		opt.Stderr = w
	}
}

// RunShellCommands runs a shell command, returns exit status and an error
func RunShellCommands(ctx context.Context, opts ...ShCmdOptions) (uint8, error) {
	opt := ShCmdOption{
		Cmds: []string{},
		Dir: GlobalArgs.TmpDir,
		Env: os.Environ(),
		Stdin: os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	for _, f := range opts {
		f(&opt)
	}

	for _, cmd := range opt.Cmds {
		if exit, e := runSingleCommand(ctx, cmd, &opt); e != nil {
			return exit, e
		}
	}
	return 0, nil
}

func runSingleCommand(ctx context.Context, cmd string, opt *ShCmdOption) (uint8, error){
	p, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return 1, err
	}

	r, err := interp.New(
		interp.Params("-e"),
		interp.Dir(opt.Dir),
		interp.Env(expand.ListEnviron(opt.Env...)),
		interp.OpenHandler(openHandler),
		interp.StdIO(opt.Stdin, opt.Stdout, opt.Stderr),
	)
	if err != nil {
		return 1, err
	}

	if opt.ShowCmd && !ShCmdLogDisabled {
		logger.WithContext(ctx).Infof("Shell Command: %s", cmd)
	}

	if e := r.Run(ctx, p); e != nil {
		if status, ok := interp.IsExitStatus(err); ok {
			return status, e
		}
		return 1, e
	}
	return 0, nil
}

type nop struct {}
func (nop) Read(_ []byte) (int, error)  { return 0, io.EOF }
func (nop) Write(p []byte) (int, error) { return len(p), nil }
func (nop) Close() error                { return nil }


func openHandler(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if path == "/dev/null" {
		return nop{}, nil
	}
	return interp.DefaultOpenHandler()(ctx, path, flag, perm)
}