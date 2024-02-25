package controller

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/web"
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
