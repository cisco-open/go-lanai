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

package webjars

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"embed"
	"github.com/spf13/cobra"
)

const (
	defaultWebjarContentPath = "META-INF/resources/webjars"
)

var (
	Cmd = &cobra.Command{
		Use:                "webjars",
		Short:              "Download Webjars and extract",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{
		Resources: []string{},
	}
)

type Arguments struct {
	GroupId    string   `flag:"group,g,required" desc:"Webjar's Group ID"`
	ArtifactId string   `flag:"artifact,a,required" desc:"Webjar's Artifact ID"`
	Version    string   `flag:"version,v,required" desc:"Webjar's Version"`
	Resources  []string `flag:"resources,r" desc:"Comma delimited list of additional resources from unpacked webjar. META-INF/resources/webjars is implicit"`
	MavenRepos []string `flag:"maven-repos" desc:"Comma delimited list of additional maven repository URLs"`
}

//go:embed pom.xml.tmpl
var TmplFS embed.FS

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

func Run(cmd *cobra.Command, _ []string) error {
	if e := generatePom(cmd.Context()); e != nil {
		return e
	}

	if e := executeMaven(cmd.Context()); e != nil {
		return e
	}
	return nil
}
