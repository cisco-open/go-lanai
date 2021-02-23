package cors

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

const (
	CorsPropertiesPrefix = "security.cors"
	listSeparator        = ","
)

var (
	allMethods = []string{
		http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		//http.MethodConnect, http.MethodOptions, http.MethodTrace,
	}
)

type CorsProperties struct {
	Enabled bool `json:"enabled"`
	// Comma-separated list of origins to allow. '*' allows all origins. Default to '*'
	AllowedOriginsStr string `json:"allowed-origins"`
	// Comma-separated list of methods to allow. '*' allows all methods. Default to '*'
	AllowedMethodsStr string `json:"allowed-methods"`
	// Comma-separated list of headers to allow in a request. '*' allows all headers. Default to '*'
	AllowedHeadersStr string `json:"allowed-headers"`
	// Comma-separated list of headers to include in a response.
	ExposedHeadersStr string `json:"exposed-headers"`
	// Whether credentials are supported. When not set, credentials are not supported.
	AllowCredentials bool `json:"allow-credentials"`
	// How long the response from a pre-flight request can be cached by clients.
	// If a duration suffix is not specified, seconds will be used.
	MaxAge utils.Duration `json:"max-age"`
}

// NewCorsProperties create a ServerProperties with default values
func NewCorsProperties() *CorsProperties {
	return &CorsProperties{
		Enabled:           false,
		AllowedOriginsStr: "*",
		AllowedMethodsStr: "*",
		AllowedHeadersStr: "*",
		ExposedHeadersStr: "",
		AllowCredentials:  false,
		MaxAge:            utils.Duration(1800 * time.Second),
	}
}

func (p CorsProperties) AllowedOrigins() []string {
	return splitAndTrim(p.AllowedOriginsStr)
}

func (p CorsProperties) AllowedMethods() []string {
	list := splitAndTrim(p.AllowedMethodsStr)
	for _, v := range list {
		if v == "*" {
			return allMethods
		}
	}
	return list
}

func (p CorsProperties) AllowedHeaders() []string {
	return splitAndTrim(p.AllowedHeadersStr)
}

func (p CorsProperties) ExposedHeaders() []string {
	return splitAndTrim(p.ExposedHeadersStr)
}

//BindCorsProperties create and bind a ServerProperties using default prefix
func BindCorsProperties(ctx *bootstrap.ApplicationContext) CorsProperties {
	props := NewCorsProperties()
	if err := ctx.Config().Bind(props, CorsPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CorsProperties"))
	}
	return *props
}

func splitAndTrim(s string) []string {
	list := strings.Split(s, listSeparator)
	for i, v := range list {
		list[i] = strings.TrimSpace(v)
	}
	return list
}