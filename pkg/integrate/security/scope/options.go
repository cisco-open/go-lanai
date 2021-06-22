package scope

func WithUsername(username string) Options {
	return func(s *Scope) {
		s.username = username
		s.userId = ""
	}
}

func WithUserId(userId string) Options {
	return func(s *Scope) {
		s.username = ""
		s.userId = userId
	}
}

func WithTenantId(tenantId string) Options {
	return func(s *Scope) {
		s.tenantName = ""
		s.tenantId = tenantId
	}
}

func WithTenantName(tenantName string) Options {
	return func(s *Scope) {
		s.tenantName = tenantName
		s.tenantId = ""
	}
}

func UseSystemAccount() Options {
	return func(s *Scope) {
		s.useSysAcct = true
	}
}
