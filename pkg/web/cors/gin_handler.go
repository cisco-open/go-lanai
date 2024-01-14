package cors

import (
    "github.com/gin-gonic/gin"
    "github.com/rs/cors"
    "net/http"
)

// Options is a configuration container to setup the CORS middleware.
type Options = cors.Options

type corsWrapper struct {
    *cors.Cors
    optionPassthrough bool
}

// build transforms wrapped cors.Cors handler into Gin middleware.
func (c corsWrapper) build() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        c.HandlerFunc(ctx.Writer, ctx.Request)
        if !c.optionPassthrough &&
            ctx.Request.Method == http.MethodOptions &&
            ctx.GetHeader("Access-Control-Request-Method") != "" {
            // Abort processing next Gin middlewares.
            ctx.AbortWithStatus(http.StatusOK)
        }
    }
}

// New creates a new CORS Gin middleware with the provided options.
func New(options Options) gin.HandlerFunc {
    return corsWrapper{cors.New(options), options.OptionsPassthrough}.build()
}

