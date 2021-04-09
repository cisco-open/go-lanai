package common

import "fmt"

const RedisNameSpace = "LANAI:SESSION" //This is to avoid confusion with records from other frameworks.
const SessionLastAccessedField = "lastAccessed"

func GetRedisSessionKey(name string, id string) string {
	return fmt.Sprintf("%s:%s:%s", RedisNameSpace, name, id)
}
