package testdata

import (
	"net/url"
	"strconv"
)

type JsonRequest struct {
	UriVar     string `uri:"var"`
	QueryVar   string `form:"q"`
	HeaderVar  string `header:"X-VAR"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

type Response struct {
	UriVar     string `json:"uri"`
	QueryVar   string `json:"q"`
	HeaderVar  string `json:"header"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

func newResponse(req *JsonRequest) *Response {
	return &Response{
		UriVar:     req.UriVar,
		QueryVar:   req.QueryVar,
		HeaderVar:  req.HeaderVar,
		JsonString: req.JsonString,
		JsonInt:    req.JsonInt,
	}
}

type JsonResponse Response

func newJsonResponse(req *JsonRequest) *JsonResponse {
	return (*JsonResponse)(newResponse(req))
}

type TextResponse Response

func newTextResponse(req *JsonRequest) *TextResponse {
	return (*TextResponse)(newResponse(req))
}

func (r TextResponse) MarshalText() ([]byte, error) {
	values := url.Values{}
	values.Set("uri", r.UriVar)
	values.Set("q", r.QueryVar)
	values.Set("header", r.HeaderVar)
	values.Set("string", r.JsonString)
	values.Set("int", strconv.Itoa(r.JsonInt))
	return []byte(values.Encode()), nil
}

type BytesResponse Response

func newBytesResponse(req *JsonRequest) *BytesResponse {
	return (*BytesResponse)(newResponse(req))
}

func (r BytesResponse) MarshalBinary() ([]byte, error) {
	return TextResponse(r).MarshalText()
}