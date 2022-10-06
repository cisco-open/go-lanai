package main

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/apidocs"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/build"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/deps"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/gittools"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/initcmd"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/noop"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/webjars"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

const (
	CliName = "lanai-cli"
)

var (
	BuildVersion = "unknown"
)

var (
	rootCmd = &cobra.Command{
		Use:                CliName,
		Short:              "A go-lanai CLI building tool.",
		Long:               "This is a go-lanai CLI building tool.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PersistentPreRunE: cmdutils.MergeRunE(
			cmdutils.EnsureGlobalDirectories(),
			cmdutils.PrintEnvironment(),
		),
	}
	logTemplate = `{{pad -25 .time}} [{{lvl 5 .}}]: {{.msg}}`
	logProps    = log.Properties{
		Levels: map[string]log.LoggingLevel{
			"default": log.LevelDebug,
		},
		Loggers: map[string]*log.LoggerProperties{
			"console": {
				Type:     log.TypeConsole,
				Format:   log.FormatText,
				Template: logTemplate,
				FixedKeys: utils.CommaSeparatedSlice{
					log.LogKeyName, log.LogKeyMessage, log.LogKeyTimestamp,
					log.LogKeyCaller, log.LogKeyLevel, log.LogKeyContext,
				},
			},
		},
		Mappings: map[string]string{},
	}
)

func init() {
	if e := log.UpdateLoggingConfiguration(&logProps); e != nil {
		panic(e)
	}
}

func main() {
	rootCmd.Version = BuildVersion
	rootCmd.AddCommand(noop.Cmd)
	rootCmd.AddCommand(webjars.Cmd)
	rootCmd.AddCommand(initcmd.Cmd)
	rootCmd.AddCommand(deps.Cmd)
	rootCmd.AddCommand(gittools.Cmd)
	rootCmd.AddCommand(build.Cmd)
	rootCmd.AddCommand(apidocs.Cmd)
	rootCmd.AddCommand(codegen.Cmd)

	cmdutils.PersistentFlags(rootCmd, &cmdutils.GlobalArgs)
	if e := rootCmd.ExecuteContext(context.Background()); e != nil {
		os.Exit(1)
	}
}
