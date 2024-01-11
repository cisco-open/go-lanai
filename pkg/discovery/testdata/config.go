package testdata

import "embed"

//go:embed bootstrap-test.yml
var TestBootstrapFS embed.FS

//go:embed application-test.yml
var TestApplicationFS embed.FS
