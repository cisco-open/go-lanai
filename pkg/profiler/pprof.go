package profiler

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"net/http/pprof"
)

type PProfRequest struct {
	Profile string `uri:"profile"`
}

type PProfController struct{}

func (c *PProfController) Mappings() []web.Mapping {
	return []web.Mapping{
		web.NewSimpleGinMapping("pprof_gin", RouteGroup, PathPrefixPProf + "/:profile", web.MethodAny, nil, c.Profile),
		web.NewSimpleMapping("pprof_index", RouteGroup, PathPrefixPProf , web.MethodAny, nil, pprof.Index),
		web.NewSimpleMapping("pprof_cli", RouteGroup, PathPrefixPProf + "/cmdline", web.MethodAny, nil, pprof.Cmdline),
		web.NewSimpleMapping("pprof_profile", RouteGroup, PathPrefixPProf + "/profile", web.MethodAny, nil, pprof.Profile),
		web.NewSimpleMapping("pprof_symbol", RouteGroup, PathPrefixPProf + "/symbol", web.MethodAny, nil, pprof.Symbol),
		web.NewSimpleMapping("pprof_trace", RouteGroup, PathPrefixPProf + "/trace", web.MethodAny, nil, pprof.Trace),
	}

}

func (c *PProfController) Profile(gc *gin.Context) {
	var req PProfRequest
	if e := gc.BindUri(&req); e != nil {
		pprof.Index(gc.Writer, gc.Request)
		return
	}

	handler := pprof.Handler(req.Profile)
	if handler == nil {
		pprof.Index(gc.Writer, gc.Request)
		return
	}

	handler.ServeHTTP(gc.Writer, gc.Request)
}
