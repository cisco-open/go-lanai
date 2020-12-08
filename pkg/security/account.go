package security

type Account interface {
	Username() string
	Password() string
	Permissions() []string
	UseMFA() bool
}

type AccountStore interface {
	LoadAccountByUsername(username string) (Account, error)
}


