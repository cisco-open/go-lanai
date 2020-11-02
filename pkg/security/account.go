package security

type Account interface {
	Username() string
	Password() string

}

type AccountStore interface {
	LoadAccountByUsername(username string) (Account, error)
}


