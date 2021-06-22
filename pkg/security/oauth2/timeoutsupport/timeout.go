package timeoutsupport

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"strconv"
	"time"
)

type RedisTimeoutApplier struct {
	client redis.Client
}

func NewRedisTimeoutApplier(client redis.Client) *RedisTimeoutApplier {
	return &RedisTimeoutApplier{
		client: client,
	}
}

func(r *RedisTimeoutApplier) ApplyTimeout(ctx context.Context, sessionId string) (valid bool, err error) {
	key := common.GetRedisSessionKey(common.DefaultName, sessionId)

	//check if session exists
	existCmd := r.client.Exists(ctx, key)
	if existCmd.Err() != nil {
		valid = false
		err = existCmd.Err()
		return
	} else {
		valid = existCmd.Val() == 1
	}

	if !valid {
		return
	}

	hmGetCmd := r.client.HMGet(ctx, key, common.SessionIdleTimeoutDuration, common.SessionAbsTimeoutTime)
	if hmGetCmd.Err() != nil {
		err = hmGetCmd.Err()
		return
	}
	result, _ := hmGetCmd.Result()

	idleTimeout, err := time.ParseDuration(result[0].(string))
	if err != nil {
		return
	}
	absTimeoutUnixTime, err := strconv.ParseInt(result[1].(string), 10, 0)
	if err != nil {
		return
	}
	absExpiration := time.Unix(absTimeoutUnixTime, 0)

	now := time.Now()
	idleExpiration := now.Add(idleTimeout)
	var expiration time.Time

	//whichever is the earliest
	if idleExpiration.Before(absExpiration) {
		expiration = idleExpiration
	} else {
		expiration = absExpiration
	}

	//update session last accessed time
	hsetCmd := r.client.HSet(ctx, key, common.SessionLastAccessedField, now.Unix())
	if hsetCmd.Err() != nil {
		err = hsetCmd.Err()
		return
	}
	expireCmd := r.client.ExpireAt(ctx, key, expiration)
	err = expireCmd.Err()
	return
}