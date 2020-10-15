package template

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Model gin.H

type ModelView struct {
	// Status when left zero, default is 200
	Status int
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
	if mv.Status != 0 {
		status = mv.Status
	}
	// TODO merge model with global overrides
	ctx.HTML(status, mv.View, mv.Model)
	return nil
}



