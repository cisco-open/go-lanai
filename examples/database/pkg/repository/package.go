package repository

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

func Use() {
	bootstrap.AddOptions(
		fx.Provide(
			NewFriendRepository,
		),
	)
}
