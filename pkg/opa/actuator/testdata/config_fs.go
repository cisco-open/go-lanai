package testdata

import "embed"

//go:embed *.yml
var TestConfigFS embed.FS

//go:embed bundles/**
var ActuatorBundleFS embed.FS