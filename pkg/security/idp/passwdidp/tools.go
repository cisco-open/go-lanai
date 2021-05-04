// +build tools

package passwdidp

import (
	_ "github.com/mholt/archiver/cmd/arc"
)

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// Keep the github.com/mholt/archiver/cmd/arc in go.mod