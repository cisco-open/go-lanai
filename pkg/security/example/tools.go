// +build tools

package example

import (
	_ "github.com/mholt/archiver/v3/cmd/arc"
)

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// Keep the github.com/mholt/archiver/cmd/arc in go.mod