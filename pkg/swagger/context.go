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
