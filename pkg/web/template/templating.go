package template

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

const (
	ModelKeyError          = "error"
	ModelKeyErrorCode      = "errorCode"
	ModelKeyStatusCode     = "statusCode"
	ModelKeyStatusText     = "statusText"
	ModelKeyMessage        = "message"
	ModelKeySession        = "session"
	ModelKeyRequestContext = "rc"
	ModelKeySecurity 	   = "security"
	ModelKeyCsrf 		   = "csrf"
)

type Model gin.H

type ModelView struct {
	// View is the name of template file
	View string
	// Model is map[string]interface{}
	Model Model
}

/**********************************
	Response Encoder
***********************************/
func TemplateEncodeResponseFunc(c context.Context, _ http.ResponseWriter, response interface{}) error {
	ctx := web.GinContext(c)
	if ctx == nil {
		return errors.New("unable to use template: context is not available")
	}

	// get status code
	status := 200
	if coder, ok := response.(httptransport.StatusCoder); ok {
		status = coder.StatusCode()
	}

	if entity, ok := response.(web.BodyContainer); ok {
		response = entity.Body()
	}

	mv, ok := response.(*ModelView)
	if !ok {
		return errors.New("unable to use template: response is not *template.ModelView")
	}

	AddGlobalModelData(ctx, mv.Model)
	ctx.HTML(status, mv.View, mv.Model)
	return nil
}

/*****************************
	JSON Error Encoder
******************************/
//goland:noinspection GoNameStartsWithPackageName
func TemplateErrorEncoder(c context.Context, err error, w http.ResponseWriter) {
	ctx := web.GinContext(c)
	if ctx == nil {
		httptransport.DefaultErrorEncoder(c, err, w)
		return
	}

	code := http.StatusInternalServerError
	if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}

	model := Model{
		ModelKeyError: err,
		ModelKeyMessage: err.Error(),
		ModelKeyStatusCode: code,
		ModelKeyStatusText: http.StatusText(code),
	}

	AddGlobalModelData(ctx, model)
	ctx.HTML(code, web.ErrorTemplate, model)
}

func AddGlobalModelData(ctx *gin.Context, model Model) {
	model[ModelKeyRequestContext] = MakeRequestContext(ctx, ctx.Request, web.ContextKeyContextPath)
	model[ModelKeySession] = ctx.Value(web.ContextKeySession)
	model[ModelKeySecurity] = ctx.Value(web.ContextKeySecurity)
	model[ModelKeyCsrf] = ctx.Value(web.ContextKeyCsrf)
}



