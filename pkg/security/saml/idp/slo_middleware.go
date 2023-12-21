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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	samlutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	ctxKeySloRequest                = "slo.request"
	supportedLogoutResponseBindings = utils.NewStringSet(saml.HTTPPostBinding)
)

// SamlSingleLogoutMiddleware is a
// 1. logout.LogoutHandler
// 2. logout.ConditionalLogoutHandler
// 3. security.AuthenticationSuccessHandler
// 4. security.AuthenticationErrorHandler
// 5. security.AuthenticationEntryPoint
// focusing on validate SAML logout request and send back SAML LogoutResponse
type SamlSingleLogoutMiddleware struct {
	*MetadataMiddleware
	SamlErrorHandler
}

func NewSamlSingleLogoutMiddleware(metaMw *MetadataMiddleware) *SamlSingleLogoutMiddleware {
	return &SamlSingleLogoutMiddleware{
		MetadataMiddleware: metaMw,
	}
}

func (mw *SamlSingleLogoutMiddleware) Order() int {
	// always perform this first
	return order.Highest
}

func (mw *SamlSingleLogoutMiddleware) SLOCondition() web.RequestMatcher {
	return matcher.RequestHasForm(samlutils.HttpParamSAMLRequest)
}

// ShouldLogout is a logout.ConditionalLogoutHandler method that intercept SP initiated SAML request. Possible outcomes are:
// - no error returned if the logout is not SAML single logout (no SAMLRequest found)
// - no error returned if the logout is a valid SAMLLogoutRequest
// - ErrorSubTypeSamlSlo if SAMLLogoutRequest is found but invalid
func (mw *SamlSingleLogoutMiddleware) ShouldLogout(ctx context.Context, r *http.Request, _ http.ResponseWriter, _ security.Authentication) error {
	gc := web.GinContext(ctx)
	samlReq := mw.newSamlLogoutRequest(r)
	var req saml.LogoutRequest
	parsedReq := samlutils.ParseSAMLObject(gc, &req)
	switch {
	case parsedReq.Err != nil && len(parsedReq.Encoded) == 0:
		// not SAML request, ignore
		return nil
	case parsedReq.Err != nil:
		// Invalid SAML request, cancel with error
		mw.populateContext(gc, samlReq)
		return ErrorSamlSloRequester.WithMessage("unable to parse SAML SamlLogoutRequest: %v", parsedReq.Err)
	}

	samlReq.Binding = parsedReq.Binding
	samlReq.Request = &req
	samlReq.RequestBuffer = parsedReq.Decoded
	if e := mw.preProcessLogoutRequest(gc, samlReq); e != nil {
		return e
	}
	return nil
}

func (mw *SamlSingleLogoutMiddleware) HandleLogout(ctx context.Context, _ *http.Request, _ http.ResponseWriter, auth security.Authentication) error {
	req, ok := ctx.Value(ctxKeySloRequest).(*SamlLogoutRequest)
	if !ok {
		return nil
	}
	if e := mw.processLogoutRequest(ctx, req, auth); e != nil {
		return e
	}
	return mw.prepareSuccessSamlResponse(ctx, req)

}

func (mw *SamlSingleLogoutMiddleware) HandleAuthenticationSuccess(ctx context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	if req, ok := ctx.Value(ctxKeySloRequest).(*SamlLogoutRequest); ok {
		// Note, in case of success, SAML Response is prepared, we just send it
		if e := req.WriteResponse(rw); e != nil {
			msg := fmt.Sprintf("unable to send logout success response: %v", e)
			mw.HandleError(ctx, r, rw, NewSamlInternalError(msg, e))
		}
	}
	return
}

func (mw *SamlSingleLogoutMiddleware) HandleAuthenticationError(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	mw.HandleError(ctx, r, rw, err)
}

func (mw *SamlSingleLogoutMiddleware) Commence(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	mw.HandleError(ctx, r, rw, err)
}

func (mw *SamlSingleLogoutMiddleware) newSamlLogoutRequest(r *http.Request) *SamlLogoutRequest {
	return &SamlLogoutRequest{
		HTTPRequest: r,
		IDP:         mw.idp,
	}
}

