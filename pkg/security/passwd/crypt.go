package passwd

import "golang.org/x/crypto/bcrypt"

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

// bcryptPasswordEncoder implements PasswordEncoder
type bcryptPasswordEncoder struct {
	cost int
}

func NewBcryptPasswordEncoder() PasswordEncoder {
	return &bcryptPasswordEncoder{
		cost: 10,
	}
}

func (enc *bcryptPasswordEncoder) Encode(raw string) string {
	encoded, e := bcrypt.GenerateFromPassword([]byte(raw), enc.cost)
	if e != nil {
		return ""
	}
	return string(encoded)
}

func (enc *bcryptPasswordEncoder) Matches(raw, encoded string) bool {
	e := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(raw))
	return e == nil
}