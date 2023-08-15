package swagger

const SwaggerPrefix = "swagger"

type SwaggerProperties struct {
	BasePath string                    `json:"base-path"`
	Spec     string                    `json:"spec"`
	Security SwaggerSecurityProperties `json:"security"`
	UI       SwaggerUIProperties       `json:"ui"`
}

type SwaggerSecurityProperties struct {
	SecureDocs bool                 `json:"secure-docs"`
	Sso        SwaggerSsoProperties `json:"sso"`
}

type SwaggerSsoProperties struct {
	BaseUrl       string `json:"base-url"`
	TokenPath     string `json:"token-path"`
	AuthorizePath string `json:"authorize-path"`
	ClientId      string `json:"client-id"`
	ClientSecret  string `json:"client-secret"`
}

type SwaggerUIProperties struct {
	Title string `json:"title"`
}

func NewSwaggerSsoProperties() *SwaggerProperties {
	return &SwaggerProperties{}
}
