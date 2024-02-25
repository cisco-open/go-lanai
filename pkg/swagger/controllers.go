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

package swagger

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/assets"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "io/fs"
    "net/http"
    "strings"
)

type UiConfiguration struct {
	ApisSorter               string   `json:"apisSorter"`
	DeepLinking              bool     `json:"deepLinking"`
	DefaultModelExpandDepth  int      `json:"defaultModelExpandDepth"`
	DefaultModelRendering    string   `json:"defaultModelRendering"`
	DefaultModelsExpandDepth int      `json:"defaultModelsExpandDepth"`
	DisplayOperationId       bool     `json:"displayOperationId"`
	DisplayRequestDuration   bool     `json:"displayRequestDuration"`
	DocExpansion             string   `json:"docExpansion"`
	Filter                   bool     `json:"filter"`
	JsonEditor               bool     `json:"jsonEditor"`
	OperationsSorter         string   `json:"operationsSorter"`
	ShowExtensions           bool     `json:"showExtensions"`
	ShowRequestHeaders       bool     `json:"showRequestHeaders"`
	SupportedSubmitMethods   []string `json:"supportedSubmitMethods"`
	TagsSorter               string   `json:"tagsSorter"`
	ValidatorUrl             string   `json:"validatorUrl"`
	Title                    string   `json:"title"`
}

type SsoConfiguration struct {
	Enabled          bool        `json:"enabled"`
	AuthorizeUrl     string      `json:"authorizeUrl"`
	ClientId         string      `json:"clientId"`
	ClientSecret     string      `json:"clientSecret"`
	TokenUrl         string      `json:"tokenUrl"`
	AdditionalParams []ParamMeta `json:"additionalParameters"`
}

type ParamMeta struct {
	Name               string `json:"name"`
	DisplayName        string `json:"displayName"`
	CandidateSourceUrl string `json:"candidateSourceUrl"`
	CandidateJsonPath  string `json:"candidateJsonPath"`
}

type Resource struct {
	Name           string `json:"name"`
	Location       string `json:"location"`
	Url            string `json:"url"`
	SwaggerVersion string `json:"swaggerVersion"`
}

//goland:noinspection GoNameStartsWithPackageName
type SwaggerController struct {
	properties        *SwaggerProperties
	buildInfoResolver bootstrap.BuildInfoResolver
	docLoader         *OASDocLoader
}

func NewSwaggerController(props SwaggerProperties, resolver bootstrap.BuildInfoResolver) *SwaggerController {
	return newSwaggerController(props, resolver)
}

func newSwaggerController(props SwaggerProperties, resolver bootstrap.BuildInfoResolver, searchFS ...fs.FS) *SwaggerController {
	return &SwaggerController{
		properties:        &props,
		buildInfoResolver: resolver,
		docLoader:         newOASDocLoader(props.Spec, searchFS...),
	}
}

func (c *SwaggerController) Mappings() []web.Mapping {
	return []web.Mapping{
		assets.New("/swagger/static", "generated/"),
		web.NewSimpleMapping("swagger-ui", "", "/swagger", http.MethodGet, nil, c.swagger),
		rest.New("swagger-configuration-ui").Get("/swagger-resources/configuration/ui").EndpointFunc(c.configurationUi).Build(),
		rest.New("swagger-configuration-security").Get("/swagger-resources/configuration/security").EndpointFunc(c.configurationSecurity).Build(),
		rest.New("swagger-configuration-sso").Get("/swagger-resources/configuration/security/sso").EndpointFunc(c.configurationSso).Build(),
		rest.New("swagger-resources").Get("/swagger-resources").EndpointFunc(c.resources).Build(),
		web.NewSimpleMapping("swagger-sso-redirect", "", "swagger-sso-redirect.html", http.MethodGet, nil, c.swaggerRedirect),
		web.NewSimpleMapping("swagger-spec", "", "/v2/api-docs", http.MethodGet, nil, c.oas2Doc),
		web.NewSimpleMapping("oas3-spec", "", "/v3/api-docs", http.MethodGet, nil, c.oas3Doc),
	}
}

func (c *SwaggerController) configurationUi(_ context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = UiConfiguration{
		DeepLinking:              true,
		DisplayOperationId:       false,
		DefaultModelsExpandDepth: 1,
		DefaultModelExpandDepth:  1,
		DefaultModelRendering:    "example",
		DisplayRequestDuration:   false,
		DocExpansion:             "none",
		Filter:                   false,
		OperationsSorter:         "alpha",
		ShowExtensions:           false,
		TagsSorter:               "alpha",
		ValidatorUrl:             "",
		SupportedSubmitMethods:   []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"},
		Title:                    c.properties.UI.Title,
	}
	return
}

