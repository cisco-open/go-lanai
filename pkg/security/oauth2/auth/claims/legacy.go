package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

func LegacyAudiance(ctx context.Context, opt *FactoryOption) utils.StringSet {
	// in the java implementation, Spring uses "aud" for resource IDs which has been deprecated
	client, ok := ctx.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client)
	if !ok || client.ResourceIDs() == nil || len(client.ResourceIDs()) == 0 {
		return utils.NewStringSet(oauth2.LegacyResourceId)
	}

	return client.ResourceIDs()
}
