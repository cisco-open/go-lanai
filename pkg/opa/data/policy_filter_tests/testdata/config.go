package testdata

import "embed"

//go:embed application-test.yml
var ConfigFS embed.FS

//go:embed bundles/model_b
var ModelBBundleFS embed.FS
