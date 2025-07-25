package cmdtest

import (
	"context"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/cisco-open/go-lanai/pkg/utils/matcher"
	"github.com/cisco-open/go-lanai/test"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
	"testing"
)

func SetupDryRun(handler TestDryRunHandler) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		cmdutils.GlobalArgs.DryRun = true
		if handler != nil {
			cmdutils.GlobalArgs.DryRunFunc = handler.Handle
		}
		return ctx, nil
	}
}

type TestDryRunHandler map[cmdutils.DryRunCmdType]cmdutils.DryRunFunc

func (h TestDryRunHandler) Handle(ctx context.Context, cmdType cmdutils.DryRunCmdType, args ...interface{}) (interface{}, error) {
	if handler, ok := h[cmdType]; ok {
		return handler(ctx, cmdType, args...)
	}
	return nil, cmdutils.ErrDryRunIgnored
}

func DryRunShellWithFilter(ignoreMatched bool, matchers ...matcher.StringMatcher) cmdutils.DryRunFunc {
	return func(ctx context.Context, cmdType cmdutils.DryRunCmdType, args ...interface{}) (interface{}, error) {
		if cmdType != cmdutils.DryRunTypeShell || len(args) < 4 {
			return nil, cmdutils.ErrDryRunIgnored
		}

		var cmd string
		var runner *interp.Runner
		var parsedCmd *syntax.File
		for _, arg := range args {
			switch v := arg.(type) {
			case string:
				cmd = v
			case *interp.Runner:
				runner = v
			case *syntax.File:
				parsedCmd = v
			case *cmdutils.ShCmdOption:
				// do nothing
			default:
				return nil, cmdutils.ErrDryRunIgnored
			}
		}

		for _, m := range matchers {
			ok, e := m.MatchesWithContext(ctx, cmd)
			if e != nil || ignoreMatched && ok || !ignoreMatched && !ok || runner == nil || parsedCmd == nil {
				continue
			}
			if e := runner.Run(ctx, parsedCmd); e != nil {
				if status, ok := interp.IsExitStatus(e); ok {
					return status, e
				}
				return uint8(1), e
			}
			return 0, nil
		}

		return nil, cmdutils.ErrDryRunIgnored
	}
}

