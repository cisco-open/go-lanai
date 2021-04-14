package main

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/initcmd"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/noop"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/webjars"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:                "go-lanai-build",
		Short:              "A go-lanai CLI building tool.",
		Long:               "This is a go-lanai CLI building tool.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PersistentPreRunE: cmdutils.MergeRunE(
			cmdutils.EnsureGlobalDirectories(),
			cmdutils.PrintEnvironment(),
		),
	}
	logTemplate = `{{pad .time -25}} [{{lvl . 5}}]: {{.msg}}`
	logProps = log.Properties{
		Levels: map[string]log.LoggingLevel{
			"default": log.LevelDebug,
		},
		Loggers:  map[string]log.LoggerProperties{
			"console": {
				Type: log.TypeConsole,
				Format: log.FormatText,
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
	rootCmd.AddCommand(noop.Cmd)
	rootCmd.AddCommand(webjars.Cmd)
	rootCmd.AddCommand(initcmd.Cmd)

	cmdutils.PersistentFlags(rootCmd, &cmdutils.GlobalArgs)
	if e := rootCmd.ExecuteContext(context.Background()); e != nil {
		os.Exit(1)
	}
}
