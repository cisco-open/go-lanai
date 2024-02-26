// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/profiler"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/assets"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
    "time"
)

const (
	HttpChartPrefix = "charts"
)

var (
	// DataGroups collected profiles to return in "data" endpoint
	DataGroups = []DataGroup{GroupBytesAllocated, GroupGCPauses, GroupCPUUsage, GroupPprof}
	// PongTimeout is the time data feed would wait for pong message before it close the websocket connection
	PongTimeout             = time.Minute
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
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed by server"), time.Now().Add(time.Second))
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
		if i%10 != 0 {
			continue
		}
		i = 0

		// If we didn't receive Pong after PongTimeout, we quit loop and close connection
		if lastPing.Sub(lastPong) > PongTimeout {
			logger.WithContext(gc.Request.Context()).Debugf("No 'pong' message received after %v, closing connection...", PongTimeout)
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
