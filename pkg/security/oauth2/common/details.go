package common

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common/internal"

func NewContextDetails() *internal.FullContextDetails {
	return &internal.FullContextDetails{
		KV: map[string]interface{}{},
	}
}


