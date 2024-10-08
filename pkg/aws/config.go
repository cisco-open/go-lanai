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

package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"os"
)

const errTmpl = `invalid AWS configuration: %v`

type ConfigLoader interface {
	Load(ctx context.Context, opts ...config.LoadOptionsFunc) (aws.Config, error)
}

// ConfigOverrideFunc used to override loaded aws.Config
type ConfigOverrideFunc func(cfg *aws.Config)

func NewConfigLoader(p Properties, customizers []config.LoadOptionsFunc, overrides []ConfigOverrideFunc) ConfigLoader {
	return &PropertiesBasedConfigLoader{
		Properties:  &p,
		Customizers: customizers,
		Overrides:   overrides,
	}
}

type PropertiesBasedConfigLoader struct {
	Properties  *Properties
	Customizers []config.LoadOptionsFunc
	Overrides   []ConfigOverrideFunc
}

func (l *PropertiesBasedConfigLoader) Load(ctx context.Context, opts ...config.LoadOptionsFunc) (cfg aws.Config, err error) {
	extraOpts := append(l.Customizers, opts...)
	opts = append([]config.LoadOptionsFunc{
		WithBasicProperties(l.Properties),
		WithCredentialsProperties(ctx, l.Properties, extraOpts...)},
		extraOpts...,
	)
	cfg, err = LoadConfig(ctx, opts...)
	if err != nil {
		return
	}

	overrides := append([]ConfigOverrideFunc{OverrideConfigWithProperties(l.Properties)}, l.Overrides...)
	for _, fn := range overrides {
		fn(&cfg)
	}
	return
}

func LoadConfig(ctx context.Context, opts ...config.LoadOptionsFunc) (aws.Config, error) {
	unnamedOpts := make([]func(*config.LoadOptions) error, len(opts))
	for i := range opts {
		unnamedOpts[i] = opts[i]
	}
	return config.LoadDefaultConfig(ctx, unnamedOpts...)
}

func WithBasicProperties(p *Properties) config.LoadOptionsFunc {
	return func(opt *config.LoadOptions) error {
		if len(p.Region) == 0 {
			return fmt.Errorf(errTmpl, "Region is not set")
		}
		opt.Region = p.Region
		//Note: Endpoint is set in OverrideConfigWithProperties
		return nil
	}
}

func WithCredentialsProperties(ctx context.Context, p *Properties, globalOpts ...config.LoadOptionsFunc) config.LoadOptionsFunc {
	return func(opt *config.LoadOptions) error {
		switch p.Credentials.Type {
		case CredentialsTypeStatic:
			opt.Credentials = credentials.NewStaticCredentialsProvider(p.Credentials.Id, p.Credentials.Secret, "static_auth")
		case CredentialsTypeSTS:
			var e error
			if opt.Credentials, e = NewStsCredentialsProvider(ctx, p, globalOpts...); e != nil {
				return fmt.Errorf(errTmpl, e)
			}
		default:
			opt.Credentials = NewEnvCredentialsProvider()
		}
		return nil
	}
}

// OverrideConfigWithProperties overrides given aws.Config with properties
func OverrideConfigWithProperties(props *Properties) ConfigOverrideFunc {
	return func(cfg *aws.Config) {
		// Note: At v1.27.27, aws.EndpointResolverWithOptionsFunc is deprecated. Each client would have its own EndpointResolver.
		// All we need is to set the BaseEndpoint. But this property cannot be set via config.LoadOptionsFunc.
		// So we set it via ConfigOverrideFunc
		if len(props.Endpoint) != 0 {
			cfg.BaseEndpoint = aws.String(props.Endpoint)
		}
	}
}

func NewStsCredentialsProvider(ctx context.Context, p *Properties, opts ...config.LoadOptionsFunc) (aws.CredentialsProvider, error) {
	tokenPath := p.Credentials.TokenFile
	if tokenPath == "" {
		tokenPath = os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	}

	roleArn := p.Credentials.RoleARN
	if roleArn == "" {
		roleArn = os.Getenv("AWS_ROLE_ARN")
	}

	// prepare config for STS
	opts = append([]config.LoadOptionsFunc{WithBasicProperties(p)}, opts...)
	cfg, e := LoadConfig(ctx, opts...)
	if e != nil {
		return nil, fmt.Errorf(`unable to prepare for STS credentials`)
	}

	overrides := []ConfigOverrideFunc{OverrideConfigWithProperties(p)}
	for _, fn := range overrides {
		fn(&cfg)
	}

	// create provider
	client := sts.NewFromConfig(cfg)
	provider := stscreds.NewWebIdentityRoleProvider(client, roleArn, stscreds.IdentityTokenFile(tokenPath), func(opts *stscreds.WebIdentityRoleOptions) {
		opts.RoleSessionName = p.Credentials.RoleSessionName
	})
	return provider, nil
}

func NewEnvCredentialsProvider() aws.CredentialsProvider {
	return aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		id := os.Getenv("AWS_ACCESS_KEY_ID")
		if id == "" {
			id = os.Getenv("AWS_ACCESS_KEY")
		}

		secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if secret == "" {
			secret = os.Getenv("AWS_SECRET_KEY")
		}
		return aws.Credentials{
			AccessKeyID:     id,
			SecretAccessKey: secret,
			SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		}, nil
	})
}
