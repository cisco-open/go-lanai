package opaaccess

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var logger = log.New("OPA.Access")

// DecisionMakerWithOPA is an access.DecisionMakerFunc that utilize OPA engine
func DecisionMakerWithOPA(opa *sdk.OPA) access.DecisionMakerFunc {
	return func(ctx context.Context, req *http.Request) (handled bool, decision error) {
		opts := PrepareOpaQuery(ctx, req)
		result, e := opa.Decision(ctx, *opts)
		if e != nil {
			return true, security.NewAccessDeniedError(e)
		}
		logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
		switch v := result.Result.(type) {
		case bool:
			if v {
				return true, nil
			}
			return true, security.NewAccessDeniedError("Access Denied")
		default:
			return true, security.NewAccessDeniedError(fmt.Errorf("unknow OPA result type %T", result.Result))
		}
	}
}

type RequestClause struct {
	Scheme string      `json:"scheme"`
	Path   string      `json:"path"`
	Method string      `json:"method"`
	Body   string      `json:"body"` // Do not use it for now
	Header http.Header `json:"header"`
	Query  url.Values  `json:"query"`
	JWT    string      `json:"jwt"`
}

func NewRequestClause(req *http.Request) *RequestClause {
	var jwt string
	if v := req.Header.Get("Authorization"); len(v) != 0 && strings.HasPrefix(v, "Bearer ") {
		jwt = strings.TrimLeft(v, "Bearer ")
	}
	return &RequestClause{
		Scheme: req.URL.Scheme,
		Path:   req.URL.Path,
		Method: req.Method,
		Header: req.Header,
		Query:  req.URL.Query(),
		JWT:    jwt,
	}
}

type PermissionQuery struct {
	Authentication security.Authentication `json:"auth"`
	Request        *RequestClause          `json:"request"`
}

func PrepareOpaQuery(ctx context.Context, req *http.Request) *sdk.DecisionOptions {
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                "apimanage/allow_api",
		Input:               nil,
		StrictBuiltinErrors: false,
	}

	auth := security.Get(ctx)
	opts.Input = PermissionQuery{
		Authentication: auth,
		Request:        NewRequestClause(req),
	}
	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Errorf("Input: %s", data)
	}
	return &opts
}
