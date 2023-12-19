package vault

import "strings"

const (
	Token      = AuthMethod("token")
	Kubernetes = AuthMethod("kubernetes")
)

var refreshable = map[AuthMethod]struct{}{
	Kubernetes: {},
}

type AuthMethod string

// UnmarshalText encoding.TextUnmarshaler
func (a *AuthMethod) UnmarshalText(data []byte) error {
	*a = AuthMethod(strings.ToUpper(string(data)))
	return nil
}

func (a AuthMethod) isRefreshable() bool {
	_, ok := refreshable[a]
	return ok
}
