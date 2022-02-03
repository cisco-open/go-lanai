package monitor

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/profiler"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"path/filepath"
	"time"
)

const (
	HttpChartPrefix = "charts"
)

var (
	DataGroups = []DataGroup{GroupBytesAllocated, GroupGCPauses, GroupCPUUsage, GroupPprof}
)

type ChartsForwardRequest struct {
	Path string `uri:"path"`
}

type ChartController struct {
	storage   DataStorage
	collector *dataCollector
	upgrader  *websocket.Upgrader
}

func NewChartController(storage DataStorage, collector *dataCollector) *ChartController {
	return &ChartController{
		storage: storage,
		collector: collector,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (c *ChartController) Mappings() []web.Mapping {
	return []web.Mapping{
		assets.New(fmt.Sprintf("%s/%s/%s", profiler.RouteGroup, HttpChartPrefix, "static"), "static/"),
		web.NewSimpleMapping("chart_ui", profiler.RouteGroup, HttpChartPrefix+"/", http.MethodGet, nil, c.ChartUI),
		web.NewSimpleGinMapping("chart_data", profiler.RouteGroup, HttpChartPrefix+"/data", http.MethodGet, nil, c.Data),
		web.NewSimpleGinMapping("chart_feed", profiler.RouteGroup, HttpChartPrefix+"/data-feed", http.MethodGet, nil, c.DataFeed),
	}
}

func (c *ChartController) ChartUI(w http.ResponseWriter, r *http.Request) {
	fs := http.FS(Content)
	file, err := fs.Open("static/index.html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func (c *ChartController) Data(gc *gin.Context) {
	callback := gc.Query("callback")

	gc.Header("Content-Type", "application/json")
	data, e := c.storage.Read(gc.Request.Context(), DataGroups...)
	if e != nil {
		c.handleError(gc, e)
		return
	}
	if _, e := fmt.Fprintf(gc.Writer, "%v(", callback); e != nil {
		c.handleError(gc, e)
		return
	}

	encoder := json.NewEncoder(gc.Writer)
	if e := encoder.Encode(data); e != nil {
		c.handleError(gc, e)
		return
	}

	if _, e := fmt.Fprint(gc.Writer, ")"); e != nil {
		c.handleError(gc, e)
		return
	}
}

func (c *ChartController) DataFeed(gc *gin.Context) {
	var (
		lastPing time.Time
		lastPong time.Time
	)

	ch, id, e := c.collector.Subscribe()
	defer c.collector.Unsubscribe(id)
	if e != nil {
		c.handleError(gc, e)
		return
	}

	ws, e := c.upgrader.Upgrade(gc.Writer, gc.Request, nil)
	if e != nil {
		c.handleError(gc, e)
		return
	}

	ws.SetPongHandler(func(s string) error {
		lastPong = time.Now()
		return nil
	})

	// read and discard all messages
	go func(c *websocket.Conn) {
		for {
			if _, _, err := c.NextReader(); err != nil {
				_ = c.Close()
				break
			}
		}
	}(ws)

	defer func() {
		c.collector.Unsubscribe(id)
		_ = ws.Close()
	}()

	var i uint

	for feed := range ch {
		_ = ws.WriteJSON(&feed)
		i++

		if i%10 == 0 {
			if diff := lastPing.Sub(lastPong); diff > time.Second*60 {
				return
			}
			now := time.Now()
			if err := ws.WriteControl(websocket.PingMessage, nil, now.Add(time.Second)); err != nil {
				return
			}
			lastPing = now
		}
	}
	logger.Infof("Loop Quit")
}

func (c *ChartController) handleError(gc *gin.Context, e error) {
	gc.AbortWithStatusJSON(http.StatusInternalServerError, map[string]interface{}{
		"error": e.Error(),
	})
}

func (c *ChartController) Chart(gc *gin.Context) {
	// re-write path by striping leading context-path
	path := gc.Request.URL.Path
	if ctxPath, ok := gc.Request.Context().Value(web.ContextKeyContextPath).(string); ok {
		var e error
		if path, e = filepath.Rel(ctxPath, path); e != nil {
			_ = gc.AbortWithError(500, e)
			return
		}
	}

	gc.Request.URL.Path = "/" + path
	h, pattern := http.DefaultServeMux.Handler(gc.Request)
	fmt.Printf("Pattern: %s", pattern)
	h.ServeHTTP(gc.Writer, gc.Request)
}

func (c *ChartController) Forward(gc *gin.Context) {
	// re-write path by striping leading context-path
	path := gc.Request.URL.Path
	if ctxPath, ok := gc.Request.Context().Value(web.ContextKeyContextPath).(string); ok {
		var e error
		if path, e = filepath.Rel(ctxPath, path); e != nil {
			_ = gc.AbortWithError(500, e)
			return
		}
	}

	gc.Request.URL.Path = "/" + path
	h, pattern := http.DefaultServeMux.Handler(gc.Request)
	fmt.Printf("Pattern: %s", pattern)
	h.ServeHTTP(gc.Writer, gc.Request)
}

//func init() {
//	http.HandleFunc("/debug/charts/data-feed", s.dataFeedHandler)
//	http.HandleFunc("/debug/charts/data", dataHandler)
//	http.HandleFunc("/debug/charts/", handleAsset("static/index.html"))
//	http.HandleFunc("/debug/charts/main.js", handleAsset("static/main.js"))
//	http.HandleFunc("/debug/charts/jquery-2.1.4.min.js", handleAsset("static/jquery-2.1.4.min.js"))
//	http.HandleFunc("/debug/charts/plotly-1.51.3.min.js", handleAsset("static/plotly-1.51.3.min.js"))
//	http.HandleFunc("/debug/charts/moment.min.js", handleAsset("static/moment.min.js"))
//
//	http.DefaultServeMux
//
//	myProcess, _ = process.NewProcess(int32(os.Getpid()))
//
//	// preallocate arrays in data, helps save on reallocations caused by append()
//	// when maxCount is large
//	data.BytesAllocated = make([]SimplePair, 0, maxCount)
//	data.GcPauses = make([]SimplePair, 0, maxCount)
//	data.CPUUsage = make([]CPUPair, 0, maxCount)
//	data.Pprof = make([]PprofPair, 0, maxCount)
//
//	go s.gatherData()
//}
