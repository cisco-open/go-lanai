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

// ShCmdOptions is the options for the RunShellCommands func
type ShCmdOptions func(opt *ShCmdOption)

type ShCmdOption struct {
	Cmds    []string
	Dir     string
	Env     []string
	ShowCmd bool
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

// RunShellCommands runs a shell command, returns exit status and an error
func RunShellCommands(ctx context.Context, opts ...ShCmdOptions) (uint8, error) {
	opt := ShCmdOption{
		Cmds: []string{},
		Dir: GlobalArgs.TmpDir,
		Env: os.Environ(),
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
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
	)
	if err != nil {
		return 1, err
	}

	if opt.ShowCmd {
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