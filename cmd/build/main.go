package main

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/build/noop"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/build/webjars"
	"github.com/spf13/cobra"
	"os"
)



var rootCmd = &cobra.Command{
	Use: "go-lanai-build",
	Short: "A go-lanai CLI building tool.",
	Long: "This is a go-lanai CLI building tool.",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
}

func init() {
	rootCmd.AddCommand(noop.Cmd)
	rootCmd.AddCommand(webjars.Cmd)
}

func main() {
	if e := rootCmd.ExecuteContext(context.Background()); e != nil {
		os.Exit(1)
	}
}