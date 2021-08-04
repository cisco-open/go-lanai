package claims

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

var (
	UserInfoBasicSpecs = map[string]ClaimSpec{
		oauth2.ClaimIssuer:          Required(Issuer),
		oauth2.ClaimSubject:         Optional(Subject),
		oauth2.ClaimAudience:        Required(Audience),

		// Custom
		oauth2.ClaimAccountType:     Optional(AccountType),
	}
)
