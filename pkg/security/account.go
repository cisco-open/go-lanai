package security

type User interface {
	Username() string
	Password() string
}

type AccountStore interface {
	LoadUserByUsername(username string) (User, error)
}


