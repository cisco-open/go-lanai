package initcmd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"os"
)

var defaultBinaries = map[string]string{
	"github.com/axw/gocov/gocov":"v1.0.0",
	"github.com/AlekSi/gocov-xml":"v1.0.0",
	"github.com/jstemmer/go-junit-report":"v0.9.1",
}

func installBinaries(ctx context.Context) error {
	opts := []cmdutils.ShCmdOptions{cmdutils.ShellShowCmd(true), cmdutils.ShellUseWorkingDir(),cmdutils.ShellStdOut(os.Stdout)}

	binaries := make(map[string]string)

	for p, v := range defaultBinaries {
		binaries[p] = v
	}
	for p, b := range Module.Binaries {
		binaries[p] = b.Version
	}

	for p, v := range binaries {
		opts = append(opts, cmdutils.ShellCmd(fmt.Sprintf("go install %s@%s", p, v)))
	}

	_, e := cmdutils.RunShellCommands(ctx, opts...)
	return e
}
