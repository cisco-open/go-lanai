package noop

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	Cmd = &cobra.Command{
		Use:                "noop",
		Short:              "Does nothing",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{
		Str: "Default String",
		StrPtr: &strVal,
		Int: 20,
		IntPtr: &intVal,
		Bool: false,
		BoolPtr: &boolVal,
	}
	strVal = "Default String Ptr"
	intVal = 10
	boolVal = true
)

type Arguments struct {
	Str     string  `flag:"str,s,required" desc:"string"`
	StrPtr  *string `flag:"strptr,p" desc:"*string"`
	Int     int     `flag:"int" desc:"int"`
	IntPtr  *int    `flag:"intptr" desc:"*int"`
	Bool    bool    `flag:"bool,b" desc:"bool"`
	BoolPtr *bool   `flag:"boolptr" desc:"*bool"`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

func Run(_ *cobra.Command, _ []string) error {
	fmt.Println()
	_ = json.NewEncoder(os.Stdout).Encode(Args)
	for _, env := range os.Environ() {
		if strings.Contains(strings.ToUpper(env), "GO") {
			fmt.Println(env)
		}
	}
	return nil
}