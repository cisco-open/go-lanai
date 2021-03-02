package cors

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	cors "github.com/rs/cors/wrapper/gin"
	"time"
)

// Customizer implements web.Customizer
type Customizer struct {
	properties CorsProperties
}

func newCustomizer(properties CorsProperties) *Customizer {
	return &Customizer{
		properties: properties,
	}
}

func (c *Customizer) Customize(ctx context.Context, r *web.Registrar) (err error) {
	if !c.properties.Enabled {
		return
	}

	mw := cors.New(cors.Options{
		AllowedOrigins:     c.properties.AllowedOrigins(),
		AllowedMethods:     c.properties.AllowedMethods(),
		AllowedHeaders:     c.properties.AllowedHeaders(),
		ExposedHeaders:     c.properties.ExposedHeaders(),
		MaxAge:             int(time.Duration(c.properties.MaxAge).Seconds()),
		AllowCredentials:   c.properties.AllowCredentials,
		OptionsPassthrough: false,
		//Debug:              true,
	})
	err = r.AddGlobalMiddlewares(mw)
	return
}

