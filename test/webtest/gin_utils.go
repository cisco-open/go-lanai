package webtest

import (
    "context"
    "github.com/cisco-open/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
    "io"
    "net/http"
	"net/http/httptest"
	"sync"
)

const ginContextRecorder = `_go-lanai/webtest/gin/recorder`

type ginContextCreator struct {
	sync.Once
	engine *gin.Engine
}

func (c *ginContextCreator) LazyInit() {
    c.Do(func() {
        gin.SetMode(gin.ReleaseMode)
        c.engine = gin.New()
        c.engine.ContextWithFallback = true
    })
}

func (c *ginContextCreator) CreateWithRequest(req *http.Request) *gin.Context {
    c.LazyInit()
    rw := httptest.NewRecorder()
    gc := gin.CreateTestContextOnly(rw, c.engine)
    gc.Set(ginContextRecorder, rw)
    if req != nil {
        gc.Request = req
        web.GinContextMerger()(gc)
    }
    return gc
}

var defaultGinContextCreator = &ginContextCreator{}

func NewGinContext(ctx context.Context, method, path string, body io.Reader, opts ...RequestOptions) *gin.Context {
    req := httptest.NewRequest(method, path, body).WithContext(ctx)
    for _, fn := range opts {
        if fn != nil {
            fn(req)
        }
    }
    return NewGinContextWithRequest(req)
}

func NewGinContextWithRequest(req *http.Request) *gin.Context {
    return defaultGinContextCreator.CreateWithRequest(req)
}

func GinContextRecorder(gc *gin.Context) *httptest.ResponseRecorder {
    recorder, _ := gc.Value(ginContextRecorder).(*httptest.ResponseRecorder)
    return recorder
}