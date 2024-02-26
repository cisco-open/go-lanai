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
    "crypto/rsa"
    "crypto/x509"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/idp"
    samlctx "github.com/cisco-open/go-lanai/pkg/security/saml"
    "github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
    "github.com/crewjam/saml"
    "github.com/crewjam/saml/samlsp"
    "github.com/golang-jwt/jwt/v4"
    dsig "github.com/russellhaering/goxmldsig"
    "net/http"
    "net/url"
)

type spOptionsHashable struct {
	URL          url.URL
	ACSPath      string
	MetadataPath string
	SLOPath      string
}

type SPOptions struct {
	spOptionsHashable
	Key               *rsa.PrivateKey
	Certificate       *x509.Certificate
	Intermediates     []*x509.Certificate
	AllowIDPInitiated bool
	SignRequest       bool
	ForceAuthn        bool
	NameIdFormat      string
}

type configurerSharedComponents struct {
	serviceProvider *saml.ServiceProvider
	tracker         samlsp.RequestTracker
	clientManager   *CacheableIdpClientManager
}

// samlConfigurer is a base implementation for both login and logout configurer.
// Many components for login and logout are shared
type samlConfigurer struct {
	properties     samlctx.SamlProperties
	idpManager     idp.IdentityProviderManager
	samlIdpManager samlctx.SamlIdentityProviderManager
	// Shared components, generated on demand
	components map[spOptionsHashable]*configurerSharedComponents
}

func newSamlConfigurer(properties samlctx.SamlProperties, idpManager idp.IdentityProviderManager) *samlConfigurer {
	return &samlConfigurer{
		properties:     properties,
		idpManager:     idpManager,
		samlIdpManager: idpManager.(samlctx.SamlIdentityProviderManager),
	}
}

// shared grab shared component based on issuer. Create if not exists.
// never returns nil
func (c *samlConfigurer) shared(hashable spOptionsHashable) *configurerSharedComponents {
	if c.components == nil {
		c.components = make(map[spOptionsHashable]*configurerSharedComponents)
	}

	shared, ok := c.components[hashable]
	if !ok {
		shared = &configurerSharedComponents{}
		c.components[hashable] = shared
	}
	return shared
}

func (c *samlConfigurer) getServiceProviderConfiguration(f *Feature) (opt SPOptions) {
	cert, err := cryptoutils.LoadCert(c.properties.CertificateFile)
	if err != nil {
		panic(security.NewInternalError("cannot load certificate from file", err))
	}
	if len(cert) > 1 {
		logger.Warnf("multiple certificate found, using first one")
	}

	key, err := cryptoutils.LoadPrivateKey(c.properties.KeyFile, c.properties.KeyPassword)
	if err != nil {
		panic(security.NewInternalError("cannot load private key from file", err))
	}
	rootURL, err := f.issuer.BuildUrl()
	if err != nil {
		panic(security.NewInternalError("cannot get issuer's base URL", err))
	}
	opts := SPOptions{
		spOptionsHashable: spOptionsHashable{
			URL:          *rootURL,
			ACSPath:      fmt.Sprintf("%s%s", rootURL.Path, f.acsPath),
			MetadataPath: fmt.Sprintf("%s%s", rootURL.Path, f.metadataPath),
			SLOPath:      fmt.Sprintf("%s%s", rootURL.Path, f.sloPath),
		},
		Key:          key,
		Certificate:  cert[0],
		SignRequest:  true,
		NameIdFormat: c.properties.NameIDFormat,
	}
	return opts
}

func (c *samlConfigurer) sharedServiceProvider(opts SPOptions) (ret saml.ServiceProvider) {
	if shared := c.shared(opts.spOptionsHashable); shared.serviceProvider != nil {
		return *shared.serviceProvider
	} else {
		defer func() {
			shared.serviceProvider = &ret
		}()
	}

	metadataURL := opts.URL.ResolveReference(&url.URL{Path: opts.MetadataPath})
	acsURL := opts.URL.ResolveReference(&url.URL{Path: opts.ACSPath})
	sloURL := opts.URL.ResolveReference(&url.URL{Path: opts.SLOPath})

	var forceAuthn *bool
	if opts.ForceAuthn {
		forceAuthn = &opts.ForceAuthn
	}
	signatureMethod := dsig.RSASHA1SignatureMethod
	if !opts.SignRequest {
		signatureMethod = ""
	}

	sp := saml.ServiceProvider{
		Key:               opts.Key,
		Certificate:       opts.Certificate,
		Intermediates:     opts.Intermediates,
		MetadataURL:       *metadataURL,
		AcsURL:            *acsURL,
		SloURL:            *sloURL,
		ForceAuthn:        forceAuthn,
		SignatureMethod:   signatureMethod,
		AllowIDPInitiated: opts.AllowIDPInitiated,
		AuthnNameIDFormat: saml.NameIDFormat(opts.NameIdFormat),
		LogoutBindings:    []string{saml.HTTPPostBinding},
	}
	return sp
}

func (c *samlConfigurer) sharedRequestTracker(opts SPOptions) (ret samlsp.RequestTracker) {
	if shared := c.shared(opts.spOptionsHashable); shared.tracker != nil {
		return shared.tracker
	} else {
		defer func() {
			shared.tracker = ret
		}()
	}

	codec := samlsp.JWTTrackedRequestCodec{
		SigningMethod: jwt.SigningMethodRS256,
		Audience:      opts.URL.String(),
		Issuer:        opts.URL.String(),
		MaxAge:        saml.MaxIssueDelay,
		Key:           opts.Key,
	}

	//we want to set sameSite to none, which requires scheme to be https
	//otherwise we fallback to default mode, which on modern browsers is lax.
	//cross site functionality is limited in lax mode. the cookie will only be
	//sent cross site within 2 minutes of its creation.
	//so with none + https, we make sure production work as expected. and the fallback
	//provides limited support for development environment.
	secure := opts.URL.Scheme == "https"
	sameSite := http.SameSiteDefaultMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	tracker := CookieRequestTracker{
		NamePrefix: "saml_",
		Codec:      codec,
		MaxAge:     saml.MaxIssueDelay,
		SameSite:   sameSite,
		Secure:     secure,
		Path:       opts.ACSPath,
	}
	return tracker
}

func (c *samlConfigurer) sharedClientManager(opts SPOptions) (ret *CacheableIdpClientManager) {
	if shared := c.shared(opts.spOptionsHashable); shared.clientManager != nil {
		return shared.clientManager
	} else {
		defer func() {
			shared.clientManager = ret
		}()
	}
	sp := c.sharedServiceProvider(opts)
	return NewCacheableIdpClientManager(sp)
}

func (c *samlConfigurer) effectiveSuccessHandler(f *Feature, ws security.WebSecurity) security.AuthenticationSuccessHandler {
	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		return security.NewAuthenticationSuccessHandler(globalHandler, f.successHandler)
	} else {
		return security.NewAuthenticationSuccessHandler(f.successHandler)
	}
}
