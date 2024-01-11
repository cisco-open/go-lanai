package testdata

import "embed"

//go:embed manager_accts_test.yml
var TestAcctsFS embed.FS

//go:embed manager_basic_test.yml
var TestBasicFS embed.FS

//go:embed manager_alt_test.yml
var TestAltFS embed.FS
