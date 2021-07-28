package auth

import (
	"context"
)

/*****************************
	Abstraction
 *****************************/

// AuthorizeRequestProcessor validate and process incoming request
// AuthorizeRequestProcessor is the entry point interface for other components to use
type AuthorizeRequestProcessor interface {
	Process(ctx context.Context, request *AuthorizeRequest) (processed *AuthorizeRequest, err error)
}

// AuthorizeRequestProcessChain invoke index processor in the processing chain
type AuthorizeRequestProcessChain interface {
	Next(ctx context.Context, request *AuthorizeRequest) (processed *AuthorizeRequest, err error)
}

// ChainedAuthorizeRequestProcessor validate and process incoming request and manually invoke index processor in the chain.
type ChainedAuthorizeRequestProcessor interface {
	Process(ctx context.Context, request *AuthorizeRequest, chain AuthorizeRequestProcessChain) (validated *AuthorizeRequest, err error)
}

/*****************************
	Common Implementations
 *****************************/

// authorizeRequestProcessor implements AuthorizeRequestProcessor
type authorizeRequestProcessor struct {
	delegates []ChainedAuthorizeRequestProcessor
}

func NewAuthorizeRequestProcessor(delegates ...ChainedAuthorizeRequestProcessor) AuthorizeRequestProcessor {
	return &authorizeRequestProcessor{delegates: delegates}
}

func (p *authorizeRequestProcessor) Process(ctx context.Context, request *AuthorizeRequest) (processed *AuthorizeRequest, err error) {
	chain := arProcessChain{delegates: p.delegates}
	return chain.Next(ctx, request)
}

// arProcessChain implements AuthorizeRequestProcessChain
type arProcessChain struct {
	index     int
	delegates []ChainedAuthorizeRequestProcessor
}

func (c arProcessChain) Next(ctx context.Context, request *AuthorizeRequest) (processed *AuthorizeRequest, err error) {
	if c.index >= len(c.delegates) {
		return request, nil
	}

	next := c.delegates[c.index]
	c.index++
	return next.Process(ctx, request, c)
}



//func (c *authorizeRequestProcessor) Add(processors ...ChainedAuthorizeRequestProcessor) {
//	c.delegates = append(c.delegates, flattenProcessors(processors)...)
//	// resort the extensions
//	order.SortStable(c.delegates, order.OrderedFirstCompare)
//}
//
//func (c *authorizeRequestProcessor) Remove(processor ChainedAuthorizeRequestProcessor) {
//	for i, item := range c.delegates {
//		if item != processor {
//			continue
//		}
//
//		// remove but keep order
//		if i+1 <= len(c.delegates) {
//			copy(c.delegates[i:], c.delegates[i+1:])
//		}
//		c.delegates = c.delegates[:len(c.delegates)-1]
//		return
//	}
//}
//
//// flattenProcessors recursively flatten any nested NestedAuthorizeRequestProcessor
//func flattenProcessors(processors []ChainedAuthorizeRequestProcessor) (ret []ChainedAuthorizeRequestProcessor) {
//	ret = make([]ChainedAuthorizeRequestProcessor, 0, len(processors))
//	for _, e := range processors {
//		switch e.(type) {
//		case *authorizeRequestProcessor:
//			flattened := flattenProcessors(e.(*authorizeRequestProcessor).delegates)
//			ret = append(ret, flattened...)
//		default:
//			ret = append(ret, e)
//		}
//	}
//	return
//}
