package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

func provideGormErrorTranslator() web.ErrorTranslator {
	return data.NewGormErrorTranslator()
}

func provideDataErrorTranslator() web.ErrorTranslator {
	return data.NewDataErrorTranslator()
}
