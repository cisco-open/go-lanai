package swagger

const SwaggerPrefix = "swagger"

type SwaggerProperties struct {
	BasePath string                    `json:"base-path"`
	Spec     string                    `json:"spec"`
	Security SwaggerSecurityProperties `json:"security"`
}

type SwaggerSecurityProperties struct {
	SecureDocs bool                 `json:"secure-docs"`
	Sso        SwaggerSsoProperties `json:"sso"`
}

type SwaggerSsoProperties struct {
	BaseUrl          string                `json:"base-url"`
	TokenPath        string                `json:"token-path"`
	AuthorizePath    string                `json:"authorize-path"`
	ClientId         string                `json:"client-id"`
	ClientSecret     string                `json:"client-secret"`
	AdditionalParams []ParameterProperties `json:"additional-params"`
}

type ParameterProperties struct {
	Name      string `json:"name"`
	SourceUrl string `json:"source-url"`
	JsonPath  string `json:"json-path"`
}

func NewSwaggerSsoProperties() *SwaggerProperties {
	return &SwaggerProperties{}
}
