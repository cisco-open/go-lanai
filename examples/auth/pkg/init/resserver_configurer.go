package serviceinit

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"

// newResServerConfigurer provide over configuration on oauth.
// resserver.ResourceServerConfigurer is required in DI container to enable resserver.Use
// resserver.Use allows tokenauth.New to be used in security.Configurer
func newResServerConfigurer() resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		//don't need to configure anything here, the default value is good enough
	}
}
