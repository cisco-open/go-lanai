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

package security

import (
	"fmt"
	"net/url"
	pathutils "path"
	"strings"
)

type UrlBuilderOptions func(opt *UrlBuilderOption)

type UrlBuilderOption struct {
	FQDN string
	Path string
}

type Issuer interface {
	Protocol() string
	Domain() string
	Port() int
	ContextPath() string
	IsSecured() bool

	// Identifier is the unique identifier of the deployed auth server
	// Typeical implementation is to use base url of issuer's domain.
	Identifier() string

	// LevelOfAssurance construct level-of-assurance string with given string
	// level-of-assurance represent how confident the auth issuer is about user's identity
	// ref: https://developer.mobileconnect.io/level-of-assurance
	LevelOfAssurance(level int) string

	// BuildUrl build a URL with given url builder options
	// Implementation specs:
	// 	1. if UrlBuilderOption.FQDN is not specified, Issuer.Domain() should be used
	//  2. if UrlBuilderOption.FQDN is not a subdomain of Issuer.Domain(), an error should be returned
	//  3. should assume UrlBuilderOption.Path doesn't includes Issuer.ContextPath and the generated URL always
	//	   include Issuer.ContextPath
	//  4. if UrlBuilderOption.Path is not specified, the generated URL could be used as a base URL
	//	5. BuildUrl should not returns error when no options provided
	BuildUrl(...UrlBuilderOptions) (*url.URL, error)
}

/***************************
	Default Impl.
 ***************************/
type DefaultIssuerDetails struct {
	Protocol    string
	Domain      string
	Port        int
	ContextPath string
	IncludePort bool
}

type DefaultIssuer struct {
	DefaultIssuerDetails
}

func NewIssuer(opts ...func(*DefaultIssuerDetails)) *DefaultIssuer {
	opt := DefaultIssuerDetails{

	}
	for _, f := range opts {
		f(&opt)
	}
	return &DefaultIssuer{
		DefaultIssuerDetails: opt,
	}
}

func (i DefaultIssuer) Protocol() string {
	return i.DefaultIssuerDetails.Protocol
}

func (i DefaultIssuer) Domain() string {
	return i.DefaultIssuerDetails.Domain
}

func (i DefaultIssuer) Port() int {
	return i.DefaultIssuerDetails.Port
}

func (i DefaultIssuer) ContextPath() string {
	return i.DefaultIssuerDetails.ContextPath
}

func (i DefaultIssuer) IsSecured() bool {
	return strings.ToLower(i.DefaultIssuerDetails.Protocol) == "https"
}

func (i DefaultIssuer) Identifier() string {
	id, _ := i.BuildUrl()
	return id.String()
}

func (i DefaultIssuer) LevelOfAssurance(level int) string {
	path := fmt.Sprintf("/loa-%d", level)
	loa, _ := i.BuildUrl(func(opt *UrlBuilderOption) {
		opt.Path = path
	})
	return loa.String()
}

func (i DefaultIssuer) BuildUrl(options ...UrlBuilderOptions) (*url.URL, error) {
	opt := UrlBuilderOption{}
	for _, f := range options {
		f(&opt)
	}
	if opt.FQDN == "" {
		opt.FQDN = i.DefaultIssuerDetails.Domain
	}

	if strings.HasSuffix(opt.FQDN, i.DefaultIssuerDetails.Domain) && strings.HasPrefix(opt.FQDN, ".") {
		return nil, fmt.Errorf("invalid subdomain %s", opt.FQDN)
	}

	ret := &url.URL{}
	ret.Scheme = i.DefaultIssuerDetails.Protocol
	ret.Host = opt.FQDN
	if i.DefaultIssuerDetails.IncludePort {
		ret.Host = fmt.Sprintf("%s:%d", ret.Host, i.DefaultIssuerDetails.Port)
	}

	ret.Path = i.DefaultIssuerDetails.ContextPath
	if opt.Path != "" {
		path := pathutils.Join(ret.Path, opt.Path)
		ret = ret.ResolveReference(&url.URL{Path: path})
	}

	return ret, nil
}
