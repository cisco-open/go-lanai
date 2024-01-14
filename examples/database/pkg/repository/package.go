package repository

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

func Use() {
	bootstrap.AddOptions(
		fx.Provide(
			NewFriendRepository,
		),
	)
}
