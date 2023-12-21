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

package acm

import (
	"context"
	awsclient "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
)

type ClientFactory interface {
	New(ctx context.Context, opts...func(opt *acm.Options)) (*acm.Client, error)
}

func NewClientFactory(loader awsclient.ConfigLoader) ClientFactory {
	return &acmFactory{
		configLoader: loader,
	}
}

type acmFactory struct {
	configLoader awsclient.ConfigLoader
}

func (f *acmFactory) New(ctx context.Context, opts...func(opt *acm.Options)) (*acm.Client, error) {
	cfg, e := f.configLoader.Load(ctx)
	if e != nil {
		return nil, e
	}
	return acm.NewFromConfig(cfg, opts...), nil
}
