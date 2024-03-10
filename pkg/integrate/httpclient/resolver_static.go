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

func NewStaticTargetResolver(baseUrls ...string) (TargetResolverFunc, error) {
	targets := make([]*url.URL, len(baseUrls))
	for i, base := range baseUrls {
		uri, e := url.Parse(base)
		if e != nil {
			return nil, e
		} else if !uri.IsAbs() {
			return nil, fmt.Errorf(`expect abslolute base URL, but got "%s"`, base)
		}
		targets[i] = uri
	}
	return func(ctx context.Context, req *Request) (*url.URL, error) {
		for _, base := range targets {
			uri := *base
			uri.Path = path.Clean(path.Join(base.Path, req.Path))
			return &uri, nil
		}
		return nil, NewNoEndpointFoundError(fmt.Errorf("base URL is not available"))
	}, nil
}


