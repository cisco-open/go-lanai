package passwd

type PasswordEncoder interface {
	Encode(rawPassword string) string
	Matches(raw, encoded string) bool
}


type noopPasswordEncoder string

func NewNoopPasswordEncoder() PasswordEncoder {
	return noopPasswordEncoder("clear text")
}
func (noopPasswordEncoder) Encode(rawPassword string) string {
	return rawPassword
}

func (noopPasswordEncoder) Matches(raw, encoded string) bool {
	return raw == encoded
}