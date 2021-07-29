package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/openid"
	"net/http"
)

// WellKnownEndpoint provide "/.well-known/**" HTTP endpoints
type WellKnownEndpoint struct {
	issuer     security.Issuer
	extra      map[string]interface{}
}

func NewWellKnownEndpoint(issuer security.Issuer, idpManager idp.IdentityProviderManager, extra map[string]interface{}) *WellKnownEndpoint {
	if extra == nil {
		extra = map[string]interface{}{}
	}
	extra[openid.OPMetaExtraSourceIDPManager] = idpManager
	return &WellKnownEndpoint{
		issuer: issuer,
		extra:  extra,
	}
}

// OpenIDConfig should mapped to GET /.well-known/openid-configuration
func (ep *WellKnownEndpoint) OpenIDConfig(ctx context.Context, _ *http.Request) (resp *openid.OPMetadata, err error) {
	c := openid.OPMetadata{MapClaims: oauth2.MapClaims{}}
	e := claims.Populate(ctx, &c,
		claims.WithSpecs(openid.OPMetadataBasicSpecs, openid.OPMetadataOptionalSpecs),
		claims.WithIssuer(ep.issuer),
		claims.WithExtraSource(ep.extra),
	)
	if e != nil {
		return nil, e
	}
	return &c, nil
}
