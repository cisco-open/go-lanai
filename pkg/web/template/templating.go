// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package template

import (
    "context"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/web"
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

type ModelValuer interface{
	~func() interface{} | ~func(ctx context.Context) interface{} | ~func(req *http.Request) interface{}
}

func StaticModelValuer(value interface{}) func() interface{} {
	return func() interface{} {
		return value
	}
}

func ContextModelValuer[T any](fn func(ctx context.Context) T) func(context.Context) interface{} {
	return func(ctx context.Context) interface{} {
		return fn(ctx)
	}
}

func RequestModelValuer[T any](fn func(req *http.Request) T) func(req *http.Request) interface{} {
	return func(req *http.Request) interface{} {
		return fn(req)
	}
}

var globalModelValuers = map[string]interface{}{}

// RegisterGlobalModelValuer register a ModelValuer with given model key. The registered ModelValuer is applied
// before any ModelView is rendered.
// Use StaticModelValuer, ContextModelValuer or RequestModelValuer to wrap values/functions as ModelValuer
func RegisterGlobalModelValuer[T ModelValuer](key string, valuer T) {
	globalModelValuers[key] = any(valuer)
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

	ctxPath := web.ContextPath(ctx)
	loc.Path = path.Join(ctxPath, loc.Path)
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
	model[ModelKeyRequestContext] = MakeRequestContext(ctx, r)
	applyGlobalModelValuers(ctx, r, model)
}

func applyGlobalModelValuers(ctx context.Context, r *http.Request, model Model) {
	for k, valuer := range globalModelValuers {
		var v interface{}
		switch fn := valuer.(type) {
		case func() interface{}:
			v = fn()
		case func(ctx context.Context) interface{}:
			v = fn(ctx)
		case func(req *http.Request) interface{}:
			v = fn(r)
		}
		if v != nil {
			model[k] = v
		}
	}

}
