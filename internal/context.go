package internal

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

type TenantAccessDetails interface {
	EffectiveAssignedTenantIds() utils.StringSet
}
