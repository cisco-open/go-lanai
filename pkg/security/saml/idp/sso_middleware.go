// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package samlidp

import (
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    samlctx "github.com/cisco-open/go-lanai/pkg/security/saml"
    "github.com/cisco-open/go-lanai/pkg/tenancy"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/crewjam/saml"
    "github.com/gin-gonic/gin"
)

type SamlAuthorizeEndpointMiddleware struct {
	*MetadataMiddleware
	accountStore       security.AccountStore
	attributeGenerator AttributeGenerator
}

func NewSamlAuthorizeEndpointMiddleware(metaMw *MetadataMiddleware,
	accountStore security.AccountStore,
	attributeGenerator AttributeGenerator) *SamlAuthorizeEndpointMiddleware {

	mw := &SamlAuthorizeEndpointMiddleware{
		MetadataMiddleware: metaMw,
		accountStore:       accountStore,
		attributeGenerator: attributeGenerator,
	}

	return mw
}

func (mw *SamlAuthorizeEndpointMiddleware) AuthorizeHandlerFunc(condition web.RequestMatcher) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if matches, err := condition.MatchesWithContext(ctx, ctx.Request); !matches || err != nil {
			return
		}

		var req *saml.IdpAuthnRequest
		var err error

		idpInitiatedMatcher := matcher.RequestWithForm("idp_init", "true")
		isIdpInit, _ := idpInitiatedMatcher.Matches(ctx.Request)
		if isIdpInit {
			entityId := ctx.Request.Form.Get("entity_id")
			if entityId == "" {
				mw.handleError(ctx, nil, NewSamlInternalError("error start idp initiated sso, no SP entity id provided"))
				return
			}

			req = &saml.IdpAuthnRequest{
				Request: saml.AuthnRequest{
					Issuer: &saml.Issuer{
						Value: entityId,
					},
					IssueInstant: saml.TimeNow(),
				},
				IDP: mw.idp,
				Now: saml.TimeNow(),
			}
		} else {
			req, err = saml.NewIdpAuthnRequest(mw.idp, ctx.Request)
			if err != nil {
				mw.handleError(ctx, nil, NewSamlInternalError("error decoding authentication request", err))
				return
			}
			if err = UnmarshalRequest(req); err != nil {
				mw.handleError(ctx, nil, err)
				return
			}
		}

		authentication := security.Get(ctx)
		//sanity check
		if authentication == nil {
			mw.handleError(ctx, nil, NewSamlInternalError("no authentication found", err))
			return
		}

		//sanity check
		if authentication.State() < security.StateAuthenticated {
			mw.handleError(ctx, nil, NewSamlInternalError("session is not authenticated"))
			return
		}

		serviceProviderID := req.Request.Issuer.Value

		// find the service provider metadata
		spDetails, spMetadata, err := mw.spMetadataManager.GetServiceProvider(serviceProviderID)
		if err != nil {
			mw.handleError(ctx, nil, NewSamlInternalError("cannot find service provider metadata"))
			return
		}
		if len(spMetadata.SPSSODescriptors) != 1 {
			mw.handleError(ctx, nil, NewSamlInternalError("expected exactly one SP SSO descriptor in SP metadata"))
			return
		}
		req.ServiceProviderMetadata = spMetadata
		req.SPSSODescriptor = &spMetadata.SPSSODescriptors[0]

		// Check that the ACS URL matches an ACS endpoint in the SP metadata.
		// After this point, we have the endpoint to send back responses whether it's success or false
		if err = DetermineACSEndpoint(req); err != nil {
			mw.handleError(ctx, nil, err)
			return
		}

		if !isIdpInit {
			if err = ValidateAuthnRequest(req, spDetails, spMetadata); err != nil {
				mw.handleError(ctx, req, err)
				return
			}
		}

		//check tenancy
		client, err := mw.samlClientStore.GetSamlClientByEntityId(ctx.Request.Context(), serviceProviderID)
		if err != nil { //we shouldn't get an error here because we already have the SP's metadata.
			//if an error does occur, it means there's a programming error
			mw.handleError(ctx, nil, NewSamlInternalError("saml client not found", err))
			return
		}
		err = mw.validateTenantRestriction(ctx, client, authentication)
		if err != nil {
			mw.handleError(ctx, req, err)
			return
		}

		if err = MakeAssertion(ctx, req, authentication, mw.attributeGenerator); err != nil {
			mw.handleError(ctx, req, err)
			return
		}

		if err = MakeAssertionEl(req, spDetails.SkipAssertionEncryption); err != nil {
			mw.handleError(ctx, req, err)
			return
		}

		if err = req.WriteResponse(ctx.Writer); err != nil {
			mw.handleError(ctx, nil, NewSamlInternalError("error writing saml response", err))
			return
		} else {
			//abort the rest of the handlers because we have already written the response successfully
			ctx.Abort()
		}
	}
}

func (mw *SamlAuthorizeEndpointMiddleware) handleError(c *gin.Context, authRequest *saml.IdpAuthnRequest, err error) {
	if !errors.Is(err, security.ErrorTypeSaml) {
		err = NewSamlInternalError("saml sso internal error", err)
	}

	if authRequest != nil {
		c.Set(CtxKeySamlAuthnRequest, authRequest)
	}

	_ = c.Error(err)
	c.Abort()
}

func (mw *SamlAuthorizeEndpointMiddleware) validateTenantRestriction(ctx context.Context, client samlctx.SamlClient, auth security.Authentication) error {
	tenantRestriction := client.GetTenantRestrictions()

	if len(tenantRestriction) == 0 {
		return nil
	}

	username, e := security.GetUsername(auth)
	if e != nil {
		return NewSamlInternalError("cannot validate tenancy restriction due to unknown username", e)
	}

	acct, e := mw.accountStore.LoadAccountByUsername(ctx, username)
	if e != nil {
		return NewSamlInternalError("cannot validate tenancy restriction due to error fetching account", e)
	}

	acctTenancy, ok := acct.(security.AccountTenancy)
	if !ok {
		return NewSamlInternalError(fmt.Sprintf("cannot validate tenancy restriction due to unsupported account implementation: %T", acct))
	}

	userAccessibleTenants := utils.NewStringSet(acctTenancy.DesignatedTenantIds()...)

	if userAccessibleTenants.Has(security.SpecialTenantIdWildcard) {
		return nil
	}

	switch tenantRestrictionType := client.GetTenantRestrictionType(); tenantRestrictionType {
	case TenantRestrictionTypeAny:
		allowed := false
		for t := range tenantRestriction {
			if tenancy.AnyHasDescendant(ctx, userAccessibleTenants, t) {
				allowed = true
				break
			}
		}
		if !allowed {
			return NewSamlInternalError("client is restricted to tenants which the authenticated user does not have access to")
		}
	default: //default to TenantRestrictionTypeAll
		for t := range tenantRestriction {
			if !tenancy.AnyHasDescendant(ctx, userAccessibleTenants, t) {
				return NewSamlInternalError("client is restricted to tenants which the authenticated user does not have access to")
			}
		}
	}

	return nil
}
