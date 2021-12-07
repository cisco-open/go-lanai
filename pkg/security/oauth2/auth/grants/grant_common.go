package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
)

var logger = log.New("OAuth2.Grant")

func CommonPreGrantValidation(c context.Context, client oauth2.OAuth2Client, request *auth.TokenRequest) error {
	// check grant
	if e := auth.ValidateGrant(c, client, request.GrantType); e != nil {
		return e
	}

	// check scope
	if e := auth.ValidateAllScopes(c, client, request.Scopes); e != nil {
		return e
	}
	return nil
}

