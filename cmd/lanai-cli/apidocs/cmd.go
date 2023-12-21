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

package apidocs

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/spf13/cobra"
)

var (
	logger = log.New("Build.APIDocs")
	Cmd    = &cobra.Command{
		Use:                "apidocs",
		Short:              "Utilities to work with OpenAPI Specs",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		//RunE:               Run,
	}
)

func init() {
	Cmd.AddCommand(ResolveCmd)
}


