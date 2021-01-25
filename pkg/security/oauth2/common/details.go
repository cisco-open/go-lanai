package common

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common/internal"

func NewContextDetails() *internal.ContextDetails {
	return &internal.ContextDetails{
		KV: map[string]interface{}{},
	}
}


