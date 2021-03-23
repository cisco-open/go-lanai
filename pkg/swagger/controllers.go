package swagger

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"net/http"
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

type SwaggerController struct {
	properties SwaggerProperties
}

func NewSwaggerController(prop SwaggerProperties) *SwaggerController {
	return &SwaggerController{
		properties: prop,
	}
}

func (s *SwaggerController) Mappings() []web.Mapping {
	return []web.Mapping{
		assets.New("swagger/static", "frontend"),
		web.NewSimpleMapping("swagger-ui", "/swagger", "GET", nil, s.swagger),
		rest.New("swagger-configuration-ui").Get("/swagger-resources/configuration/ui").EndpointFunc(s.configurationUi).Build(),
		rest.New("swagger-configuration-security").Get("/swagger-resources/configuration/security").EndpointFunc(s.configurationSecurity).Build(),
		rest.New("swagger-configuration-sso").Get("/swagger-resources/configuration/security/sso").EndpointFunc(s.configurationSso).Build(),
		rest.New("swagger-resources").Get("/swagger-resources").EndpointFunc(s.resources).Build(),
		web.NewSimpleMapping("swagger-sso-redirect", "swagger-sso-redirect.html", "GET", nil, s.swaggerRedirect),
		web.NewSimpleMapping("swagger-spec", "v2/api-docs", "GET", nil, s.swaggerSpec),
	}
}

func (s *SwaggerController) configurationUi(ctx context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = UiConfiguration{
		DeepLinking: true,
		DisplayOperationId: false,
		DefaultModelsExpandDepth: 1,
		DefaultModelExpandDepth: 1,
		DefaultModelRendering: "example",
		DisplayRequestDuration: false,
		DocExpansion: "none",
		Filter: false,
		OperationsSorter: "alpha",
		ShowExtensions: false,
		TagsSorter: "alpha",
		ValidatorUrl: "",
		SupportedSubmitMethods: []string{ "get", "put", "post", "delete", "options", "head", "patch", "trace"},
	}
	return
}

func (s *SwaggerController) swagger(w http.ResponseWriter, r *http.Request) {
	fs := http.FS(Content)
	file, err := fs.Open("frontend/swagger-ui.html")
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

func (s *SwaggerController) configurationSecurity(ctx context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = struct{}{}
	return
}

func (s *SwaggerController) configurationSso(ctx context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = SsoConfiguration{
		TokenUrl: fmt.Sprintf("%s%s", s.properties.Security.Sso.BaseUrl, s.properties.Security.Sso.TokenPath),
		AuthorizeUrl: fmt.Sprintf("%s%s", s.properties.Security.Sso.BaseUrl, s.properties.Security.Sso.AuthorizePath),
		ClientId: s.properties.Security.Sso.ClientId,
		ClientSecret: s.properties.Security.Sso.ClientSecret,
	}
	return
}

func (s *SwaggerController) resources(ctx context.Context, _ web.EmptyRequest) (response interface{}, err error) {
	response = []Resource{
		{
			Name: "platform",
			Url: "/v2/api-docs?group=platform",
			SwaggerVersion: "2.0",
			Location: "/v2/api-docs?group=platform",
		},
	}
	return
}


func (s *SwaggerController) swaggerRedirect(w http.ResponseWriter, r *http.Request) {
	fs := http.FS(Content)
	file, err := fs.Open("frontend/swagger-sso-redirect.html")
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

func (s *SwaggerController) swaggerSpec(w http.ResponseWriter, r *http.Request) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(s.properties.Spec)
	resp, err := swagger.MarshalJSON()

	if err == nil {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, string(resp))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}