package swagger

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
}

type SsoConfiguration struct {
	Enabled      bool   `json:"enabled"`
	AuthorizeUrl string `json:"authorizeUrl"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	TokenUrl     string `json:"tokenUrl"`
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
}

func NewSwaggerController(props SwaggerProperties, resolver bootstrap.BuildInfoResolver) *SwaggerController {
	return &SwaggerController{
		properties:        &props,
		buildInfoResolver: resolver,
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
		web.NewSimpleMapping("swagger-spec", "", "/v2/api-docs", http.MethodGet, nil, c.swaggerSpec),
		web.NewSimpleMapping("oas3-spec", "", "/v3/api-docs", http.MethodGet, nil, c.oas3Spec),
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
	response = SsoConfiguration{
		Enabled:      c.properties.Security.Sso.ClientId != "",
		TokenUrl:     fmt.Sprintf("%s%s", c.properties.Security.Sso.BaseUrl, c.properties.Security.Sso.TokenPath),
		AuthorizeUrl: fmt.Sprintf("%s%s", c.properties.Security.Sso.BaseUrl, c.properties.Security.Sso.AuthorizePath),
		ClientId:     c.properties.Security.Sso.ClientId,
		ClientSecret: c.properties.Security.Sso.ClientSecret,
	}
	return
}

func (c *SwaggerController) resources(_ context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = []Resource{
		{
			Name:           "platform",
			Url:            "/v3/api-docs?group=platform",
			SwaggerVersion: "2.0",
			Location:       "/v3/api-docs?group=platform",
		},
	}
	return
}

func (c *SwaggerController) swaggerSpec(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	file, err := os.Open(c.properties.Spec)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	var m OAS2Specs
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	if err != nil {
		return
	}

	// host
	var host string
	fwdAddress := r.Header.Get("X-Forwarded-Host") // capitalisation doesn't matter
	if fwdAddress != "" {
		ips := strings.Split(fwdAddress, ",")
		host = strings.TrimSpace(ips[0])
	} else {
		host = r.Host
	}
	m.Host = host

	// version
	m.Info.Version = c.msxVersion()

	// write to response
	w.Header().Set(web.HeaderContentType, "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&m)
}

func (c *SwaggerController) oas3Spec(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	file, err := os.Open(c.properties.Spec)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	var m OAS3Specs
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	if err != nil {
		return
	}

	// version
	m.Info.Version = c.msxVersion()

	// host
	fwdAddress := r.Header.Get("X-Forwarded-Host") // capitalisation doesn't matter
	if fwdAddress != "" {
		ips := strings.Split(fwdAddress, ",")
		server := OAS3Server{
			URL:         "{schema}://{host}",
			Description: "Current API Server",
			Variables: map[string]interface{}{
				"schema": "http", //TODO should be dynamic value
				"host":   strings.TrimSpace(ips[0]),
			},
		}
		m.Servers = append([]OAS3Server{server}, m.Servers...)
	}

	// write to response
	w.Header().Set(web.HeaderContentType, "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&m)
}

func (c *SwaggerController) swaggerRedirect(w http.ResponseWriter, r *http.Request) {
	fs := http.FS(Content)
	file, err := fs.Open("generated/swagger-sso-redirect.html")
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

func (c *SwaggerController) msxVersion() string {
	if c.buildInfoResolver != nil {
		return c.buildInfoResolver.Resolve().Version
	}

	if strings.ToLower(bootstrap.BuildVersion) == "unknown" {
		return ""
	}
	return bootstrap.BuildVersion
}
