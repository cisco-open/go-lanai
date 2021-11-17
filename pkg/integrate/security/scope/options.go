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
		s.tenantExternalId = ""
		s.tenantId = tenantId
	}
}

func WithTenantExternalId(tenantExternalId string) Options {
	return func(s *Scope) {
		s.tenantExternalId = tenantExternalId
		s.tenantId = ""
	}
}

func UseSystemAccount() Options {
	return func(s *Scope) {
		s.useSysAcct = true
	}
}
