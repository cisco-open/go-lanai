package oauth2

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

type ClientDetails interface {
	ClientId() string
	AssignedTenantIds() utils.StringSet
	Scopes() utils.StringSet
}
