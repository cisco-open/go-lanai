package auth

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

const (
	_ = iota * 100
	TokenEnhancerOrderExpiry
	TokenEnhancerOrderBasicClaims
	TokenEnhancerOrderDetailsClaims
	TokenEnhancerOrderResourceIdClaims
	TokenEnhancerOrderRefreshToken
	//TokenEnhancerOrder
)

var (
	StandardResponseTypes = utils.NewStringSet("token", "code")
)
