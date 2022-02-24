package monitor

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/profiler"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	HttpChartPrefix = "charts"
)

var (
	DataGroups = []DataGroup{GroupBytesAllocated, GroupGCPauses, GroupCPUUsage, GroupPprof}
	errWSWriterNotAvailable = errors.New("WebSocket writer not available")
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
		storage:   storage,
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
	// Subscribe data collector
	ch, id, e := c.collector.Subscribe()
	defer c.collector.Unsubscribe(id)
	if e != nil {
		c.handleError(gc, e)
		return
	}

	// Upgrade to websocket connection
	ws, e := c.upgrader.Upgrade(gc.Writer, gc.Request, nil)
	if e != nil {
		c.handleError(gc, e)
		return
	}
	defer func() {
		_ = ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second))
		_ = ws.Close()
	}()

	// read and discard all messages
	go c.wsReadSink(gc.Request.Context())(ws)

	// write feed when received from collector, and send PingMessage every 10 feeds
	var lastPing, lastPong time.Time
	ws.SetPongHandler(func(s string) error {
		lastPong = time.Now()
		return nil
	})

LOOP:
	for i := uint(0); true; i++ {
		select {
		case feed := <-ch:
			switch e := c.wsWriteJson(ws, &feed); {
			case errors.Is(errWSWriterNotAvailable, e):
				break LOOP
			}
		case <-gc.Request.Context().Done():
			break LOOP
		}
		if i % 10 != 0 {
			continue
		}

		// If we didn't receive Pong after 1 mins, we quit loop and close connection
		if lastPing.Sub(lastPong) > time.Minute {
			break LOOP
		}

		// Ping
		lastPing = time.Now()
		if e := ws.WriteControl(websocket.PingMessage, nil, lastPing.Add(time.Second)); e != nil {
			break LOOP
		}
	}
}

func (c *ChartController) handleError(gc *gin.Context, e error) {
	gc.AbortWithStatusJSON(http.StatusInternalServerError, map[string]interface{}{
		"error": e.Error(),
	})
}

func (c *ChartController) wsWriteJson(ws *websocket.Conn, v interface{}) error {
	switch w, e := ws.NextWriter(websocket.TextMessage); {
	case e != nil:
		return errWSWriterNotAvailable
	default:
		return json.NewEncoder(w).Encode(v)
	}
}

func (c *ChartController) wsReadSink(ctx context.Context) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
	LOOP:
		for e := error(nil); e == nil; _, _, e = ws.NextReader() {
			select {
			case <-ctx.Done():
				break LOOP
			default:
			}
		}
	}
}