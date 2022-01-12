package swagger

// OAS2Specs is Swagger 2.0 Specification
// https://swagger.io/docs/specification/2-0/basic-structure/
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md
type OAS2Specs struct {
	OpenAPIVersion string                 `json:"swagger"`
	Info           OAS2Info               `json:"info"`
	Host           string                 `json:"host"`
	BasePath       string                 `json:"basePath"`
	Schemes        []string               `json:"schemes"`
	Consumes       []string               `json:"consumes"`
	Produces       []string               `json:"produces"`
	Paths          map[string]interface{} `json:"paths"`
	Definitions    map[string]interface{} `json:"definitions"`
	Parameters     map[string]interface{} `json:"parameters"`
	Responses      map[string]interface{} `json:"responses"`
	SecDefs        map[string]interface{} `json:"securityDefinitions"`
	Security       []interface{}          `json:"security"`
	Tags           []interface{}          `json:"tags"`
	ExtDocs        map[string]interface{} `json:"externalDocs"`
}

type OAS2Info struct {
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	TermsOfService string                 `json:"termsOfService"`
	Contact        map[string]interface{} `json:"contact"`
	License        map[string]interface{} `json:"license"`
	Version        string                 `json:"version"`
}

// OAS3Specs is Swagger 3.0 Specification
// https://swagger.io/docs/specification/basic-structure/
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#schema
type OAS3Specs struct {
	OpenAPIVersion string                 `json:"openapi"`
	Info           OAS3Info               `json:"info"`
	JsonDialect    string                 `json:"jsonSchemaDialect"`
	Servers        []OAS3Server           `json:"servers"`
	Paths          map[string]interface{} `json:"paths"`
	WebHooks       map[string]interface{} `json:"webhooks"`
	Components     map[string]interface{} `json:"components"`
	Security       []interface{}          `json:"security"`
	Tags           []interface{}          `json:"tags"`
	ExtDocs        map[string]interface{} `json:"externalDocs"`
}

type OAS3Info struct {
	Title          string                 `json:"title"`
	Summary        string                 `json:"summary"`
	Description    string                 `json:"description"`
	TermsOfService string                 `json:"termsOfService"`
	Contact        map[string]interface{} `json:"contact"`
	License        map[string]interface{} `json:"license"`
	Version        string                 `json:"version"`
}

type OAS3Server struct {
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Variables   map[string]interface{} `json:"variables"`
}
