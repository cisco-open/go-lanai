package template

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

type Model gin.H

type ModelView struct {
	// StatusCode when left zero, default is 200
	StatusCode int
	// View is the name of template file
	View string
	// Model is map[string]interface{}
	Model Model
}

/**********************************
	Response Encoder
***********************************/
func ginTemplateEncodeResponseFunc(c context.Context, _ http.ResponseWriter, response interface{}) error {
	ctx, ok := c.(*gin.Context)
	if !ok {
		return errors.New("unable to use template: context is not available")
	}
	mv, ok := response.(*ModelView)
	if !ok {
		return errors.New("unable to use template: response is not *template.ModelView")
	}

	status := 200
	if mv.StatusCode != 0 {
		status = mv.StatusCode
	}
	// TODO merge model with global overrides
	ctx.HTML(status, mv.View, mv.Model)
	return nil
}

/*****************************
	JSON Error Encoder
******************************/
func templateErrorEncoder(c context.Context, err error, w http.ResponseWriter) {
	ctx, ok := c.(*gin.Context)
	if !ok {
		httptransport.DefaultErrorEncoder(c, err, w)
		return
	}

	code := http.StatusInternalServerError
	if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}

	// TODO merge model with global overrides
	ctx.HTML(code, "error.tmpl", gin.H{
		"error": err,
		"StatusCode": code,
		"StatusText": http.StatusText(code),
	})
}