func (c *SwaggerController) swagger(w http.ResponseWriter, r *http.Request) {
	fs := http.FS(Content)
	file, err := fs.Open("generated/swagger-ui.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func (c *SwaggerController) configurationSecurity(_ context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = struct{}{}
	return
}

func (c *SwaggerController) configurationSso(_ context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	var params []ParamMeta

	for _, v := range c.properties.Security.Sso.AdditionalParams {
		params = append(params, ParamMeta{
			Name:               v.Name,
			DisplayName:        v.DisplayName,
			CandidateSourceUrl: v.CandidateSourceUrl,
			CandidateJsonPath:  v.CandidateJsonPath,
		})
	}

	response = SsoConfiguration{
		Enabled:          c.properties.Security.Sso.ClientId != "",
		TokenUrl:         fmt.Sprintf("%s%s", c.properties.Security.Sso.BaseUrl, c.properties.Security.Sso.TokenPath),
		AuthorizeUrl:     fmt.Sprintf("%s%s", c.properties.Security.Sso.BaseUrl, c.properties.Security.Sso.AuthorizePath),
		ClientId:         c.properties.Security.Sso.ClientId,
		ClientSecret:     c.properties.Security.Sso.ClientSecret,
		AdditionalParams: params,
	}
	return
}

func (c *SwaggerController) resources(_ context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = []Resource{
		{
			Name:           "platform",
			Url:            "/v3/api-docs?group=platform",
			SwaggerVersion: "3.0",
			Location:       "/v3/api-docs?group=platform",
		},
	}
	return
}

func (c *SwaggerController) oas2Doc(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.WithContext(r.Context()).Errorf("Failed to serve OAS document: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	doc, e := c.docLoader.Load()
	if e != nil {
		err = e
		return
	}

	switch oas, e := c.process(doc, r); {
	case e == nil && doc.Version == OASv2:
		// write to response
		w.Header().Set(web.HeaderContentType, "application/json")
		err = json.NewEncoder(w).Encode(oas)
	case e == nil:
		err = fmt.Errorf("OAS3 document is not supported by /v2 endpoint")
	default:
		err = e
	}
}

func (c *SwaggerController) oas3Doc(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.WithContext(r.Context()).Errorf("Failed to serve OAS document: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	doc, e := c.docLoader.Load()
	if e != nil {
		err = e
		return
	}

	oas, e := c.process(doc, r)
	if e != nil {
		err = e
		return
	}

	// write to response
	w.Header().Set(web.HeaderContentType, "application/json")
	err = json.NewEncoder(w).Encode(oas)
}

func (c *SwaggerController) swaggerRedirect(w http.ResponseWriter, r *http.Request) {
	fs := http.FS(Content)
	path := "generated/swagger-sso-redirect.html"
	file, err := fs.Open(path)
	if err != nil {
		logger.WithContext(r.Context()).Errorf("Unable to open file '%s': %v", path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logger.WithContext(r.Context()).Errorf("Unable to stat file '%s': %v", path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func (c *SwaggerController) msxVersion() string {
	if c.buildInfoResolver != nil {
		return c.buildInfoResolver.Resolve().Version
	}

	if strings.ToLower(bootstrap.BuildVersion) == "unknown" {
		return ""
	}
	return bootstrap.BuildVersion
}

func (c *SwaggerController) process(doc *OpenApiSpec, r *http.Request) (interface{}, error) {
	switch doc.Version {
	case OASv2:
		return doc.OAS2, c.processOAS2(doc.OAS2, r)
	case OASv30, OASv31:
		return doc.OAS3, c.processOAS3(doc.OAS3, r)
	}
	return nil, fmt.Errorf("unknown OAS document version")
}

func (c *SwaggerController) processOAS2(oas *OAS2, r *http.Request) error {
	// host
	var host string
	fwdAddress := r.Header.Get("X-Forwarded-Host") // capitalisation doesn't matter
	if fwdAddress != "" {
		ips := strings.Split(fwdAddress, ",")
		host = strings.TrimSpace(ips[0])
	} else {
		host = r.Host
	}
	oas.Host = host

	// version
	oas.Info.Version = c.msxVersion()
	return nil
}

func (c *SwaggerController) processOAS3(oas *OAS3, r *http.Request) error {
	// version
	oas.Info.Version = c.msxVersion()

	// host
	oas.Servers = nil
	fwdAddr := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")) // capitalisation doesn't matter
	if len(fwdAddr) != 0 {
		host := strings.Split(fwdAddr, ",")[0]

		fwdProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
		schema := "http"
		if len(fwdProto) != 0 {
			schema = strings.Split(fwdProto, ",")[0]
		}
		serverUrl := fmt.Sprintf("%s://%s", strings.TrimSpace(schema), strings.TrimSpace(host))
		if c.properties.BasePath != "" {
			basePath := strings.TrimSpace(c.properties.BasePath)
			if !strings.HasPrefix(c.properties.BasePath, "/") {
				basePath = fmt.Sprintf("/%s", basePath)
			}
			serverUrl = fmt.Sprintf("%s%s", serverUrl, basePath)
		}
		server := OAS3Server{
			URL:         serverUrl,
			Description: "Current API Server",
		}

		oas.Servers = append([]OAS3Server{server}, oas.Servers...)
	}
	return nil
}
