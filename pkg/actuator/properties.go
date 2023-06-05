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
	Security      SecurityProperties                 `json:"security"`
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
	Enabled *bool `json:"enabled"`
}

type SecurityProperties struct {
	EnabledByDefault bool                                  `json:"enabled-by-default"`
	Permissions      utils.CommaSeparatedSlice             `json:"permissions"`
	Endpoints        map[string]EndpointSecurityProperties `json:"endpoint"`
}

type EndpointSecurityProperties struct {
	Enabled     *bool                     `json:"enabled"`
	Permissions utils.CommaSeparatedSlice `json:"permissions"`
}

//NewManagementProperties create a ManagementProperties with default values
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
		Security: SecurityProperties{
			EnabledByDefault: false,
			Permissions:      []string{},
			Endpoints:        map[string]EndpointSecurityProperties{
				"alive": {
					Enabled: utils.ToPtr(false),
				},
				"info": {
					Enabled: utils.ToPtr(false),
				},
				"health": {
					Enabled: utils.ToPtr(false),
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
		panic(errors.Wrap(err, "failed to bind ManagementProperties"))
	}
	return *props
}
