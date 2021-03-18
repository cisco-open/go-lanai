package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	util_matcher "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strings"
)

var (
	excludeRequest = util_matcher.Or(&healthMatcher, &corsPreflightMatcher)
)

type TracingWebCustomizer struct {
	tracer opentracing.Tracer
}

func NewTracingWebCustomizer(tracer opentracing.Tracer) *TracingWebCustomizer{
	return &TracingWebCustomizer{
		tracer: tracer,
	}
}

// we want TracingWebCustomizer before anything else
func (c TracingWebCustomizer) Order() int {
	return order.Highest
}

func (c *TracingWebCustomizer) Customize(ctx context.Context, r *web.Registrar) error {
	// for gin
	r.AddGlobalMiddlewares(GinTracing(c.tracer, tracing.OpNameHttp, excludeRequest))

	// for go-kit endpoints, because we are unable to finish the created span,
	// so we rely on Gin middleware to create/finish span
	//t := kithttp.ServerBefore(kitopentracing.HTTPToContext(c.tracer, tracing.OpNameHttp, logger))
	//r.AddOption(t)
	return nil
}


/*********************
	common funcs
 *********************/
func opNameWithRequest(opName string, r *http.Request) string {
	return opName + " " + r.URL.Path
}

/*********************
	exlusion matcher
 *********************/
var (
	healthMatcher = exlusionMatcher{
		matches: func(r *http.Request) bool {
			return strings.HasSuffix(r.URL.Path, "/health") && r.Method == http.MethodGet
		},
	}

	corsPreflightMatcher = exlusionMatcher{
		matches: func(r *http.Request) bool {
			return r.Method == http.MethodOptions
		},
	}
)

// exlusionMatcher is specialized web.RequestMatcher that do faster matching (simplier and relaxed logic)
type exlusionMatcher struct {
	matches func(*http.Request) bool
}

func (m exlusionMatcher) Matches(i interface{}) (bool, error) {
	r, ok := i.(*http.Request)
	return ok && m.matches(r) , nil
}

func (m exlusionMatcher) MatchesWithContext(ctx context.Context, i interface{}) (bool, error) {
	return m.Matches(i)
}

func (m exlusionMatcher) Or(matcher ...util_matcher.Matcher) util_matcher.ChainableMatcher {
	return util_matcher.Or(m, matcher...)
}

func (m exlusionMatcher) And(matcher ...util_matcher.Matcher) util_matcher.ChainableMatcher {
	return util_matcher.And(m, matcher...)
}


