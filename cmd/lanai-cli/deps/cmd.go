package deps

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
)

var logger = log.New("deps")

func init() {
	cmdutils.PersistentFlags(UpdateDepCmd, &UpdateDepArgs)
	cmdutils.PersistentFlags(DropReplaceCmd, &DropReplaceArgs)
}