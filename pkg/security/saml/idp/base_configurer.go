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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"fmt"
	dsig "github.com/russellhaering/goxmldsig"
	"net/url"
)

type samlConfigurer struct {
	properties      samlctx.SamlProperties
	samlClientStore samlctx.SamlClientStore
}

func (c *samlConfigurer) getIdentityProviderConfiguration(f *Feature) *Options {
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

	var signingMethod string
	switch f.signingMethod {
	case dsig.RSASHA1SignatureMethod:
		fallthrough
	case dsig.RSASHA256SignatureMethod:
		fallthrough
	case dsig.RSASHA512SignatureMethod:
		signingMethod = f.signingMethod
	default:
		signingMethod = dsig.RSASHA1SignatureMethod
	}

	return &Options{
		Key:  key,
		Cert: cert[0],
		//usually this is the metadata url, but to keep consistent with existing implementation, we just use the context path
		EntityIdUrl: *rootURL,
		SsoUrl: *rootURL.ResolveReference(&url.URL{
			Path:     fmt.Sprintf("%s%s", rootURL.Path, f.ssoLocation.Path),
			RawQuery: f.ssoLocation.RawQuery,
		}),
		SloUrl: *rootURL.ResolveReference(&url.URL{
			Path: fmt.Sprintf("%s%s", rootURL.Path, f.logoutUrl),
		}),
		SigningMethod:          signingMethod,
		serviceProviderManager: c.samlClientStore,
	}
}

func (c *samlConfigurer) metadataMiddleware(f *Feature) *MetadataMiddleware {
	opts := c.getIdentityProviderConfiguration(f)
	return NewMetadataMiddleware(opts, c.samlClientStore)
}
