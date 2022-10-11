package testdata

import "embed"

//go:embed *.yml
var TestConfigFS embed.FS