func (mw *SamlSingleLogoutMiddleware) preProcessLogoutRequest(gc *gin.Context, req *SamlLogoutRequest) error {
	defer mw.populateContext(gc, req)

	// Note: we return Requester errors until we can determine the callback binding
	if req.Request.Issuer == nil || len(req.Request.Issuer.Value) == 0 {
		return ErrorSamlSloRequester.WithMessage("logout request missing Issuer")
	}

	// find the service provider metadata
	spId := req.Request.Issuer.Value
	spDetails, sp, e := mw.spMetadataManager.GetServiceProvider(spId)
	if e != nil {
		return ErrorSamlSloRequester.WithMessage("cannot find service provider metadata [%s]", spId)
	}

	// resolve SLO response endpoint
	req.SPMeta = sp
	if len(req.SPMeta.SPSSODescriptors) != 1 {
		return ErrorSamlSloRequester.WithMessage("expected exactly one SP SSO descriptor in SP metadata [%s]", spId)
	}

	spDesc := req.SPMeta.SPSSODescriptors[0]
	req.SPSSODescriptor = &spDesc
	if e := mw.determineSloEndpoint(gc, req); e != nil {
		return e
	}

	// validate request and relay state
	req.RelayState = req.HTTPRequest.FormValue(samlutils.HttpParamRelayState)
	if e := mw.validateLogoutRequest(gc, req, &spDetails); e != nil {
		return e
	}
	return nil
}

func (mw *SamlSingleLogoutMiddleware) determineSloEndpoint(_ *gin.Context, req *SamlLogoutRequest) error {
	// find first supported binding.
	// Note: we only support POST binding for now, because of crewjam/saml 0.4.8 limitation
	var found *saml.Endpoint
	for i := range req.SPSSODescriptor.SingleLogoutServices {
		ep := req.SPSSODescriptor.SingleLogoutServices[i]
		if supportedLogoutResponseBindings.Has(ep.Binding) && len(ep.Location) != 0 {
			found = &ep
			break
		}
	}
	if found == nil {
		return ErrorSamlSloRequester.WithMessage("SAML SLO unable to find supported response bindings from SP. Should be one of %v", supportedLogoutResponseBindings.Values())
	} else if len(found.ResponseLocation) == 0 {
		found.ResponseLocation = found.Location
	}
	req.Callback = found
	return nil
}

func (mw *SamlSingleLogoutMiddleware) validateLogoutRequest(_ *gin.Context, req *SamlLogoutRequest, spDetails *SamlSpDetails) error {
	if !spDetails.SkipAuthRequestSignatureVerification {
		if e := req.VerifySignature(); e != nil {
			return ErrorSamlSloResponder.WithMessage(e.Error())
		}
	}
	return req.Validate()
}

func (mw *SamlSingleLogoutMiddleware) processLogoutRequest(_ context.Context, req *SamlLogoutRequest, auth security.Authentication) error {
	if auth.State() < security.StatePrincipalKnown {
		// no additional check are needed
		return nil
	}

	nameID := req.Request.NameID
	switch saml.NameIDFormat(nameID.Format) {
	case saml.EmailAddressNameIDFormat:
		fallthrough
	case saml.TransientNameIDFormat:
		fallthrough
	case saml.PersistentNameIDFormat:
		return ErrorSamlSloResponder.WithMessage("unsupported NameID format [%s]", nameID.Format)
	default:
		// we assume it's username
		if username, e := security.GetUsername(auth); e != nil || username != nameID.Value {
			logger.Warnf("SAML SLO rejected: NameID doesn't match current session. Caused by: %v", e)
			return ErrorSamlSloResponder.WithMessage("NameID doesn't match current session")
		}
	}
	return nil
}

func (mw *SamlSingleLogoutMiddleware) populateContext(gc *gin.Context, req *SamlLogoutRequest) {
	gc.Set(ctxKeySloRequest, req)
}

func (mw *SamlSingleLogoutMiddleware) prepareSuccessSamlResponse(ctx context.Context, req *SamlLogoutRequest) error {
	resp, e := MakeLogoutResponse(req, saml.StatusSuccess, "")
	if e != nil {
		logger.WithContext(ctx).Warnf("SAML SLO unable to sign logout response")
		return security.NewAuthenticationWarningError("Unable to send SAML Logout Response")
	}
	req.Response = resp
	return nil
}
