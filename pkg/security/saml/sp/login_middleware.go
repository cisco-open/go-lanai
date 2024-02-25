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

package sp

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/idp"
    "github.com/cisco-open/go-lanai/pkg/security/redirect"
    samlutils "github.com/cisco-open/go-lanai/pkg/security/saml/utils"
    netutil "github.com/cisco-open/go-lanai/pkg/utils/net"
    "github.com/crewjam/saml"
    "github.com/crewjam/saml/samlsp"
    "github.com/gin-gonic/gin"
    "net/http"
)

// SPLoginMiddleware
/**
A SAML service provider should be able to work with multiple identity providers.
Because the saml package assumes a service provider is configured with one idp only,
we use the internal field to store information about this service provider,
and we will create new saml.ServiceProvider struct for each new idp connection when its needed.
*/
type SPLoginMiddleware struct {
	SPMetadataMiddleware

	// list of bindings, can be saml.HTTPPostBinding or saml.HTTPRedirectBinding
	// order indicates preference
	requestTracker samlsp.RequestTracker

	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler

	fallbackEntryPoint security.AuthenticationEntryPoint
}

func NewLoginMiddleware(sp saml.ServiceProvider, tracker samlsp.RequestTracker,
	idpManager idp.IdentityProviderManager,
	clientManager *CacheableIdpClientManager,
	handler security.AuthenticationSuccessHandler, authenticator security.Authenticator,
	errorPath string) *SPLoginMiddleware {

	return &SPLoginMiddleware{
		SPMetadataMiddleware: SPMetadataMiddleware{
			internal:      sp,
			idpManager:    idpManager,
			clientManager: clientManager,
		},
		requestTracker:     tracker,
		successHandler:     handler,
		authenticator:      authenticator,
		fallbackEntryPoint: redirect.NewRedirectWithURL(errorPath),
	}
}

// MakeAuthenticationRequest Since we support multiple domains each with different IDP, the auth request specify which matching ACS should be
// used for IDP to call back.
func (sp *SPLoginMiddleware) MakeAuthenticationRequest(ctx context.Context, r *http.Request, w http.ResponseWriter) error {
	host := netutil.GetForwardedHostName(r)
	client, ok := sp.clientManager.GetClientByDomain(host)

	if !ok {
		logger.WithContext(ctx).Debugf("cannot find idp for domain %s", host)
		return security.NewExternalSamlAuthenticationError("cannot find idp for this domain")
	}

	location, binding := sp.resolveBinding(client.GetSSOBindingLocation)
	if location == "" {
		return security.NewExternalSamlAuthenticationError("idp does not have supported bindings.")
	}

	// Note: we only support post for result binding
	authReq, err := samlutils.NewFixedAuthenticationRequest(client, location, binding, saml.HTTPPostBinding)
	if err != nil {
		return security.NewExternalSamlAuthenticationError("cannot make auth request to binding location", err)
	}

	relayState, err := sp.requestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		return security.NewExternalSamlAuthenticationError("cannot track saml auth request", err)
	}

	switch binding {
	case saml.HTTPRedirectBinding:
		if e := sp.redirectBindingExecutor(authReq, relayState, client)(w, r); e != nil {
			return security.NewExternalSamlAuthenticationError("cannot make auth request with HTTP redirect binding", e)
		}
	case saml.HTTPPostBinding:
		if e := sp.postBindingExecutor(authReq, relayState)(w, r); e != nil {
			return security.NewExternalSamlAuthenticationError("cannot post auth request", e)
		}
	}
	return nil
}

// ACSHandlerFunc Assertion Consumer Service handler endpoint. IDP redirect to this endpoint with authentication response
func (sp *SPLoginMiddleware) ACSHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := saml.Response{}
		switch rs := samlutils.ParseSAMLObject(c, &resp); {
		case rs.Err != nil:
			sp.handleError(c, security.NewExternalSamlAuthenticationError(fmt.Errorf("cannot process ACS request: %v", rs.Err)))
			return
		case rs.Binding != saml.HTTPPostBinding:
			sp.handleError(c, security.NewExternalSamlAuthenticationError(fmt.Errorf("unsupported binding [%s]", rs.Binding)))
			return
		}

		r := c.Request
		client, ok := sp.clientManager.GetClientByEntityId(resp.Issuer.Value)
		if !ok {
			sp.handleError(c, security.NewExternalSamlAuthenticationError("cannot find idp metadata corresponding for assertion"))
			return
		}

		var possibleRequestIDs []string
		if sp.internal.AllowIDPInitiated {
			possibleRequestIDs = append(possibleRequestIDs, "")
		}

		trackedRequests := sp.requestTracker.GetTrackedRequests(r)
		for _, tr := range trackedRequests {
			possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
		}

		assertion, err := client.ParseResponse(r, possibleRequestIDs)
		if err != nil {
			logger.WithContext(c).Error("error processing assertion", "err", err)
			sp.handleError(c, security.NewExternalSamlAuthenticationError(err.Error(), err))
			return
		}

		candidate := &AssertionCandidate{
			Assertion: assertion,
		}
		auth, err := sp.authenticator.Authenticate(c, candidate)

		if err != nil {
			sp.handleError(c, security.NewExternalSamlAuthenticationError(err))
			return
		}

		before := security.Get(c)
		sp.handleSuccess(c, before, auth)
	}
}

func (sp *SPLoginMiddleware) Commence(c context.Context, r *http.Request, w http.ResponseWriter, _ error) {
	err := sp.MakeAuthenticationRequest(c, r, w)
	if err != nil {
		sp.fallbackEntryPoint.Commence(c, r, w, err)
	}
}

func (sp *SPLoginMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		security.MustSet(c, new)
	}
	sp.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	if c.Writer.Written() {
		c.Abort()
	}
}

func (sp *SPLoginMiddleware) handleError(c *gin.Context, err error) {
	if trackedRequestIndex := c.Request.Form.Get("RelayState"); trackedRequestIndex != "" {
		_ = sp.requestTracker.StopTrackingRequest(c.Writer, c.Request, trackedRequestIndex)
	}
	security.MustClear(c)
	_ = c.Error(err)
	c.Abort()
}
