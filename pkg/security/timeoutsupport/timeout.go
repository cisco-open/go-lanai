package timeoutsupport

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"strconv"
	"time"
)

//TODO: double check if token-session connection is only saved for certain clients or all clients

type RedisTimeoutApplier struct {
	client redis.Client
}

func NewRedisTimeoutApplier(ctx context.Context, cf redis.ClientFactory, dbIndex int) *RedisTimeoutApplier {
	client, err := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = dbIndex
	})

	if err != nil {
		panic(err)
	}

	return &RedisTimeoutApplier{
		client: client,
	}
}

//TODO: double check all token and context have expiration, so we don't need to manually expire them here, as long as they are not usable, it's good enough
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

	hmGetCmd := r.client.HMGet(ctx, key, common.SessionIdleTimeoutMilli, common.SessionAbsTimeoutTime)
	if hmGetCmd.Err() != nil {
		err = hmGetCmd.Err()
		return
	}
	result, _ := hmGetCmd.Result()

	//TODO: error handling of these conversions
	idleTimeoutMilli, _ := strconv.ParseInt(result[0].(string), 10, 0)
	idleTimeout := time.Duration(idleTimeoutMilli) * time.Millisecond
	absTimeoutUnixTime, _ := strconv.ParseInt(result[1].(string), 10, 0)
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
	var args []interface{}
	args = append(args, common.SessionLastAccessedField, now.Unix())
	hsetCmd := r.client.HSet(ctx, key, args...)
	if hsetCmd.Err() != nil {
		err = hsetCmd.Err()
		return
	}
	expireCmd := r.client.ExpireAt(ctx, key, expiration)
	err = expireCmd.Err()
	return
}