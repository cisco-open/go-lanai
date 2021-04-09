package common

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"

//TODO: here we want to have the redis db index, and the redis key name and the redis set key name
// move these out of the session package into a package that doesn't have side effects.

type redisTimeoutApplier struct {
	rc redis.Client
}
