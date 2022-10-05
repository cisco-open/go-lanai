package template

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"net/url"
	"path"
)

const (
	ModelKeyError          = "error"
	ModelKeyErrorCode      = "errorCode"
	ModelKeyStatusCode     = "statusCode"
	ModelKeyStatusText     = "statusText"
	ModelKeyMessage        = "message"
	ModelKeySession        = "session"
	ModelKeyRequestContext = "rc"
	ModelKeySecurity       = "security"
	ModelKeyCsrf           = "csrf"
)

var (
	viewRedirect          = "redirect:"
	modelKeyRedirectSC    = "redirect.sc"
	modelKeyRedirectLoc   = "redirect.location"
	modelKeyIgnoreCtxPath = "redirect.noCtxPath"
)

type Model gin.H

type ModelView struct {
	// View is the name of template file
	View string
	// Model is map[string]interface{}
	Model Model
}

func RedirectView(location string, statusCode int, ignoreContextPath bool) *ModelView {
	if statusCode < 300 || statusCode > 399 {
		statusCode = http.StatusFound
	}
	return &ModelView{
		View: viewRedirect,
		Model: Model{
			modelKeyRedirectSC:    statusCode,
			modelKeyRedirectLoc:   location,
			modelKeyIgnoreCtxPath: ignoreContextPath,
		},
	}
}

func isRedirect(mv *ModelView) (ret bool) {
	if mv.View != viewRedirect {
		return
	}
	if _, ok := mv.Model[modelKeyRedirectLoc].(string); !ok {
		return
	}
	if _, ok := mv.Model[modelKeyRedirectSC].(int); !ok {
		return
	}
	return true
}

func redirect(ctx context.Context, mv *ModelView) (int, string) {
	sc, _ := mv.Model[modelKeyRedirectSC].(int)
	location := mv.Model[modelKeyRedirectLoc].(string)

	loc, e := url.Parse(location)
	if e != nil {
		return sc, location
	}

	ignoreCtxPath, _ := mv.Model[modelKeyIgnoreCtxPath].(bool)
	if loc.IsAbs() || ignoreCtxPath {
		return sc, loc.String()
	}

	if ctxPath, ok := ctx.Value(web.ContextKeyContextPath).(string); ok {
		loc.Path = path.Join(ctxPath, loc.Path)
	}
	return sc, loc.String()
}

/**********************************
	Response Encoder
***********************************/

//goland:noinspection GoNameStartsWithPackageName
func TemplateEncodeResponseFunc(ctx context.Context, _ http.ResponseWriter, response interface{}) error {
	gc := web.GinContext(ctx)
	if gc == nil {
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

	var mv *ModelView
	switch v := response.(type) {
	case *ModelView:
		mv = v
	case ModelView:
		mv = &v
	default:
		return errors.New("unable to use template: response is not *template.ModelView")
	}

	switch {
	case isRedirect(mv):
		gc.Redirect(redirect(ctx, mv))
	default:
		AddGlobalModelData(ctx, mv.Model, gc.Request)
		gc.HTML(status, mv.View, mv.Model)
	}
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
	//nolint:errorlint
	if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}

	model := Model{
		ModelKeyError:      err,
		ModelKeyMessage:    err.Error(),
		ModelKeyStatusCode: code,
		ModelKeyStatusText: http.StatusText(code),
	}

	AddGlobalModelData(c, model, ctx.Request)
	ctx.HTML(code, web.ErrorTemplate, model)
}

func AddGlobalModelData(ctx context.Context, model Model, r *http.Request) {
	model[ModelKeyRequestContext] = MakeRequestContext(ctx, r, web.ContextKeyContextPath)
	model[ModelKeySession] = ctx.Value(web.ContextKeySession)
	model[ModelKeySecurity] = ctx.Value(web.ContextKeySecurity)
	model[ModelKeyCsrf] = ctx.Value(web.ContextKeyCsrf)
}
