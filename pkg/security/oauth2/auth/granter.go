package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
)

type TokenGranter interface {
	// Grant create oauth2.AccessToken based on given TokenRequest
	// returns
	// 	- (nil, nil) if the TokenGranter doesn't support given request
	// 	- (non-nil, nil) if the TokenGranter support given request and created a token without error
	// 	- (nil, non-nil) if the TokenGranter support given request but rejected the request
	Grant(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error)
}

// CompositeTokenGranter implements TokenGranter
type CompositeTokenGranter struct {
	delegates []TokenGranter
}

func NewCompositeTokenGranter(delegates...TokenGranter) *CompositeTokenGranter {
	return &CompositeTokenGranter{
		delegates: delegates,
	}
}

func (g *CompositeTokenGranter) Grant(ctx context.Context, request *TokenRequest) (oauth2.AccessToken, error) {
	for _, granter := range g.delegates {
		if token, e := granter.Grant(ctx, request); e != nil {
			return nil, e
		} else if token != nil {
			return token, nil
		}
	}
	return nil, oauth2.NewGranterNotAvailableError(fmt.Sprintf("grant type [%s] is not supported", request.GrantType))
}

func (g *CompositeTokenGranter) Add(granter TokenGranter) *CompositeTokenGranter {
	g.delegates = append(g.delegates, granter)
	return g
}

func (g *CompositeTokenGranter) Delegates() []TokenGranter {
	return g.delegates
}
