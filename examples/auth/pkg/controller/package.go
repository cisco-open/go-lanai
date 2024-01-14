package controller

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

func Use() {
	bootstrap.AddOptions(
		fx.Invoke(configureWeb),
	)
}

func configureWeb(r *web.Registrar) {
	r.MustRegister(NewLoginFormController())
}
