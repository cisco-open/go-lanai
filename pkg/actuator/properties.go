package actuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
)

const (
	ManagementPropertiesPrefix = "management"
)

type ManagementProperties struct {
	Enabled       bool                               `json:"enabled"`
	Endpoints     EndpointsProperties                `json:"endpoints"`
	BasicEndpoint map[string]BasicEndpointProperties `json:"endpoint"`
}

type EndpointsProperties struct {
	EnabledByDefault bool                   `json:"enabled-by-default"`
	Web              WebEndpointsProperties `json:"web"`
}

type WebEndpointsProperties struct {
	BasePath string                `json:"base-path"`
	Mappings map[string]string     `json:"path-mapping"`
	Exposure WebExposureProperties `json:"exposure"`
}

type WebExposureProperties struct {
	// Endpoint IDs that should be included or '*' for all.
	Include utils.StringSet `json:"include"`
	// Endpoint IDs that should be excluded or '*' for all.
	Exclude utils.StringSet `json:"exclude"`
}

type BasicEndpointProperties struct {
	Enabled bool                   `json:"enabled"`
}

//NewSessionProperties create a SessionProperties with default values
func NewManagementProperties() *ManagementProperties {
	return &ManagementProperties{
		Enabled: true,
		Endpoints: EndpointsProperties{
			Web: WebEndpointsProperties{
				BasePath: "/manage",
				Mappings: map[string]string{},
				Exposure: WebExposureProperties{
					Include: utils.NewStringSet("*"),
					Exclude: utils.NewStringSet(),
				},
			},
		},
		BasicEndpoint: map[string]BasicEndpointProperties{},
	}
}

//BindManagementProperties create and bind SessionProperties, with a optional prefix
func BindManagementProperties(ctx *bootstrap.ApplicationContext) ManagementProperties {
	props := NewManagementProperties()
	if err := ctx.Config().Bind(props, ManagementPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SessionProperties"))
	}
	return *props
}
