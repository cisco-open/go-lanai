package swagger

import (
	"encoding/json"
	"errors"
)

const (
	OASv2  = "2.0"
	OASv30 = "3.0.0"
	OASv31 = "3.1.0"
)

type OASVersion string

type oasDoc struct {
	OASLegacyVer OASVersion `json:"swagger"`
	OAS3Ver      OASVersion `json:"openapi"`
}

type OpenApiSpec struct {
	Version OASVersion
	OAS2   *OAS2
	OAS3   *OAS3
}

func (s *OpenApiSpec) UnmarshalJSON(data []byte) error {
	var doc oasDoc
	if e := json.Unmarshal(data, &doc); e != nil {
		return e
	}

	var specPtr interface{}
	switch {
	case doc.OAS3Ver == OASv30, doc.OAS3Ver == OASv31:
		s.Version = doc.OAS3Ver
		s.OAS3 = &OAS3{}
		specPtr = s.OAS3
	case len(doc.OAS3Ver) == 0 && doc.OASLegacyVer == OASv2:
		s.Version = OASv2
		s.OAS2 = &OAS2{}
		specPtr = s.OAS2
	default:
		return errors.New("unknown OAS document version")
	}
	return json.Unmarshal(data, specPtr)
}

// OAS2 is Swagger 2.0 Specification
// https://swagger.io/docs/specification/2-0/basic-structure/
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md
type OAS2 struct {
	OpenAPIVersion string                 `json:"swagger"`
	Info           OAS2Info               `json:"info"`
	Host           string                 `json:"host,omitempty"`
	BasePath       string                 `json:"basePath,omitempty"`
	Schemes        []string               `json:"schemes,omitempty"`
	Consumes       []string               `json:"consumes,omitempty"`
	Produces       []string               `json:"produces,omitempty"`
	Paths          map[string]interface{} `json:"paths,omitempty"`
	Definitions    map[string]interface{} `json:"definitions,omitempty"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Responses      map[string]interface{} `json:"responses,omitempty"`
	SecDefs        map[string]interface{} `json:"securityDefinitions,omitempty"`
	Security       []interface{}          `json:"security,omitempty"`
	Tags           []interface{}          `json:"tags,omitempty"`
	ExtDocs        map[string]interface{} `json:"externalDocs,omitempty"`
}

type OAS2Info struct {
	Title          string                 `json:"title"`
	Description    string                 `json:"description,omitempty"`
	TermsOfService string                 `json:"termsOfService,omitempty"`
	Contact        map[string]interface{} `json:"contact,omitempty"`
	License        map[string]interface{} `json:"license,omitempty"`
	Version        string                 `json:"version,omitempty"`
}

// OAS3 is Swagger 3.0 Specification
// https://swagger.io/docs/specification/basic-structure/
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#schema
type OAS3 struct {
	OpenAPIVersion string                 `json:"openapi"`
	Info           OAS3Info               `json:"info"`
	JsonDialect    string                 `json:"jsonSchemaDialect,omitempty"`
	Servers        []OAS3Server           `json:"servers,omitempty"`
	Paths          map[string]interface{} `json:"paths,omitempty"`
	WebHooks       map[string]interface{} `json:"webhooks,omitempty"`
	Components     map[string]interface{} `json:"components,omitempty"`
	Security       []interface{}          `json:"security,omitempty"`
	Tags           []interface{}          `json:"tags,omitempty"`
	ExtDocs        map[string]interface{} `json:"externalDocs,omitempty"`
}

type OAS3Info struct {
	Title          string                 `json:"title"`
	Summary        string                 `json:"summary,omitempty"`
	Description    string                 `json:"description,omitempty"`
	TermsOfService string                 `json:"termsOfService,omitempty"`
	Contact        map[string]interface{} `json:"contact,omitempty"`
	License        map[string]interface{} `json:"license,omitempty"`
	Version        string                 `json:"version,omitempty"`
}

type OAS3Server struct {
	URL         string                 `json:"url"`
	Description string                 `json:"description,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}
