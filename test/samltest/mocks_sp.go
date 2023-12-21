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

package samltest

import (
	"crypto/rsa"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
	"net/url"
)

type SPMockOptions func(opt *SPMockOption)
type SPMockOption struct {
	Properties SPProperties
	IDP        *saml.IdentityProvider
}

// SPWithPropertiesPrefix returns a SP mock option that bind properties from application config and with given prefix
func SPWithPropertiesPrefix(appCfg bootstrap.ApplicationConfig, prefix string) SPMockOptions {
	return func(opt *SPMockOption) {
		if e := appCfg.Bind(&opt.Properties, prefix); e != nil {
			panic(e)
		}
	}
}

// SPWithIDP returns a SP mock option that set given IDP
func SPWithIDP(idp *saml.IdentityProvider) SPMockOptions {
	return func(opt *SPMockOption) {
		opt.IDP = idp
	}
}

// MustNewMockedSP similar to NewMockedSP, panic instead of returning error
func MustNewMockedSP(opts ...SPMockOptions) *saml.ServiceProvider {
	sp, e := NewMockedSP(opts...)
	if e != nil {
		panic(e)
	}
	return sp
}

// NewMockedSP create a mocked SP with given SPMockOptions.
// Returns error if any mocked value are incorrect. e.g. file not exists
func NewMockedSP(opts ...SPMockOptions) (*saml.ServiceProvider, error) {
	defaultEntityID, _ := DefaultIssuer.BuildUrl()
	opt := SPMockOption{
		Properties: SPProperties{
			ProviderProperties: ProviderProperties{
				EntityID: defaultEntityID.String(),
			},
			ACSPath: "/acs",
			SLOPath: "/slo",
		},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	var e error
	var spCerts []*x509.Certificate
	var privKey *rsa.PrivateKey
	var acsUrl, sloUrl *url.URL

	if spCerts, e = cryptoutils.LoadCert(opt.Properties.CertsSource); e != nil {
		return nil, e
	}
	if privKey, e = cryptoutils.LoadPrivateKey(opt.Properties.PrivateKeySource, ""); e != nil && len(opt.Properties.PrivateKeySource) != 0 {
		return nil, e
	}
	if acsUrl, e = resolveAbsUrl(opt.Properties.EntityID, opt.Properties.ACSPath); e != nil {
		return nil, e
	}
	if sloUrl, e = resolveAbsUrl(opt.Properties.EntityID, opt.Properties.SLOPath); e != nil && len(opt.Properties.SLOPath) != 0 {
		return nil, e
	}

	sp := saml.ServiceProvider{
		EntityID:          opt.Properties.EntityID,
		Key:               privKey,
		Certificate:       spCerts[0],
		AcsURL:            *acsUrl,
		SloURL:            *sloUrl,
		SignatureMethod:   dsig.RSASHA256SignatureMethod,
		AllowIDPInitiated: true,
		AuthnNameIDFormat: saml.UnspecifiedNameIDFormat,
		LogoutBindings:    []string{saml.HTTPPostBinding},
	}

	switch {
	case opt.IDP != nil:
		sp.IDPMetadata = opt.IDP.Metadata()
	case opt.Properties.IDP != nil:
		idp, e := NewMockedIDP(func(idpopt *IDPMockOption) { idpopt.Properties = *opt.Properties.IDP })
		if e == nil {
			sp.IDPMetadata = idp.Metadata()
		}
	}
	return &sp, nil
}

func resolveAbsUrl(baseUrl, toResolveUrl string) (*url.URL, error) {
	base, e := url.Parse(baseUrl)
	if e != nil {
		return nil, e
	}
	toResolve, e := url.Parse(toResolveUrl)
	if e != nil {
		return nil, e
	}

	return base.ResolveReference(&url.URL{RawPath: toResolve.RawPath, RawQuery: toResolve.RawQuery, Fragment: toResolve.Fragment}), nil

}
