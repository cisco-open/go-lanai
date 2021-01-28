package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

/*****************************
	Abstraction
 *****************************/
// AuthorizeRequestProcessor validate and process incoming request.
type AuthorizeRequestProcessor interface {
	Process(ctx context.Context, request *AuthorizeRequest) (validated *AuthorizeRequest, err error)
}

/*****************************
	Common Implementations
 *****************************/
type CompositeAuthorizeRequestProcessor struct {
	delegates []AuthorizeRequestProcessor
}

func NewCompositeAuthorizeRequestProcessor(delegates ...AuthorizeRequestProcessor) *CompositeAuthorizeRequestProcessor {
	return &CompositeAuthorizeRequestProcessor{delegates: delegates}
}

func (e *CompositeAuthorizeRequestProcessor) Process(ctx context.Context, request *AuthorizeRequest) (validated *AuthorizeRequest, err error) {
	for _, processor := range e.delegates {
		current, err := processor.Process(ctx, request)
		if err != nil {
			return nil, err
		}
		request = current
	}
	return request, nil
}

func (e *CompositeAuthorizeRequestProcessor) Add(enhancers... AuthorizeRequestProcessor) {
	e.delegates = append(e.delegates, flattenProcessors(enhancers)...)
	// resort the delegates
	order.SortStable(e.delegates, order.OrderedFirstCompare)
}

func (e *CompositeAuthorizeRequestProcessor) Remove(enhancer AuthorizeRequestProcessor) {
	for i, item := range e.delegates {
		if item != enhancer {
			continue
		}

		// remove but keep order
		if i + 1 <= len(e.delegates) {
			copy(e.delegates[i:], e.delegates[i+1:])
		}
		e.delegates = e.delegates[:len(e.delegates) - 1]
		return
	}
}

// flattenProcessors recursively flatten any nested CompositeAuthorizeRequestProcessor
func flattenProcessors(processors []AuthorizeRequestProcessor) (ret []AuthorizeRequestProcessor) {
	ret = make([]AuthorizeRequestProcessor, 0, len(processors))
	for _, e := range processors {
		switch e.(type) {
		case *CompositeAuthorizeRequestProcessor:
			flattened := flattenProcessors(e.(*CompositeAuthorizeRequestProcessor).delegates)
			ret = append(ret, flattened...)
		default:
			ret = append(ret, e)
		}
	}
	return
}

