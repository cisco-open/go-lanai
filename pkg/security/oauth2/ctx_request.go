package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
)

/******************************
	OAuth2Request
******************************/
var excludedParameters = utils.NewStringSet(ParameterPassword, ParameterClientSecret)

type OAuth2Request interface {
	Parameters() map[string]string
	ClientId() string
	Scopes() utils.StringSet
	Approved() bool
	GrantType() string
	RedirectUri() string
	ResponseTypes() utils.StringSet
	Extensions() map[string]interface{}
}

/******************************
	Implementation
******************************/
type RequestDetails struct {
	Parameters    map[string]string      `json:"parameters"`
	ClientId      string                 `json:"clientId"`
	Scopes        utils.StringSet        `json:"scope"`
	Approved      bool                   `json:"approved"`
	GrantType     string                 `json:"grantType"`
	RedirectUri   string                 `json:"redirectUri"`
	ResponseTypes utils.StringSet        `json:"responseTypes"`
	Extensions    map[string]interface{} `json:"extensions"`
}

type RequestOptionsFunc func(*RequestDetails)

type oauth2Request struct {
	RequestDetails
}

func NewOAuth2Request(optFuncs ...RequestOptionsFunc) OAuth2Request {
	request := oauth2Request{ RequestDetails: RequestDetails{
		Parameters: map[string]string{},
		Scopes: utils.NewStringSet(),
		ResponseTypes: utils.NewStringSet(),
		Extensions: map[string]interface{}{},
	}}

	for _, optFunc := range optFuncs {
		optFunc(&request.RequestDetails)
	}

	for param, _ := range excludedParameters {
		delete(request.RequestDetails.Parameters, param)
	}
	return &request
}

func (r *oauth2Request) Parameters() map[string]string {
	return r.RequestDetails.Parameters
}

func (r *oauth2Request) ClientId() string {
	return r.RequestDetails.ClientId
}

func (r *oauth2Request) Scopes() utils.StringSet {
	return r.RequestDetails.Scopes
}

func (r *oauth2Request) Approved() bool {
	return r.RequestDetails.Approved
}

func (r *oauth2Request) GrantType() string {
	return r.RequestDetails.GrantType
}

func (r *oauth2Request) RedirectUri() string {
	return r.RequestDetails.RedirectUri
}

func (r *oauth2Request) ResponseTypes() utils.StringSet {
	return r.RequestDetails.ResponseTypes
}

func (r *oauth2Request) Extensions() map[string]interface{} {
	return r.RequestDetails.Extensions
}

// json.Marshaler
func (r *oauth2Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.RequestDetails)
}

// json.Unmarshaler
func (r *oauth2Request) UnmarshalJSON(data []byte) error {
	if e := json.Unmarshal(data, &r.RequestDetails); e != nil {
		return e
	}
	return nil
}


