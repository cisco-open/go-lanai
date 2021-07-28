package auth

const (
	_ = iota * 100
	TokenEnhancerOrderExpiry
	TokenEnhancerOrderBasicClaims
	TokenEnhancerOrderDetailsClaims
	TokenEnhancerOrderResourceIdClaims
	TokenEnhancerOrderTokenDetails
	TokenEnhancerOrderRefreshToken
	//TokenEnhancerOrder
)

const (
	ctxKeyValidResponseType      = "kVRT"
)

