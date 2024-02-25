package repository

import (
	"github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/model"
	"github.com/cisco-open/go-lanai/pkg/data/repo"
)

type FriendsRepository struct {
	repo.GormApi // In case we need to work with the lower level API directly
	repo.CrudRepository
}

func NewFriendRepository(factory repo.Factory) *FriendsRepository {
	crud := factory.NewCRUD(&model.Friend{})

	ret := FriendsRepository{
		CrudRepository: crud,
	}
	if gf, ok := factory.(*repo.GormFactory); ok {
		ret.GormApi = gf.NewGormApi()
	}
	return &ret
}
