package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

// RedisContextDetailsStore implements security.ContextDetailsStore
type RedisContextDetailsStore struct {

}

func (r *RedisContextDetailsStore) ReadContextDetails(c context.Context, key interface{}) (security.ContextDetails, error) {
	panic("implement me")
}

func (r *RedisContextDetailsStore) SaveContextDetails(c context.Context, details security.ContextDetails) (key interface{}, err error) {
	panic("implement me")
}

func (r *RedisContextDetailsStore) RemoveContextDetails(c context.Context, key interface{}) error {
	panic("implement me")
}

