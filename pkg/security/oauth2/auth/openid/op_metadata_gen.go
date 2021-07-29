package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
)

/************************
	OPMetadata Specs
 ************************/

const (
	OPMetaExtraSourceIDPManager = "idpManager"
)

var errorOPMetaClaimNotAvailable = fmt.Errorf("claim N/A")

type constantClaimSpec struct {
	val interface{}
}

func (s constantClaimSpec) Calculate(_ context.Context, _ *claims.FactoryOption) (v interface{}, err error) {
	return s.val, nil
}

func (s constantClaimSpec) Required(_ context.Context, _ *claims.FactoryOption) bool {
	return false
}

type opMetaClaimSpec struct {
	fn claims.ClaimFactoryFunc
}

func (s opMetaClaimSpec) Calculate(ctx context.Context, opt *claims.FactoryOption) (v interface{}, err error) {
	return s.fn(ctx, opt)
}

func (s opMetaClaimSpec) Required(_ context.Context, _ *claims.FactoryOption) bool {
	return false
}

func opMetaFixedSet(strings ...string) claims.ClaimSpec {
	return constantClaimSpec{
		val: utils.NewStringSet(strings...),
	}
}

func opMetaFixedBool(v bool) claims.ClaimSpec {
	return constantClaimSpec{
		val: v,
	}
}

func opMetaAcrValues(acrLevels ...int) claims.ClaimSpec {
	return opMetaClaimSpec{
		fn: func(_ context.Context, opt *claims.FactoryOption) (v interface{}, err error) {
			if opt.Issuer == nil {
				return nil, errorOPMetaClaimNotAvailable
			}
			values := utils.NewStringSet()
			for _, lvl := range acrLevels {
				values.Add(opt.Issuer.LevelOfAssurance(lvl))
			}
			return values, nil
		},
	}
}

func opMetaEndpoint(epName string) claims.ClaimSpec {
	return opMetaClaimSpec{
		fn: func(ctx context.Context, opt *claims.FactoryOption) (v interface{}, err error) {
			if opt.ExtraSource == nil || opt.Issuer == nil {
				return nil, errorOPMetaClaimNotAvailable
			}

			relative, ok := opt.ExtraSource[epName].(string)
			if !ok {
				return nil, errorOPMetaClaimNotAvailable
			}

			// figure out domain
			idpMgt, ok := opt.ExtraSource[OPMetaExtraSourceIDPManager].(idp.IdentityProviderManager)
			if !ok {
				return nil, errorOPMetaClaimNotAvailable
			}
			var domain string
			for _, flow := range []idp.AuthenticationFlow{idp.InternalIdpForm, idp.ExternalIdpSAML} {
				idps := idpMgt.GetIdentityProvidersWithFlow(ctx, flow)
				if len(idps) != 0 {
					domain = idps[0].Domain()
					break
				}
			}

			uri, e := opt.Issuer.BuildUrl(func(opt *security.UrlBuilderOption) {
				opt.FQDN = domain
				opt.Path = relative
			})
			if e != nil {
				return nil, e
			}
			return uri.String(), nil
		},
	}
}