package vault

import "strings"

type AuthMethod int

const (
	Token AuthMethod = iota
)

const (
	TokenText   = "TOKEN"
)


var (
	authMethodAtoI = map[string]AuthMethod{
		strings.ToUpper(TokenText):   Token,
	}

	authMethodItoA = map[AuthMethod]string{
		Token:   TokenText,
	}
)

// fmt.Stringer
func (a AuthMethod) String() string {
	if s, ok := authMethodItoA[a]; ok {
		return s
	}
	return "unknown"
}

// encoding.TextMarshaler
func (a AuthMethod) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// encoding.TextUnmarshaler
func (a *AuthMethod) UnmarshalText(data []byte) error {
	value := strings.ToUpper(string(data))
	if v, ok := authMethodAtoI[value]; ok {
		*a = v
	}
	return nil
}