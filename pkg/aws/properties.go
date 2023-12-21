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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	ConfigRootACM = "aws"
)

const (
	CredentialsTypeStatic CredentialsType = `static`
	CredentialsTypeSTS    CredentialsType = `sts`
)

type CredentialsType string

// Properties describes common config used to consume AWS services
type Properties struct {
	//Region for AWS client defaults to us-east-1
	Region string `json:"region"`
	//Endpoint for AWS client default empty can be used to override if consuming localstack
	Endpoint string `json:"endpoint"`
	//Credentials to be used to authenticate the AWS client
	Credentials Credentials `json:"credentials"`
}

// Credentials defines the type of credentials to use for AWS
type Credentials struct {
	//Type is one of static, env or sts.  Defaults to env.
	Type CredentialsType `json:"type"`

	//The following is only relevant to static credential
	//Id is the AWS_ACCESS_KEY_ID for the account
	Id string `json:"id"`
	//Secret is the AWS_SECRET_ACCESS_KEY
	Secret string `json:"secret"`

	//The follow is relevant to sts credentials (Used in EKS)
	//RoleARN defines role to be assumed by application if omitted environment variable AWS_ROLE_ARN will be used
	RoleARN string `json:"role-arn"`
	//TokenFile is the path to the STS OIDC token file if omitted environment variable AWS_WEB_IDENTITY_TOKEN_FILE will be used
	TokenFile string `json:"token-file"`
	//RoleSessionName username to associate with session e.g. service account
	RoleSessionName string `json:"role-session-name"`
}

func NewProperties() Properties {
	return Properties{
		Region: "us-east-1",
		Credentials: Credentials{
			Type: "env",
		},
	}
}

func BindAwsProperties(ctx *bootstrap.ApplicationContext) Properties {
	props := NewProperties()
	if err := ctx.Config().Bind(&props, ConfigRootACM); err != nil {
		panic(errors.Wrap(err, "failed to bind acm.Properties"))
	}
	return props
}
