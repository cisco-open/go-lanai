package common

import (
	"fmt"
	"time"
)

const RedisNameSpace = "LANAI:SESSION" //This is to avoid confusion with records from other frameworks.
const SessionLastAccessedField = "lastAccessed"
const SessionIdleTimeoutDuration = "idleTimeout"
const SessionAbsTimeoutTime = "absTimeout"
const DefaultName = "SESSION"

type TimeoutSetting int

const (
	IdleTimeoutEnabled     TimeoutSetting = 1 << iota
	AbsoluteTimeoutEnabled TimeoutSetting = 1 << iota
)

func GetRedisSessionKey(name string, id string) string {
	return fmt.Sprintf("%s:%s:%s", RedisNameSpace, name, id)
}

func CalculateExpiration(setting TimeoutSetting, idleExpiration time.Time, absExpiration time.Time) (canExpire bool, expiration time.Time) {
	switch setting {
	case AbsoluteTimeoutEnabled:
		return true, absExpiration
	case IdleTimeoutEnabled:
		return true, idleExpiration
	case AbsoluteTimeoutEnabled | IdleTimeoutEnabled:
		//whichever is the earliest
		if idleExpiration.Before(absExpiration) {
			return true, idleExpiration
		} else {
			return true, absExpiration
		}
	default:
		return false, time.Time{}
	}
}