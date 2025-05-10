package httpclient

import (
	"context"
	"fmt"
	"net/url"
	"path"
)

/***********************
	BaseUrlTargetResolver
 ***********************/

func NewStaticTargetResolver(baseUrl string) (TargetResolverFunc, error) {
	base, e := url.Parse(baseUrl)
	if e != nil {
		return nil, e
	} else if !base.IsAbs() {
		return nil, fmt.Errorf(`expect abslolute base URL, but got "%s"`, baseUrl)
	}
	return func(ctx context.Context, req *Request) (*url.URL, error) {
		uri := *base
		uri.Path = path.Clean(path.Join(base.Path, req.Path))
		return &uri, nil
	}, nil
}
