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
	"gotest.tools/gotestsum":"v1.8.0",
	"github.com/golangci/golangci-lint/cmd/golangci-lint":"v1.46.2",
	"github.com/jstemmer/go-junit-report":"v0.9.1",
}

func installBinaries(ctx context.Context) error {
	opts := []cmdutils.ShCmdOptions{cmdutils.ShellShowCmd(true), cmdutils.ShellUseWorkingDir(),cmdutils.ShellStdOut(os.Stdout)}

	binaries := make(map[string]string)

	for p, v := range defaultBinaries {
		binaries[p] = v
	}
	for _, b := range Module.Binaries {
		if b.Package == "" || b.Version == "" {
			logger.Warnf(`Invalid binaries entry in Module.yml: package="%s", version="%s"`, b.Package, b.Version)
			continue
		}
		binaries[b.Package] = b.Version
	}

	for p, v := range binaries {
		installCmd := fmt.Sprintf("go install %s@%s", p, v)
		opts = append(opts, cmdutils.ShellCmd(installCmd))
	}

	_, e := cmdutils.RunShellCommands(ctx, opts...)
	return e
}
