package opa

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
	"net/url"
)

/********************
	Common Inputs
 ********************/

type InputApiAccess struct {
	Authentication security.Authentication `json:"auth,omitempty"`
	Request        *RequestClause          `json:"request,omitempty"`
}

/**************************
	Common Input Blocks
 **************************/

type RequestClause struct {
	Scheme string      `json:"scheme,omitempty"`
	Path   string      `json:"path,omitempty"`
	Method string      `json:"method,omitempty"`
	Header http.Header `json:"header,omitempty"`
	Query  url.Values  `json:"query,omitempty"`
}

func NewRequestClause(req *http.Request) *RequestClause {
	return &RequestClause{
		Scheme: req.URL.Scheme,
		Path:   req.URL.Path,
		Method: req.Method,
		Header: req.Header,
		Query:  req.URL.Query(),
	}
}


