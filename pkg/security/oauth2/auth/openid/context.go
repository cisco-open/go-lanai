package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

var (
	SupportedGrantTypes = utils.NewStringSet(
		oauth2.GrantTypeAuthCode,
		oauth2.GrantTypeImplicit,
		oauth2.GrantTypePassword,
		oauth2.GrantTypeSwitchUser,
		oauth2.GrantTypeSwitchTenant,
	)
)
